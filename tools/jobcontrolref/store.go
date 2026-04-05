package jobcontrolref

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	DefaultPollAfterSeconds   = 5
	DefaultLeaseSeconds       = 300
	DefaultResultTTLSeconds   = 86400
	DefaultRenewAfterFraction = 4
)

var (
	ErrJobNotFound       = errors.New("job not found")
	ErrLeaseConflict     = errors.New("lease conflict")
	ErrResultUnavailable = errors.New("result unavailable")
)

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

type AuthMode string

const AuthModeBearer AuthMode = "bearer"

type RetrievalMode string

const RetrievalModeReferenceOnly RetrievalMode = "reference-only"

type SubmitJobRequest struct {
	IdempotencyKey            string
	JobType                   string
	SubmittedBy               string
	Priority                  string
	RequestedTimeoutSeconds   int
	RequestedResultTTLSeconds int
	Labels                    map[string]string
	Payload                   map[string]any
}

type SubmitJobResponse struct {
	JobID  string
	State  State
	JobRef string
}

type WorkerIdentity struct {
	WorkerID         string
	AuthMode         AuthMode
	AuthSubject      string
	AuthIssuer       string
	MachineName      string
	WorkerVersion    string
	SupportsCancel   bool
	AcceptedJobTypes []string
	MaxLeaseSeconds  int
}

type Lease struct {
	LeaseID              string
	IssuedAt             time.Time
	LeaseExpiresAt       time.Time
	LeaseDurationSeconds int
	RenewAfterSeconds    int
}

type ResultReference struct {
	Kind           string
	URI            string
	ContentType    string
	SizeBytes      int64
	ChecksumSHA256 string
	Metadata       map[string]string
}

type ResultAccessRecord struct {
	ResultRef     ResultReference
	RetrievalMode RetrievalMode
	Available     bool
	ExpiresAt     *time.Time
	SizeBytes     int64
}

type JobRecord struct {
	JobID                string
	JobType              string
	SubmittedBy          string
	Attempt              int
	State                State
	SubmittedAt          time.Time
	UpdatedAt            time.Time
	ProgressPercent      int
	StatusMessage        string
	ErrorCode            *string
	ClaimedBy            *string
	ClaimedByAuthSubject *string
	Lease                *Lease
	CancelRequested      bool
	Payload              map[string]any
	ResultRef            *ResultReference
}

type JobStoreRecord struct {
	Job                JobRecord
	Revision           int
	ETag               string
	ResultAccess       *ResultAccessRecord
	RetentionExpiresAt *time.Time
}

type ClaimedJob struct {
	JobID                   string
	JobType                 string
	Attempt                 int
	Lease                   Lease
	RequestedTimeoutSeconds int
	CancelRequested         bool
	Payload                 map[string]any
}

type ClaimNextResponse struct {
	WorkerID         string
	PollAfterSeconds int
	Job              *ClaimedJob
}

type StatusUpdate struct {
	JobID           string
	WorkerID        string
	State           State
	UpdatedAt       time.Time
	ProgressPercent int
	StatusMessage   string
	ErrorCode       *string
	ResultRef       *ResultReference
}

type CancelJobResponse struct {
	JobID           string
	CancelRequested bool
	State           State
}

type Store struct {
	mu           sync.Mutex
	jobs         map[string]*storedJob
	idempotency  map[string]string
	nextSequence int
}

type storedJob struct {
	record                    JobRecord
	revision                  int
	resultAccess              *ResultAccessRecord
	retentionExpiresAt        *time.Time
	requestedTimeoutSeconds   int
	requestedResultTTLSeconds int
}

func NewStore() *Store {
	return &Store{
		jobs:        make(map[string]*storedJob),
		idempotency: make(map[string]string),
	}
}

func (s *Store) Submit(req SubmitJobRequest, now time.Time) (SubmitJobResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.IdempotencyKey != "" {
		if jobID, ok := s.idempotency[req.IdempotencyKey]; ok {
			entry := s.jobs[jobID]
			return SubmitJobResponse{JobID: entry.record.JobID, State: entry.record.State, JobRef: jobRef(jobID)}, nil
		}
	}

	s.nextSequence++
	jobID := fmt.Sprintf("job_%04d", s.nextSequence)
	timeout := req.RequestedTimeoutSeconds
	if timeout <= 0 {
		timeout = 1800
	}
	resultTTL := req.RequestedResultTTLSeconds
	if resultTTL <= 0 {
		resultTTL = DefaultResultTTLSeconds
	}

	entry := &storedJob{
		record: JobRecord{
			JobID:           jobID,
			JobType:         req.JobType,
			SubmittedBy:     req.SubmittedBy,
			Attempt:         1,
			State:           StateQueued,
			SubmittedAt:     now,
			UpdatedAt:       now,
			ProgressPercent: 0,
			Payload:         cloneAnyMap(req.Payload),
		},
		revision:                  1,
		requestedTimeoutSeconds:   timeout,
		requestedResultTTLSeconds: resultTTL,
	}

	s.jobs[jobID] = entry
	if req.IdempotencyKey != "" {
		s.idempotency[req.IdempotencyKey] = jobID
	}

	return SubmitJobResponse{JobID: jobID, State: StateQueued, JobRef: jobRef(jobID)}, nil
}

