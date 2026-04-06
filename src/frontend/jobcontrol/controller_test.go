package jobcontrol

import (
	"errors"
	"testing"
	"time"
)

type stubAuthVerifier struct {
	operatorIdentity OperatorIdentity
	operatorErr      error
	workerIdentity   WorkerIdentity
	workerErr        error
}

func (s stubAuthVerifier) VerifyOperator(string) (OperatorIdentity, error) {
	if s.operatorErr != nil {
		return OperatorIdentity{}, s.operatorErr
	}
	return s.operatorIdentity, nil
}

func (s stubAuthVerifier) VerifyWorker(string) (WorkerIdentity, error) {
	if s.workerErr != nil {
		return WorkerIdentity{}, s.workerErr
	}
	return s.workerIdentity, nil
}

type stubService struct {
	submitResp    SubmitJobResponse
	submitErr     error
	getResp       JobStoreRecord
	getErr        error
	cancelResp    CancelJobResponse
	cancelErr     error
	claimResp     ClaimNextResponse
	claimErr      error
	statusResp    JobStoreRecord
	statusErr     error
	resultResp    ResultAccessRecord
	resultErr     error
	lastSubmitReq SubmitJobRequest
	lastGetJobID  string
	lastCancelID  string
	lastWorker    WorkerIdentity
	lastUpdate    StatusUpdate
}

func (s *stubService) Submit(req SubmitJobRequest, _ time.Time) (SubmitJobResponse, error) {
	s.lastSubmitReq = req
	if s.submitErr != nil {
		return SubmitJobResponse{}, s.submitErr
	}
	return s.submitResp, nil
}

func (s *stubService) Get(jobID string) (JobStoreRecord, error) {
	s.lastGetJobID = jobID
	if s.getErr != nil {
		return JobStoreRecord{}, s.getErr
	}
	return s.getResp, nil
}

func (s *stubService) ClaimNext(WorkerIdentity, time.Time) (ClaimNextResponse, error) {
	if s.claimErr != nil {
		return ClaimNextResponse{}, s.claimErr
	}
	return s.claimResp, nil
}

func (s *stubService) ApplyStatusUpdate(_ string, worker WorkerIdentity, update StatusUpdate, _ time.Time) (JobStoreRecord, error) {
	s.lastWorker = worker
	s.lastUpdate = update
	if s.statusErr != nil {
		return JobStoreRecord{}, s.statusErr
	}
	return s.statusResp, nil
}

func (s *stubService) RequestCancel(jobID string, _ time.Time) (CancelJobResponse, error) {
	s.lastCancelID = jobID
	if s.cancelErr != nil {
		return CancelJobResponse{}, s.cancelErr
	}
	return s.cancelResp, nil
}

func (s *stubService) ResolveResultAccess(string, time.Time) (ResultAccessRecord, error) {
	if s.resultErr != nil {
		return ResultAccessRecord{}, s.resultErr
	}
	return s.resultResp, nil
}

func TestControllerSubmitUsesVerifiedOperatorSubject(t *testing.T) {
	service := &stubService{submitResp: SubmitJobResponse{JobID: "job_0001", State: StateQueued, JobRef: "/v1/jobs/job_0001"}}
	controller := NewController(stubAuthVerifier{operatorIdentity: OperatorIdentity{Subject: "operator:alice", AuthMode: AuthModeBearer}}, service)

	_, err := controller.Submit("token", SubmitJobRequest{
		SubmittedBy: "spoofed-subject",
		JobType:     "local-analysis",
		Payload:     map[string]any{"input": "demo"},
	}, time.Date(2026, 4, 5, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if got := service.lastSubmitReq.SubmittedBy; got != "operator:alice" {
		t.Fatalf("submitted_by = %q, want %q", got, "operator:alice")
	}
}

func TestControllerGetRequiresOperatorAuth(t *testing.T) {
	controller := NewController(stubAuthVerifier{operatorErr: errors.New("bad token")}, &stubService{})

	_, err := controller.Get("bad", "job_0001")
	if err == nil {
		t.Fatal("Get() error = nil, want auth failure")
	}
}

func TestControllerCancelPassesThrough(t *testing.T) {
	service := &stubService{cancelResp: CancelJobResponse{JobID: "job_0001", CancelRequested: true, State: StateRunning}}
	controller := NewController(stubAuthVerifier{operatorIdentity: OperatorIdentity{Subject: "operator:bob", AuthMode: AuthModeBearer}}, service)

	resp, err := controller.Cancel("token", "job_0001", time.Date(2026, 4, 5, 11, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if resp.JobID != "job_0001" || !resp.CancelRequested {
		t.Fatalf("Cancel() = %#v, want requested cancel for job_0001", resp)
	}
	if service.lastCancelID != "job_0001" {
		t.Fatalf("cancelled job id = %q, want %q", service.lastCancelID, "job_0001")
	}
}

func TestControllerClaimNextUsesVerifiedWorker(t *testing.T) {
	service := &stubService{claimResp: ClaimNextResponse{WorkerID: "worker-1", PollAfterSeconds: 5}}
	controller := NewController(stubAuthVerifier{workerIdentity: WorkerIdentity{WorkerID: "worker-1", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-1"}}, service)

	resp, err := controller.ClaimNext("worker-token", time.Date(2026, 4, 5, 11, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ClaimNext() error = %v", err)
	}
	if resp.WorkerID != "worker-1" {
		t.Fatalf("worker id = %q, want %q", resp.WorkerID, "worker-1")
	}
}

func TestControllerUpdateStatusOverridesWorkerAndJobFields(t *testing.T) {
	service := &stubService{statusResp: JobStoreRecord{Job: JobRecord{JobID: "job_0001"}}}
	controller := NewController(stubAuthVerifier{workerIdentity: WorkerIdentity{WorkerID: "worker-7", AuthMode: AuthModeBearer, AuthSubject: "worker:worker-7"}}, service)

	_, err := controller.UpdateStatus("worker-token", "job_0001", StatusUpdate{
		JobID:           "wrong-id",
		WorkerID:        "wrong-worker",
		State:           StateSucceeded,
		ProgressPercent: 100,
	}, time.Date(2026, 4, 5, 11, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if service.lastUpdate.JobID != "job_0001" {
		t.Fatalf("update job id = %q, want %q", service.lastUpdate.JobID, "job_0001")
	}
	if service.lastUpdate.WorkerID != "worker-7" {
		t.Fatalf("update worker id = %q, want %q", service.lastUpdate.WorkerID, "worker-7")
	}
	if service.lastWorker.AuthSubject != "worker:worker-7" {
		t.Fatalf("worker auth subject = %q, want %q", service.lastWorker.AuthSubject, "worker:worker-7")
	}
}

func TestControllerResultAccessRequiresOperatorAuth(t *testing.T) {
	controller := NewController(stubAuthVerifier{operatorErr: errors.New("bad token")}, &stubService{})

	_, err := controller.ResultAccess("bad", "job_0001", time.Date(2026, 4, 5, 11, 20, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ResultAccess() error = nil, want auth failure")
	}
}
