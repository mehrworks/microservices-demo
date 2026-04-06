package jobcontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	IdempotencyKey            string            `json:"idempotency_key,omitempty"`
	JobType                   string            `json:"job_type"`
	SubmittedBy               string            `json:"submitted_by"`
	Priority                  string            `json:"priority,omitempty"`
	RequestedTimeoutSeconds   int               `json:"requested_timeout_seconds,omitempty"`
	RequestedResultTTLSeconds int               `json:"requested_result_ttl_seconds,omitempty"`
	Labels                    map[string]string `json:"labels,omitempty"`
	Payload                   map[string]any    `json:"payload"`
}

type SubmitJobResponse struct {
	JobID  string `json:"job_id"`
	State  State  `json:"state"`
	JobRef string `json:"job_ref"`
}

type WorkerIdentity struct {
	WorkerID         string   `json:"worker_id"`
	AuthMode         AuthMode `json:"auth_mode"`
	AuthSubject      string   `json:"auth_subject"`
	AuthIssuer       string   `json:"auth_issuer,omitempty"`
	MachineName      string   `json:"machine_name,omitempty"`
	WorkerVersion    string   `json:"worker_version,omitempty"`
	SupportsCancel   bool     `json:"supports_cancel,omitempty"`
	AcceptedJobTypes []string `json:"accepted_job_types,omitempty"`
	MaxLeaseSeconds  int      `json:"max_lease_seconds,omitempty"`
}

type Lease struct {
	LeaseID              string    `json:"lease_id"`
	IssuedAt             time.Time `json:"issued_at"`
	LeaseExpiresAt       time.Time `json:"lease_expires_at"`
	LeaseDurationSeconds int       `json:"lease_duration_seconds"`
	RenewAfterSeconds    int       `json:"renew_after_seconds,omitempty"`
}

type ResultReference struct {
	Kind           string            `json:"kind"`
	URI            string            `json:"uri"`
	ContentType    string            `json:"content_type"`
	SizeBytes      int64             `json:"size_bytes,omitempty"`
	ChecksumSHA256 string            `json:"checksum_sha256,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ResultAccessRecord struct {
	ResultRef     ResultReference `json:"result_ref"`
	RetrievalMode RetrievalMode   `json:"retrieval_mode"`
	Available     bool            `json:"available"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
	SizeBytes     int64           `json:"size_bytes,omitempty"`
}

