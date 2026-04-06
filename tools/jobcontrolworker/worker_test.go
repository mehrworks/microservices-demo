package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

func TestRunOnceCompletesClaimedJob(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Date(2026, 4, 6, 7, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/worker/claim":
			if r.Method != http.MethodPost {
				t.Fatalf("claim method = %s, want POST", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer worker-secret" {
				t.Fatalf("authorization = %q, want %q", got, "Bearer worker-secret")
			}
			var worker frontendjobcontrol.WorkerIdentity
			if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
				t.Fatalf("decode worker identity: %v", err)
			}
			if worker.WorkerID != "local-worker-1" {
				t.Fatalf("worker id = %q, want %q", worker.WorkerID, "local-worker-1")
			}
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.ClaimNextResponse{
				WorkerID:         "worker:static-bearer",
				PollAfterSeconds: 5,
				Job: &frontendjobcontrol.ClaimedJob{
					JobID:   "job_0001",
					JobType: "local-analysis",
					Attempt: 1,
					Lease: frontendjobcontrol.Lease{
						LeaseID:              "lease_job_0001_worker:static-bearer",
						IssuedAt:             now,
						LeaseExpiresAt:       now.Add(5 * time.Minute),
						LeaseDurationSeconds: 300,
						RenewAfterSeconds:    240,
					},
					Payload: map[string]any{"input": "demo"},
				},
			})
		case "/internal/jobs/job_0001/status":
			if r.Method != http.MethodPost {
				t.Fatalf("status method = %s, want POST", r.Method)
			}
			var update frontendjobcontrol.StatusUpdate
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				t.Fatalf("decode status update: %v", err)
			}
			if update.State != frontendjobcontrol.StateSucceeded {
				t.Fatalf("state = %q, want %q", update.State, frontendjobcontrol.StateSucceeded)
			}
			if update.ResultRef == nil || !strings.HasPrefix(update.ResultRef.URI, "file://") {
				t.Fatalf("result ref = %#v, want local file uri", update.ResultRef)
			}
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.JobStoreRecord{Job: frontendjobcontrol.JobRecord{JobID: "job_0001"}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	worker := &Worker{
		Client:    &Client{BaseURL: server.URL, WorkerToken: "worker-secret", HTTPClient: server.Client()},
		Identity:  frontendjobcontrol.WorkerIdentity{WorkerID: "local-worker-1", AuthMode: frontendjobcontrol.AuthModeBearer, AuthSubject: "worker:local-worker-1"},
		ResultDir: tempDir,
		Now:       func() time.Time { return now },
	}

	jobID, processed, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if !processed {
		t.Fatal("processed = false, want true")
	}
	if jobID != "job_0001" {
		t.Fatalf("job id = %q, want %q", jobID, "job_0001")
	}

	resultPath := filepath.Join(tempDir, "job_0001.json")
	if _, err := os.Stat(resultPath); err != nil {
		t.Fatalf("result file stat error = %v", err)
	}
}

func TestRunOnceRenewsLeaseDuringLongProcessing(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Date(2026, 4, 6, 7, 30, 0, 0, time.UTC)
	currentTime := now
	renewCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/worker/claim":
			_ = json.NewDecoder(r.Body).Decode(&frontendjobcontrol.WorkerIdentity{})
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.ClaimNextResponse{
				WorkerID:         "local-worker-1",
				PollAfterSeconds: 5,
				Job: &frontendjobcontrol.ClaimedJob{
					JobID:   "job_lease",
					JobType: "local-analysis",
					Attempt: 1,
					Lease: frontendjobcontrol.Lease{
						LeaseID:              "lease_job_lease_local-worker-1",
						IssuedAt:             currentTime,
						LeaseExpiresAt:       currentTime.Add(30 * time.Second),
						LeaseDurationSeconds: 30,
						RenewAfterSeconds:    10,
					},
					Payload: map[string]any{"input": "demo"},
				},
			})
		case "/internal/jobs/job_lease/lease:renew":
			renewCalled = true
			var worker frontendjobcontrol.WorkerIdentity
			if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
				t.Fatalf("decode renew worker identity: %v", err)
			}
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.LeaseRenewResponse{
				JobID: "job_lease",
				Lease: frontendjobcontrol.Lease{
					LeaseID:              "lease_job_lease_local-worker-1_renewed",
					IssuedAt:             currentTime,
					LeaseExpiresAt:       currentTime.Add(30 * time.Second),
					LeaseDurationSeconds: 30,
					RenewAfterSeconds:    10,
				},
			})
		case "/internal/jobs/job_lease/status":
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.JobStoreRecord{Job: frontendjobcontrol.JobRecord{JobID: "job_lease"}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	worker := &Worker{
		Client:    &Client{BaseURL: server.URL, WorkerToken: "worker-secret", HTTPClient: server.Client()},
		Identity:  frontendjobcontrol.WorkerIdentity{WorkerID: "local-worker-1", AuthMode: frontendjobcontrol.AuthModeBearer, AuthSubject: "worker:local-worker-1", MaxLeaseSeconds: 30},
		ResultDir: tempDir,
		Now:       func() time.Time { return currentTime },
		Sleep: func(d time.Duration) {
			currentTime = currentTime.Add(d)
		},
		ProcessDelay: 12 * time.Second,
	}

	jobID, processed, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if !processed || jobID != "job_lease" {
		t.Fatalf("processed=%v jobID=%q, want true and %q", processed, jobID, "job_lease")
	}
	if !renewCalled {
		t.Fatal("expected lease renew call during long processing")
	}
}

func TestRunOnceNoJobAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(frontendjobcontrol.ClaimNextResponse{WorkerID: "worker:static-bearer", PollAfterSeconds: 5, Job: nil})
	}))
	defer server.Close()

	worker := &Worker{
		Client:   &Client{BaseURL: server.URL, WorkerToken: "worker-secret", HTTPClient: server.Client()},
		Identity: frontendjobcontrol.WorkerIdentity{WorkerID: "local-worker-1", AuthMode: frontendjobcontrol.AuthModeBearer, AuthSubject: "worker:local-worker-1"},
	}

	jobID, processed, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if processed {
		t.Fatal("processed = true, want false")
	}
	if jobID != "" {
		t.Fatalf("job id = %q, want empty", jobID)
	}
}

func TestRunLoopProcessesJobAfterIdlePoll(t *testing.T) {
	now := time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC)
	currentTime := now
	claimCalls := 0
	slept := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/worker/claim":
			claimCalls++
			if claimCalls == 1 {
				_ = json.NewEncoder(w).Encode(frontendjobcontrol.ClaimNextResponse{WorkerID: "local-worker-1", PollAfterSeconds: 2, Job: nil})
				return
			}
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.ClaimNextResponse{
				WorkerID:         "local-worker-1",
				PollAfterSeconds: 2,
				Job: &frontendjobcontrol.ClaimedJob{
					JobID:   "job_loop",
					JobType: "local-analysis",
					Attempt: 1,
					Lease: frontendjobcontrol.Lease{
						LeaseID:              "lease_job_loop_local-worker-1",
						IssuedAt:             currentTime,
						LeaseExpiresAt:       currentTime.Add(30 * time.Second),
						LeaseDurationSeconds: 30,
						RenewAfterSeconds:    10,
					},
					Payload: map[string]any{"input": "demo"},
				},
			})
		case "/internal/jobs/job_loop/status":
			_ = json.NewEncoder(w).Encode(frontendjobcontrol.JobStoreRecord{Job: frontendjobcontrol.JobRecord{JobID: "job_loop"}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	worker := &Worker{
		Client:   &Client{BaseURL: server.URL, WorkerToken: "worker-secret", HTTPClient: server.Client()},
		Identity: frontendjobcontrol.WorkerIdentity{WorkerID: "local-worker-1", AuthMode: frontendjobcontrol.AuthModeBearer, AuthSubject: "worker:local-worker-1"},
		Now:      func() time.Time { return currentTime },
		Sleep: func(d time.Duration) {
			slept++
			currentTime = currentTime.Add(d)
		},
	}

	if err := worker.RunLoop(context.Background(), LoopConfig{IdlePollInterval: 2 * time.Second, MaxJobs: 1}); err != nil {
		t.Fatalf("RunLoop() error = %v", err)
	}
	if claimCalls < 2 {
		t.Fatalf("claim calls = %d, want at least 2", claimCalls)
	}
	if slept == 0 {
		t.Fatal("expected loop to sleep before retrying claim")
	}
}
