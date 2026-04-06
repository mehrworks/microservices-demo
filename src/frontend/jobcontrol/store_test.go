package jobcontrol

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestClaimConflictAndLeaseExpiry(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 5, 8, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now.Add(71 * time.Second) }

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
	now := time.Date(2026, 4, 5, 9, 0, 0, 0, time.UTC)

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
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now.Add(2 * time.Second) }

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

func TestPersistentStoreReloadsSubmittedJob(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "jobcontrol-state.json")
	store, err := NewPersistentStore(statePath)
	if err != nil {
		t.Fatalf("NewPersistentStore() error = %v", err)
	}
	now := time.Date(2026, 4, 6, 6, 0, 0, 0, time.UTC)

	resp, err := store.Submit(SubmitJobRequest{
		JobType:     "local-analysis",
		SubmittedBy: "operator:alice",
		Payload:     map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	reloaded, err := NewPersistentStore(statePath)
	if err != nil {
		t.Fatalf("NewPersistentStore(reload) error = %v", err)
	}

	record, err := reloaded.Get(resp.JobID)
	if err != nil {
		t.Fatalf("Get() after reload error = %v", err)
	}
	if record.Job.JobID != resp.JobID {
		t.Fatalf("reloaded job id = %q, want %q", record.Job.JobID, resp.JobID)
	}
	if record.Job.State != StateQueued {
		t.Fatalf("reloaded state = %q, want %q", record.Job.State, StateQueued)
	}
}

func TestPersistentStoreReloadsResultAccess(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "jobcontrol-state.json")
	store, err := NewPersistentStore(statePath)
	if err != nil {
		t.Fatalf("NewPersistentStore() error = %v", err)
	}
	now := time.Date(2026, 4, 6, 6, 10, 0, 0, time.UTC)

	resp, err := store.Submit(SubmitJobRequest{
		JobType:                   "local-analysis",
		SubmittedBy:               "operator:alice",
		RequestedResultTTLSeconds: 300,
		Payload:                   map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 60}
	if _, err := store.ClaimNext(worker, now.Add(5*time.Second)); err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}

	ref := &ResultReference{Kind: "local-file", URI: "file:///tmp/output.json", ContentType: "application/json", SizeBytes: 99}
	if _, err := store.ApplyStatusUpdate(resp.JobID, worker, StatusUpdate{
		State:           StateSucceeded,
		ProgressPercent: 100,
		StatusMessage:   "done",
		ResultRef:       ref,
	}, now.Add(30*time.Second)); err != nil {
		t.Fatalf("ApplyStatusUpdate() error = %v", err)
	}

	reloaded, err := NewPersistentStore(statePath)
	if err != nil {
		t.Fatalf("NewPersistentStore(reload) error = %v", err)
	}

	access, err := reloaded.ResolveResultAccess(resp.JobID, now.Add(31*time.Second))
	if err != nil {
		t.Fatalf("ResolveResultAccess() after reload error = %v", err)
	}
	if access.ResultRef.URI != ref.URI {
		t.Fatalf("reloaded result uri = %q, want %q", access.ResultRef.URI, ref.URI)
	}
	if !access.Available {
		t.Fatal("reloaded result access unavailable, want available")
	}
}

func TestRenewLeaseExtendsLease(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC)

	resp, err := store.Submit(SubmitJobRequest{
		JobType:     "local-analysis",
		SubmittedBy: "operator:alice",
		Payload:     map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 60}
	claim, err := store.ClaimNext(worker, now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}
	if claim.Job == nil {
		t.Fatal("ClaimNext() returned nil job")
	}

	renewed, err := store.RenewLease(resp.JobID, worker, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("RenewLease() error = %v", err)
	}
	if !renewed.Lease.LeaseExpiresAt.After(claim.Job.Lease.LeaseExpiresAt) {
		t.Fatalf("renewed expiry = %v, want after %v", renewed.Lease.LeaseExpiresAt, claim.Job.Lease.LeaseExpiresAt)
	}
}

func TestRenewLeaseFailsAfterExpiryAndRequeuesJob(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 6, 8, 30, 0, 0, time.UTC)

	resp, err := store.Submit(SubmitJobRequest{
		JobType:     "local-analysis",
		SubmittedBy: "operator:alice",
		Payload:     map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 30}
	if _, err := store.ClaimNext(worker, now.Add(1*time.Second)); err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}

	_, err = store.RenewLease(resp.JobID, worker, now.Add(40*time.Second))
	if !errors.Is(err, ErrLeaseConflict) {
		t.Fatalf("RenewLease() error = %v, want %v", err, ErrLeaseConflict)
	}

	record, err := store.Get(resp.JobID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if record.Job.State != StateQueued {
		t.Fatalf("state = %q, want %q", record.Job.State, StateQueued)
	}
	if record.Job.Attempt != 2 {
		t.Fatalf("attempt = %d, want %d", record.Job.Attempt, 2)
	}
}

func TestGetRequeuesExpiredRunningJob(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now.Add(40 * time.Second) }

	resp, err := store.Submit(SubmitJobRequest{
		JobType:     "local-analysis",
		SubmittedBy: "operator:alice",
		Payload:     map[string]any{"input": "demo"},
	}, now)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	worker := WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1", MaxLeaseSeconds: 30}
	if _, err := store.ClaimNext(worker, now.Add(1*time.Second)); err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}

	record, err := store.Get(resp.JobID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if record.Job.State != StateQueued {
		t.Fatalf("state = %q, want %q", record.Job.State, StateQueued)
	}
	if record.Job.Attempt != 2 {
		t.Fatalf("attempt = %d, want %d", record.Job.Attempt, 2)
	}
}
