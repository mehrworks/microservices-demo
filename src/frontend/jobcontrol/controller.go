package jobcontrol

import (
	"fmt"
	"time"
)

type Controller struct {
	auth    AuthVerifier
	service Service
}

func NewController(auth AuthVerifier, service Service) *Controller {
	return &Controller{auth: auth, service: service}
}

func (c *Controller) Submit(rawToken string, req SubmitJobRequest, now time.Time) (SubmitJobResponse, error) {
	identity, err := c.verifyOperator(rawToken)
	if err != nil {
		return SubmitJobResponse{}, err
	}

	req.SubmittedBy = identity.Subject
	return c.service.Submit(req, now)
}

func (c *Controller) Get(rawToken, jobID string) (JobStoreRecord, error) {
	if _, err := c.verifyOperator(rawToken); err != nil {
		return JobStoreRecord{}, err
	}

	return c.service.Get(jobID)
}

func (c *Controller) ResultAccess(rawToken, jobID string, now time.Time) (ResultAccessRecord, error) {
	if _, err := c.verifyOperator(rawToken); err != nil {
		return ResultAccessRecord{}, err
	}

	return c.service.ResolveResultAccess(jobID, now)
}

func (c *Controller) Cancel(rawToken, jobID string, now time.Time) (CancelJobResponse, error) {
	if _, err := c.verifyOperator(rawToken); err != nil {
		return CancelJobResponse{}, err
	}

	return c.service.RequestCancel(jobID, now)
}

func (c *Controller) ClaimNext(rawToken string, requested WorkerIdentity, now time.Time) (ClaimNextResponse, error) {
	worker, err := c.verifyWorker(rawToken)
	if err != nil {
		return ClaimNextResponse{}, err
	}
	worker, err = mergeWorkerIdentity(worker, requested)
	if err != nil {
		return ClaimNextResponse{}, fmt.Errorf("verify worker: %w", err)
	}

	return c.service.ClaimNext(worker, now)
}

func (c *Controller) RenewLease(rawToken, jobID string, requested WorkerIdentity, now time.Time) (LeaseRenewResponse, error) {
	worker, err := c.verifyWorker(rawToken)
	if err != nil {
		return LeaseRenewResponse{}, err
	}
	worker, err = mergeWorkerIdentity(worker, requested)
	if err != nil {
		return LeaseRenewResponse{}, fmt.Errorf("verify worker: %w", err)
	}

	return c.service.RenewLease(jobID, worker, now)
}

func (c *Controller) UpdateStatus(rawToken, jobID string, update StatusUpdate, now time.Time) (JobStoreRecord, error) {
	worker, err := c.verifyWorker(rawToken)
	if err != nil {
		return JobStoreRecord{}, err
	}

	update.JobID = jobID
	update.WorkerID = worker.WorkerID
	return c.service.ApplyStatusUpdate(jobID, worker, update, now)
}

func (c *Controller) verifyOperator(rawToken string) (OperatorIdentity, error) {
	if c == nil || c.auth == nil {
		return OperatorIdentity{}, fmt.Errorf("verify operator: auth verifier not configured")
	}
	if c.service == nil {
		return OperatorIdentity{}, fmt.Errorf("verify operator: job control service not configured")
	}

	identity, err := c.auth.VerifyOperator(rawToken)
	if err != nil {
		return OperatorIdentity{}, fmt.Errorf("verify operator: %w", err)
	}
	if identity.Subject == "" {
		return OperatorIdentity{}, fmt.Errorf("verify operator: empty subject")
	}

	return identity, nil
}

func (c *Controller) verifyWorker(rawToken string) (WorkerIdentity, error) {
	if c == nil || c.auth == nil {
		return WorkerIdentity{}, fmt.Errorf("verify worker: auth verifier not configured")
	}
	if c.service == nil {
		return WorkerIdentity{}, fmt.Errorf("verify worker: job control service not configured")
	}

	identity, err := c.auth.VerifyWorker(rawToken)
	if err != nil {
		return WorkerIdentity{}, fmt.Errorf("verify worker: %w", err)
	}
	if identity.WorkerID == "" || identity.AuthSubject == "" {
		return WorkerIdentity{}, fmt.Errorf("verify worker: incomplete worker identity")
	}

	return identity, nil
}

func mergeWorkerIdentity(verified, requested WorkerIdentity) (WorkerIdentity, error) {
	resolved := verified

	if requested.WorkerID != "" && requested.WorkerID != verified.WorkerID {
		return WorkerIdentity{}, ErrUnauthorized
	}
	if requested.AuthSubject != "" && requested.AuthSubject != verified.AuthSubject {
		return WorkerIdentity{}, ErrUnauthorized
	}
	if requested.AuthMode != "" && requested.AuthMode != verified.AuthMode {
		return WorkerIdentity{}, ErrUnauthorized
	}

	if requested.AuthIssuer != "" {
		resolved.AuthIssuer = requested.AuthIssuer
	}
	if requested.MachineName != "" {
		resolved.MachineName = requested.MachineName
	}
	if requested.WorkerVersion != "" {
		resolved.WorkerVersion = requested.WorkerVersion
	}
	resolved.SupportsCancel = requested.SupportsCancel || resolved.SupportsCancel
	if len(requested.AcceptedJobTypes) > 0 {
		resolved.AcceptedJobTypes = append([]string(nil), requested.AcceptedJobTypes...)
	}
	if requested.MaxLeaseSeconds > 0 {
		resolved.MaxLeaseSeconds = requested.MaxLeaseSeconds
	}

	return resolved, nil
}
