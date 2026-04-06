package jobcontrol

import "time"

type OperatorIdentity struct {
	Subject  string
	AuthMode AuthMode
}

type OperatorAPI interface {
	Submit(rawToken string, req SubmitJobRequest, now time.Time) (SubmitJobResponse, error)
	Get(rawToken, jobID string) (JobStoreRecord, error)
	ResultAccess(rawToken, jobID string, now time.Time) (ResultAccessRecord, error)
	Cancel(rawToken, jobID string, now time.Time) (CancelJobResponse, error)
}

type WorkerAPI interface {
	ClaimNext(rawToken string, requested WorkerIdentity, now time.Time) (ClaimNextResponse, error)
	RenewLease(rawToken, jobID string, requested WorkerIdentity, now time.Time) (LeaseRenewResponse, error)
	UpdateStatus(rawToken, jobID string, update StatusUpdate, now time.Time) (JobStoreRecord, error)
}

type AuthVerifier interface {
	VerifyOperator(rawToken string) (OperatorIdentity, error)
	VerifyWorker(rawToken string) (WorkerIdentity, error)
}

type Service interface {
	Submit(req SubmitJobRequest, now time.Time) (SubmitJobResponse, error)
	Get(jobID string) (JobStoreRecord, error)
	ClaimNext(worker WorkerIdentity, now time.Time) (ClaimNextResponse, error)
	RenewLease(jobID string, worker WorkerIdentity, now time.Time) (LeaseRenewResponse, error)
	ApplyStatusUpdate(jobID string, worker WorkerIdentity, update StatusUpdate, now time.Time) (JobStoreRecord, error)
	RequestCancel(jobID string, now time.Time) (CancelJobResponse, error)
	ResolveResultAccess(jobID string, now time.Time) (ResultAccessRecord, error)
}
