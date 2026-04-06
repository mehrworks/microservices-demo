package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("job control API error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}

type Client struct {
	BaseURL     string
	WorkerToken string
	HTTPClient  *http.Client
}

func (c *Client) ClaimNext(ctx context.Context, worker frontendjobcontrol.WorkerIdentity) (frontendjobcontrol.ClaimNextResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/internal/worker/claim", worker)
	if err != nil {
		return frontendjobcontrol.ClaimNextResponse{}, err
	}

	var resp frontendjobcontrol.ClaimNextResponse
	if err := c.do(req, http.StatusOK, &resp); err != nil {
		return frontendjobcontrol.ClaimNextResponse{}, err
	}
	return resp, nil
}

func (c *Client) UpdateStatus(ctx context.Context, jobID string, update frontendjobcontrol.StatusUpdate) (frontendjobcontrol.JobStoreRecord, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/internal/jobs/"+jobID+"/status", update)
	if err != nil {
		return frontendjobcontrol.JobStoreRecord{}, err
	}

	var resp frontendjobcontrol.JobStoreRecord
	if err := c.do(req, http.StatusOK, &resp); err != nil {
		return frontendjobcontrol.JobStoreRecord{}, err
	}
	return resp, nil
}

func (c *Client) RenewLease(ctx context.Context, jobID string, worker frontendjobcontrol.WorkerIdentity) (frontendjobcontrol.LeaseRenewResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/internal/jobs/"+jobID+"/lease:renew", worker)
	if err != nil {
		return frontendjobcontrol.LeaseRenewResponse{}, err
	}

	var resp frontendjobcontrol.LeaseRenewResponse
	if err := c.do(req, http.StatusOK, &resp); err != nil {
		return frontendjobcontrol.LeaseRenewResponse{}, err
	}
	return resp, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	base.Path = strings.TrimRight(base.Path, "/") + path

	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, base.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.WorkerToken)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, expectedStatus int, out any) error {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type Worker struct {
	Client       *Client
	Identity     frontendjobcontrol.WorkerIdentity
	ResultDir    string
	Now          func() time.Time
	Sleep        func(time.Duration)
	ProcessDelay time.Duration
}

type LoopConfig struct {
	IdlePollInterval time.Duration
	MaxJobs          int
}

type CycleResult struct {
	JobID     string
	Processed bool
	PollAfter time.Duration
}

func (w *Worker) RunOnce(ctx context.Context) (string, bool, error) {
	result, err := w.RunCycle(ctx)
	return result.JobID, result.Processed, err
}

func (w *Worker) RunCycle(ctx context.Context) (CycleResult, error) {
	if w == nil || w.Client == nil {
		return CycleResult{}, fmt.Errorf("worker client not configured")
	}
	nowFn := w.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	claim, err := w.Client.ClaimNext(ctx, w.Identity)
	if err != nil {
		return CycleResult{}, err
	}
	pollAfter := time.Duration(claim.PollAfterSeconds) * time.Second
	if pollAfter <= 0 {
		pollAfter = 5 * time.Second
	}
	if claim.Job == nil {
		return CycleResult{Processed: false, PollAfter: pollAfter}, nil
	}

	now := nowFn()
	if err := w.simulateProcessing(ctx, claim.Job, nowFn); err != nil {
		return CycleResult{JobID: claim.Job.JobID, Processed: false, PollAfter: pollAfter}, err
	}
	now = nowFn()
	resultRef, err := w.writeSyntheticResult(claim.Job, now)
	if err != nil {
		return CycleResult{JobID: claim.Job.JobID, Processed: false, PollAfter: pollAfter}, err
	}

	_, err = w.Client.UpdateStatus(ctx, claim.Job.JobID, frontendjobcontrol.StatusUpdate{
		State:           frontendjobcontrol.StateSucceeded,
		UpdatedAt:       now,
		ProgressPercent: 100,
		StatusMessage:   "completed by local worker",
		ResultRef:       resultRef,
	})
	if err != nil {
		return CycleResult{JobID: claim.Job.JobID, Processed: false, PollAfter: pollAfter}, err
	}

	return CycleResult{JobID: claim.Job.JobID, Processed: true, PollAfter: pollAfter}, nil
}

func (w *Worker) RunLoop(ctx context.Context, cfg LoopConfig) error {
	sleepFn := w.Sleep
	contextAwareSleep := false
	if sleepFn == nil {
		sleepFn = time.Sleep
		contextAwareSleep = true
	}
	idlePoll := cfg.IdlePollInterval
	if idlePoll <= 0 {
		idlePoll = 5 * time.Second
	}

	processedCount := 0
	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		result, err := w.RunCycle(ctx)
		if err != nil {
			return err
		}

		if result.Processed {
			processedCount++
			if cfg.MaxJobs > 0 && processedCount >= cfg.MaxJobs {
				return nil
			}
			continue
		}

		delay := result.PollAfter
		if delay <= 0 {
			delay = idlePoll
		}
		if err := sleepWithContext(ctx, sleepFn, delay, contextAwareSleep); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}
}

func (w *Worker) simulateProcessing(ctx context.Context, job *frontendjobcontrol.ClaimedJob, nowFn func() time.Time) error {
	if w.ProcessDelay <= 0 {
		return nil
	}
	sleepFn := w.Sleep
	if sleepFn == nil {
		sleepFn = time.Sleep
	}
	remaining := w.ProcessDelay
	lease := job.Lease
	renewAfter := time.Duration(lease.RenewAfterSeconds) * time.Second
	if renewAfter <= 0 {
		renewAfter = remaining
	}

	for remaining > 0 {
		step := remaining
		if remaining > renewAfter {
			step = renewAfter
		}
		sleepFn(step)
		remaining -= step
		if remaining <= 0 {
			break
		}
		resp, err := w.Client.RenewLease(ctx, job.JobID, w.Identity)
		if err != nil {
			return err
		}
		lease = resp.Lease
		job.Lease = resp.Lease
		renewAfter = time.Duration(lease.RenewAfterSeconds) * time.Second
		if renewAfter <= 0 {
			renewAfter = remaining
		}
		_ = nowFn
	}

	return nil
}

func (w *Worker) writeSyntheticResult(job *frontendjobcontrol.ClaimedJob, now time.Time) (*frontendjobcontrol.ResultReference, error) {
	resultDir := w.ResultDir
	if resultDir == "" {
		resultDir = "/tmp/jobcontrol-results"
	}
	if err := os.MkdirAll(resultDir, 0o755); err != nil {
		return nil, err
	}

	resultPath := filepath.Join(resultDir, job.JobID+".json")
	payload := map[string]any{
		"job_id":       job.JobID,
		"job_type":     job.JobType,
		"attempt":      job.Attempt,
		"processed_at": now.Format(time.RFC3339Nano),
		"payload":      job.Payload,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(resultPath, data, 0o644); err != nil {
		return nil, err
	}

	return &frontendjobcontrol.ResultReference{
		Kind:        "local-file",
		URI:         "file://" + resultPath,
		ContentType: "application/json",
		SizeBytes:   int64(len(data)),
	}, nil
}

func sleepWithContext(ctx context.Context, sleepFn func(time.Duration), d time.Duration, contextAware bool) error {
	if d <= 0 {
		return nil
	}
	if !contextAware {
		sleepFn(d)
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