func (s *Store) ClaimNext(worker WorkerIdentity, now time.Time) (ClaimNextResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expireLeasesLocked(now)

	jobIDs := make([]string, 0, len(s.jobs))
	for jobID := range s.jobs {
		jobIDs = append(jobIDs, jobID)
	}
	sort.Slice(jobIDs, func(i, j int) bool {
		return s.jobs[jobIDs[i]].record.SubmittedAt.Before(s.jobs[jobIDs[j]].record.SubmittedAt)
	})

	for _, jobID := range jobIDs {
		entry := s.jobs[jobID]
		if entry.record.State != StateQueued {
			continue
		}
		if !acceptsJobType(worker.AcceptedJobTypes, entry.record.JobType) {
			continue
		}

		lease := newLease(jobID, worker, now)
		entry.record.State = StateRunning
		entry.record.UpdatedAt = now
		entry.record.StatusMessage = "claimed by worker"
		entry.record.ClaimedBy = stringPtr(worker.WorkerID)
		entry.record.ClaimedByAuthSubject = stringPtr(worker.AuthSubject)
		entry.record.Lease = &lease
		entry.revision++

		return ClaimNextResponse{
			WorkerID:         worker.WorkerID,
			PollAfterSeconds: DefaultPollAfterSeconds,
			Job: &ClaimedJob{
				JobID:                   entry.record.JobID,
				JobType:                 entry.record.JobType,
				Attempt:                 entry.record.Attempt,
				Lease:                   lease,
				RequestedTimeoutSeconds: entry.requestedTimeoutSeconds,
				CancelRequested:         entry.record.CancelRequested,
				Payload:                 cloneAnyMap(entry.record.Payload),
			},
		}, nil
	}

	return ClaimNextResponse{WorkerID: worker.WorkerID, PollAfterSeconds: DefaultPollAfterSeconds, Job: nil}, nil
}

func (s *Store) ApplyStatusUpdate(jobID string, worker WorkerIdentity, update StatusUpdate, now time.Time) (JobStoreRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.jobs[jobID]
	if !ok {
		return JobStoreRecord{}, ErrJobNotFound
	}
	if entry.record.Lease == nil || entry.record.State != StateRunning {
		return JobStoreRecord{}, ErrLeaseConflict
	}
	if entry.record.ClaimedBy == nil || *entry.record.ClaimedBy != worker.WorkerID {
		return JobStoreRecord{}, ErrLeaseConflict
	}
	if entry.record.ClaimedByAuthSubject == nil || *entry.record.ClaimedByAuthSubject != worker.AuthSubject {
		return JobStoreRecord{}, ErrLeaseConflict
	}
	if now.After(entry.record.Lease.LeaseExpiresAt) {
		return JobStoreRecord{}, ErrLeaseConflict
	}

	entry.record.UpdatedAt = now
	entry.record.ProgressPercent = update.ProgressPercent
	entry.record.StatusMessage = update.StatusMessage
	entry.record.ErrorCode = update.ErrorCode

	switch update.State {
	case StateRunning:
		entry.record.State = StateRunning
	case StateSucceeded, StateFailed, StateCancelled:
		entry.record.State = update.State
		entry.record.Lease = nil
		if update.ResultRef != nil {
			entry.record.ResultRef = cloneResultRef(update.ResultRef)
			expiresAt := now.Add(time.Duration(entry.requestedResultTTLSeconds) * time.Second)
			entry.retentionExpiresAt = &expiresAt
			entry.resultAccess = &ResultAccessRecord{
				ResultRef:     *cloneResultRef(update.ResultRef),
				RetrievalMode: RetrievalModeReferenceOnly,
				Available:     true,
				ExpiresAt:     &expiresAt,
				SizeBytes:     update.ResultRef.SizeBytes,
			}
		}
	default:
		return JobStoreRecord{}, fmt.Errorf("unsupported status transition: %s", update.State)
	}

	entry.revision++
	return entry.snapshot(), nil
}

func (s *Store) RequestCancel(jobID string, now time.Time) (CancelJobResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.jobs[jobID]
	if !ok {
		return CancelJobResponse{}, ErrJobNotFound
	}

	entry.record.UpdatedAt = now
	switch entry.record.State {
	case StateQueued:
		entry.record.State = StateCancelled
		entry.record.CancelRequested = true
	case StateRunning:
		entry.record.CancelRequested = true
	case StateSucceeded, StateFailed, StateCancelled:
		// leave terminal state as-is
	}

	entry.revision++
	return CancelJobResponse{JobID: jobID, CancelRequested: entry.record.CancelRequested, State: entry.record.State}, nil
}

