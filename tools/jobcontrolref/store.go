package jobcontrolref

import frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"

type State = frontendjobcontrol.State

const (
	StateQueued    = frontendjobcontrol.StateQueued
	StateRunning   = frontendjobcontrol.StateRunning
	StateSucceeded = frontendjobcontrol.StateSucceeded
	StateFailed    = frontendjobcontrol.StateFailed
	StateCancelled = frontendjobcontrol.StateCancelled
)

type AuthMode = frontendjobcontrol.AuthMode

const AuthModeBearer = frontendjobcontrol.AuthModeBearer

type RetrievalMode = frontendjobcontrol.RetrievalMode

const RetrievalModeReferenceOnly = frontendjobcontrol.RetrievalModeReferenceOnly

var (
	ErrJobNotFound       = frontendjobcontrol.ErrJobNotFound
	ErrLeaseConflict     = frontendjobcontrol.ErrLeaseConflict
	ErrResultUnavailable = frontendjobcontrol.ErrResultUnavailable
)

const (
	DefaultPollAfterSeconds   = frontendjobcontrol.DefaultPollAfterSeconds
	DefaultLeaseSeconds       = frontendjobcontrol.DefaultLeaseSeconds
	DefaultResultTTLSeconds   = frontendjobcontrol.DefaultResultTTLSeconds
	DefaultRenewAfterFraction = frontendjobcontrol.DefaultRenewAfterFraction
)

type SubmitJobRequest = frontendjobcontrol.SubmitJobRequest
type SubmitJobResponse = frontendjobcontrol.SubmitJobResponse
type WorkerIdentity = frontendjobcontrol.WorkerIdentity
type Lease = frontendjobcontrol.Lease
type ResultReference = frontendjobcontrol.ResultReference
type ResultAccessRecord = frontendjobcontrol.ResultAccessRecord
type JobRecord = frontendjobcontrol.JobRecord
type JobStoreRecord = frontendjobcontrol.JobStoreRecord
type ClaimedJob = frontendjobcontrol.ClaimedJob
type ClaimNextResponse = frontendjobcontrol.ClaimNextResponse
type StatusUpdate = frontendjobcontrol.StatusUpdate
type CancelJobResponse = frontendjobcontrol.CancelJobResponse
type Store = frontendjobcontrol.Store

func NewStore() *Store {
	return frontendjobcontrol.NewStore()
}

func NewPersistentStore(statePath string) (*Store, error) {
	return frontendjobcontrol.NewPersistentStore(statePath)
}