type JobRecord struct {
	JobID                string           `json:"job_id"`
	JobType              string           `json:"job_type"`
	SubmittedBy          string           `json:"submitted_by"`
	Attempt              int              `json:"attempt"`
	State                State            `json:"state"`
	SubmittedAt          time.Time        `json:"submitted_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	ProgressPercent      int              `json:"progress_percent,omitempty"`
	StatusMessage        string           `json:"status_message,omitempty"`
	ErrorCode            *string          `json:"error_code,omitempty"`
	ClaimedBy            *string          `json:"claimed_by,omitempty"`
	ClaimedByAuthSubject *string          `json:"claimed_by_auth_subject,omitempty"`
	Lease                *Lease           `json:"lease,omitempty"`
	CancelRequested      bool             `json:"cancel_requested,omitempty"`
	Payload              map[string]any   `json:"payload"`
	ResultRef            *ResultReference `json:"result_ref,omitempty"`
}

type JobStoreRecord struct {
	Job                JobRecord           `json:"job"`
	Revision           int                 `json:"revision"`
	ETag               string              `json:"etag"`
	ResultAccess       *ResultAccessRecord `json:"result_access,omitempty"`
	RetentionExpiresAt *time.Time          `json:"retention_expires_at,omitempty"`
}

type ClaimedJob struct {
	JobID                   string         `json:"job_id"`
	JobType                 string         `json:"job_type"`
	Attempt                 int            `json:"attempt"`
	Lease                   Lease          `json:"lease"`
	RequestedTimeoutSeconds int            `json:"requested_timeout_seconds,omitempty"`
	CancelRequested         bool           `json:"cancel_requested,omitempty"`
	Payload                 map[string]any `json:"payload"`
}

type ClaimNextResponse struct {
	WorkerID         string      `json:"worker_id"`
	PollAfterSeconds int         `json:"poll_after_seconds"`
	Job              *ClaimedJob `json:"job"`
}

type LeaseRenewResponse struct {
	JobID string `json:"job_id"`
	Lease Lease  `json:"lease"`
}

type StatusUpdate struct {
	JobID           string           `json:"job_id,omitempty"`
	WorkerID        string           `json:"worker_id,omitempty"`
	State           State            `json:"state"`
	UpdatedAt       time.Time        `json:"updated_at,omitempty"`
	ProgressPercent int              `json:"progress_percent,omitempty"`
	StatusMessage   string           `json:"status_message,omitempty"`
	ErrorCode       *string          `json:"error_code,omitempty"`
	ResultRef       *ResultReference `json:"result_ref,omitempty"`
}

type CancelJobResponse struct {
	JobID           string `json:"job_id"`
	CancelRequested bool   `json:"cancel_requested"`
	State           State  `json:"state"`
}

type Store struct {
	mu           sync.Mutex
	jobs         map[string]*storedJob
	idempotency  map[string]string
	nextSequence int
	statePath    string
	now          func() time.Time
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
		now:         time.Now,
	}
}

func NewPersistentStore(statePath string) (*Store, error) {
	store := NewStore()
	store.statePath = statePath

	if statePath == "" {
		return store, nil
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
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
	if err := s.saveLocked(); err != nil {
		return SubmitJobResponse{}, err
	}

	return SubmitJobResponse{JobID: jobID, State: StateQueued, JobRef: jobRef(jobID)}, nil
}

func (s *Store) ClaimNext(worker WorkerIdentity, now time.Time) (ClaimNextResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := s.expireLeasesLocked(now)

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
		if err := s.saveLocked(); err != nil {
			return ClaimNextResponse{}, err
		}

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

	if changed {
		if err := s.saveLocked(); err != nil {
			return ClaimNextResponse{}, err
		}
	}

	return ClaimNextResponse{WorkerID: worker.WorkerID, PollAfterSeconds: DefaultPollAfterSeconds, Job: nil}, nil
}

func (s *Store) RenewLease(jobID string, worker WorkerIdentity, now time.Time) (LeaseRenewResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.jobs[jobID]
	if !ok {
		return LeaseRenewResponse{}, ErrJobNotFound
	}
	if entry.record.State != StateRunning || entry.record.Lease == nil {
		return LeaseRenewResponse{}, ErrLeaseConflict
	}
	if entry.record.ClaimedBy == nil || *entry.record.ClaimedBy != worker.WorkerID {
		return LeaseRenewResponse{}, ErrLeaseConflict
	}
	if entry.record.ClaimedByAuthSubject == nil || *entry.record.ClaimedByAuthSubject != worker.AuthSubject {
		return LeaseRenewResponse{}, ErrLeaseConflict
	}
	if now.After(entry.record.Lease.LeaseExpiresAt) {
		entry.record.State = StateQueued
		entry.record.Attempt++
		entry.record.UpdatedAt = now
		entry.record.StatusMessage = "lease expired; awaiting reclaim"
		entry.record.ClaimedBy = nil
		entry.record.ClaimedByAuthSubject = nil
		entry.record.Lease = nil
		entry.revision++
		if err := s.saveLocked(); err != nil {
			return LeaseRenewResponse{}, err
		}
		return LeaseRenewResponse{}, ErrLeaseConflict
	}

	lease := newLease(jobID, worker, now)
	entry.record.Lease = &lease
	entry.record.UpdatedAt = now
	entry.record.StatusMessage = "lease renewed"
	entry.revision++
	if err := s.saveLocked(); err != nil {
		return LeaseRenewResponse{}, err
	}

	return LeaseRenewResponse{JobID: jobID, Lease: lease}, nil
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
	if err := s.saveLocked(); err != nil {
		return JobStoreRecord{}, err
	}
	return entry.snapshot(), nil
}

func (s *Store) RequestCancel(jobID string, now time.Time) (CancelJobResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.expireAndPersistLocked(now)

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
		// terminal state remains as-is
	}

	entry.revision++
	if err := s.saveLocked(); err != nil {
		return CancelJobResponse{}, err
	}
	return CancelJobResponse{JobID: jobID, CancelRequested: entry.record.CancelRequested, State: entry.record.State}, nil
}

func (s *Store) ResolveResultAccess(jobID string, now time.Time) (ResultAccessRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.expireAndPersistLocked(now)

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

	nowFn := s.now
	if nowFn == nil {
		nowFn = time.Now
	}
	_ = s.expireAndPersistLocked(nowFn())

	entry, ok := s.jobs[jobID]
	if !ok {
		return JobStoreRecord{}, ErrJobNotFound
	}

	return entry.snapshot(), nil
}

func (s *Store) expireAndPersistLocked(now time.Time) error {
	if changed := s.expireLeasesLocked(now); changed {
		return s.saveLocked()
	}
	return nil
}

func (s *Store) expireLeasesLocked(now time.Time) bool {
	changed := false
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
		changed = true
	}
	return changed
}

type persistedStore struct {
	NextSequence int                     `json:"next_sequence"`
	Idempotency  map[string]string       `json:"idempotency"`
	Jobs         map[string]persistedJob `json:"jobs"`
}

type persistedJob struct {
	Record                    JobRecord           `json:"record"`
	Revision                  int                 `json:"revision"`
	ResultAccess              *ResultAccessRecord `json:"result_access,omitempty"`
	RetentionExpiresAt        *time.Time          `json:"retention_expires_at,omitempty"`
	RequestedTimeoutSeconds   int                 `json:"requested_timeout_seconds"`
	RequestedResultTTLSeconds int                 `json:"requested_result_ttl_seconds"`
}

func (s *Store) saveLocked() error {
	if s.statePath == "" {
		return nil
	}

	snapshot := persistedStore{
		NextSequence: s.nextSequence,
		Idempotency:  make(map[string]string, len(s.idempotency)),
		Jobs:         make(map[string]persistedJob, len(s.jobs)),
	}
	for key, value := range s.idempotency {
		snapshot.Idempotency[key] = value
	}
	for jobID, entry := range s.jobs {
		snapshot.Jobs[jobID] = persistedJob{
			Record:                    cloneJobRecord(entry.record),
			Revision:                  entry.revision,
			ResultAccess:              cloneResultAccessPtr(entry.resultAccess),
			RetentionExpiresAt:        cloneTimePtr(entry.retentionExpiresAt),
			RequestedTimeoutSeconds:   entry.requestedTimeoutSeconds,
			RequestedResultTTLSeconds: entry.requestedResultTTLSeconds,
		}
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.statePath), 0o755); err != nil {
		return err
	}
	tmpPath := s.statePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.statePath)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var snapshot persistedStore
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	s.nextSequence = snapshot.NextSequence
	s.idempotency = make(map[string]string, len(snapshot.Idempotency))
	for key, value := range snapshot.Idempotency {
		s.idempotency[key] = value
	}
	s.jobs = make(map[string]*storedJob, len(snapshot.Jobs))
	for jobID, entry := range snapshot.Jobs {
		s.jobs[jobID] = &storedJob{
			record:                    cloneJobRecord(entry.Record),
			revision:                  entry.Revision,
			resultAccess:              cloneResultAccessPtr(entry.ResultAccess),
			retentionExpiresAt:        cloneTimePtr(entry.RetentionExpiresAt),
			requestedTimeoutSeconds:   entry.RequestedTimeoutSeconds,
			requestedResultTTLSeconds: entry.RequestedResultTTLSeconds,
		}
	}

	return nil
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

func cloneTimePtr(in *time.Time) *time.Time {
	if in == nil {
		return nil
	}
	v := *in
	return &v
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