func (s *Store) ResolveResultAccess(jobID string, now time.Time) (ResultAccessRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.jobs[jobID]
	if !ok {
		return ResultAccessRecord{}, ErrJobNotFound
	}

	if entry.resultAccess != nil {
		if entry.resultAccess.ExpiresAt != nil && now.After(*entry.resultAccess.ExpiresAt) {
			return ResultAccessRecord{}, ErrResultUnavailable
		}
		return cloneResultAccess(entry.resultAccess), nil
	}

	if entry.record.State == StateSucceeded && entry.record.ResultRef != nil {
		return ResultAccessRecord{
			ResultRef:     *cloneResultRef(entry.record.ResultRef),
			RetrievalMode: RetrievalModeReferenceOnly,
			Available:     true,
			ExpiresAt:     entry.retentionExpiresAt,
			SizeBytes:     entry.record.ResultRef.SizeBytes,
		}, nil
	}

	return ResultAccessRecord{}, ErrResultUnavailable
}

func (s *Store) Get(jobID string) (JobStoreRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.jobs[jobID]
	if !ok {
		return JobStoreRecord{}, ErrJobNotFound
	}

	return entry.snapshot(), nil
}

func (s *Store) expireLeasesLocked(now time.Time) {
	for _, entry := range s.jobs {
		if entry.record.State != StateRunning || entry.record.Lease == nil {
			continue
		}
		if now.Before(entry.record.Lease.LeaseExpiresAt) {
			continue
		}

		entry.record.State = StateQueued
		entry.record.Attempt++
		entry.record.UpdatedAt = now
		entry.record.StatusMessage = "lease expired; awaiting reclaim"
		entry.record.ClaimedBy = nil
		entry.record.ClaimedByAuthSubject = nil
		entry.record.Lease = nil
		entry.revision++
	}
}

func (e *storedJob) snapshot() JobStoreRecord {
	var retention *time.Time
	if e.retentionExpiresAt != nil {
		copyValue := *e.retentionExpiresAt
		retention = &copyValue
	}

	return JobStoreRecord{
		Job:                cloneJobRecord(e.record),
		Revision:           e.revision,
		ETag:               fmt.Sprintf("%s:r%d", e.record.JobID, e.revision),
		ResultAccess:       cloneResultAccessPtr(e.resultAccess),
		RetentionExpiresAt: retention,
	}
}

func jobRef(jobID string) string {
	return fmt.Sprintf("/v1/jobs/%s", jobID)
}

func acceptsJobType(accepted []string, jobType string) bool {
	if len(accepted) == 0 {
		return true
	}
	for _, candidate := range accepted {
		if candidate == jobType {
			return true
		}
	}
	return false
}

func newLease(jobID string, worker WorkerIdentity, now time.Time) Lease {
	leaseSeconds := worker.MaxLeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = DefaultLeaseSeconds
	}
	renewAfter := leaseSeconds * DefaultRenewAfterFraction / 5
	if renewAfter <= 0 {
		renewAfter = leaseSeconds
	}

	return Lease{
		LeaseID:              fmt.Sprintf("lease_%s_%s", jobID, worker.WorkerID),
		IssuedAt:             now,
		LeaseExpiresAt:       now.Add(time.Duration(leaseSeconds) * time.Second),
		LeaseDurationSeconds: leaseSeconds,
		RenewAfterSeconds:    renewAfter,
	}
}

func stringPtr(v string) *string {
	return &v
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneResultRef(in *ResultReference) *ResultReference {
	if in == nil {
		return nil
	}
	out := *in
	out.Metadata = cloneStringMap(in.Metadata)
	return &out
}

func cloneResultAccess(in *ResultAccessRecord) ResultAccessRecord {
	out := *cloneResultAccessPtr(in)
	return out
}

func cloneResultAccessPtr(in *ResultAccessRecord) *ResultAccessRecord {
	if in == nil {
		return nil
	}
	out := *in
	out.ResultRef = *cloneResultRef(&in.ResultRef)
	if in.ExpiresAt != nil {
		expires := *in.ExpiresAt
		out.ExpiresAt = &expires
	}
	return &out
}

func cloneLease(in *Lease) *Lease {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneJobRecord(in JobRecord) JobRecord {
	out := in
	out.Payload = cloneAnyMap(in.Payload)
	out.ResultRef = cloneResultRef(in.ResultRef)
	out.Lease = cloneLease(in.Lease)
	if in.ClaimedBy != nil {
		v := *in.ClaimedBy
		out.ClaimedBy = &v
	}
	if in.ClaimedByAuthSubject != nil {
		v := *in.ClaimedByAuthSubject
		out.ClaimedByAuthSubject = &v
	}
	if in.ErrorCode != nil {
		v := *in.ErrorCode
		out.ErrorCode = &v
	}
	return out
}
