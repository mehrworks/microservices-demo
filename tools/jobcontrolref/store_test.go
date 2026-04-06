package jobcontrolref

import (
	"errors"
	"testing"
	"time"
)

func TestClaimConflictAndLeaseExpiry(t *testing.T) {
	store := NewStore()
	now := time.Now()

	resp, err := store.Submit(SubmitJobRequest{
		JobType:                 "local-analysis",
		SubmittedBy:             "frontend-session-1",
		RequestedTimeoutSeconds: 1800,
		Payload:                 map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker1 := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 60}
	worker2 := WorkerIdentity{WorkerID: "worker-2", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-2", MaxLeaseSeconds: 60}

	claim1, err := store.ClaimNext(worker1, now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("ClaimNext(worker1) error = %v", err)
	}
	if claim1.Job == nil || claim1.Job.JobID != resp.JobID {
		t.Fatalf("ClaimNext(worker1) = %#v, want claimed job %q", claim1, resp.JobID)
	}
	if claim1.Job.Attempt != 1 {
		t.Fatalf("first attempt = %d, want 1", claim1.Job.Attempt)
	}

	claim2, err := store.ClaimNext(worker2, now.Add(10*time.Second))
	if err != nil {
		t.Fatalf("ClaimNext(worker2 immediate) error = %v", err)
	}
	if claim2.Job != nil {
		t.Fatalf("ClaimNext(worker2 immediate) = %#v, want no job", claim2)
	}

	claim3, err := store.ClaimNext(worker2, now.Add(70*time.Second))
	if err != nil {
		t.Fatalf("ClaimNext(worker2 after expiry) error = %v", err)
	}
	if claim3.Job == nil {
		t.Fatal("ClaimNext(worker2 after expiry) = nil job, want reclaimed job")
	}
	if claim3.Job.JobID != resp.JobID {
		t.Fatalf("reclaimed job id = %q, want %q", claim3.Job.JobID, resp.JobID)
	}
	if claim3.Job.Attempt != 2 {
		t.Fatalf("reclaimed attempt = %d, want 2", claim3.Job.Attempt)
	}

	record, err := store.Get(resp.JobID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if record.Job.ClaimedBy == nil || *record.Job.ClaimedBy != worker2.WorkerID {
		t.Fatalf("claimed_by = %#v, want %q", record.Job.ClaimedBy, worker2.WorkerID)
	}
	if record.Job.ClaimedByAuthSubject == nil || *record.Job.ClaimedByAuthSubject != worker2.AuthSubject {
		t.Fatalf("claimed_by_auth_subject = %#v, want %q", record.Job.ClaimedByAuthSubject, worker2.AuthSubject)
	}
}

func TestApplyStatusAndResolveResultAccess(t *testing.T) {
	store := NewStore()
	now := time.Now()

	resp, err := store.Submit(SubmitJobRequest{
		IdempotencyKey:            "job-key-1",
		JobType:                   "local-analysis",
		SubmittedBy:               "frontend-session-2",
		RequestedTimeoutSeconds:   1200,
		RequestedResultTTLSeconds: 120,
		Payload:                   map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 120}
	claim, err := store.ClaimNext(worker, now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}
	if claim.Job == nil {
		t.Fatal("ClaimNext() returned nil job")
	}

	ref := &ResultReference{
		Kind:        "local-file",
		URI:         "file:///tmp/output.json",
		ContentType: "application/json",
		SizeBytes:   2048,
	}

	_, err = store.ApplyStatusUpdate(resp.JobID, worker, StatusUpdate{
		JobID:           resp.JobID,
		WorkerID:        worker.WorkerID,
		State:           StateSucceeded,
		UpdatedAt:       now.Add(30 * time.Second),
		ProgressPercent: 100,
		StatusMessage:   "done",
		ResultRef:       ref,
	}, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("ApplyStatusUpdate() error = %v", err)
	}

	access, err := store.ResolveResultAccess(resp.JobID, now.Add(31*time.Second))
	if err != nil {
		t.Fatalf("ResolveResultAccess() error = %v", err)
	}
	if access.RetrievalMode != RetrievalModeReferenceOnly {
		t.Fatalf("retrieval mode = %q, want %q", access.RetrievalMode, RetrievalModeReferenceOnly)
	}
	if !access.Available {
		t.Fatal("result access not available, want available")
	}
	if access.ResultRef.URI != ref.URI {
		t.Fatalf("result uri = %q, want %q", access.ResultRef.URI, ref.URI)
	}

	_, err = store.ResolveResultAccess(resp.JobID, now.Add(3*time.Minute))
	if !errors.Is(err, ErrResultUnavailable) {
		t.Fatalf("ResolveResultAccess(after expiry) error = %v, want %v", err, ErrResultUnavailable)
	}
}

func TestCancelQueuedJob(t *testing.T) {
	store := NewStore()
	now := time.Now()

	resp, err := store.Submit(SubmitJobRequest{
		JobType:     "local-analysis",
		SubmittedBy: "frontend-session-3",
		Payload:     map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	cancelResp, err := store.RequestCancel(resp.JobID, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("RequestCancel() error = %v", err)
	}
	if !cancelResp.CancelRequested {
		t.Fatal("cancel requested = false, want true")
	}
	if cancelResp.State != StateCancelled {
		t.Fatalf("cancel state = %q, want %q", cancelResp.State, StateCancelled)
	}

	record, err := store.Get(resp.JobID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if record.Job.State != StateCancelled {
		t.Fatalf("stored state = %q, want %q", record.Job.State, StateCancelled)
	}
	if !record.Job.CancelRequested {
		t.Fatal("stored cancel_requested = false, want true")
	}
}
