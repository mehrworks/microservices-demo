package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

const (
	jobControlOperatorTokenEnv   = "JOB_CONTROL_OPERATOR_TOKEN"
	jobControlOperatorSubjectEnv = "JOB_CONTROL_OPERATOR_SUBJECT"
	jobControlWorkerTokenEnv     = "JOB_CONTROL_WORKER_TOKEN"
	jobControlStatePathEnv       = "JOB_CONTROL_STATE_PATH"
	jobControlWorkerIDEnv        = "JOB_CONTROL_WORKER_ID"
	jobControlWorkerSubjectEnv   = "JOB_CONTROL_WORKER_SUBJECT"
	jobControlWorkerIssuerEnv    = "JOB_CONTROL_WORKER_ISSUER"
	jobControlWorkerMachineEnv   = "JOB_CONTROL_WORKER_MACHINE_NAME"
	jobControlWorkerVersionEnv   = "JOB_CONTROL_WORKER_VERSION"
	jobControlWorkerLeaseEnv     = "JOB_CONTROL_WORKER_MAX_LEASE_SECONDS"
)

type jobControlErrorEnvelope struct {
	ErrorCode string         `json:"error_code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

func configureJobControlController(log logrus.FieldLogger) *frontendjobcontrol.Controller {
	operatorToken := os.Getenv(jobControlOperatorTokenEnv)
	if operatorToken == "" {
		log.Info("job control internal API disabled; operator token not configured")
		return nil
	}

	workerToken := os.Getenv(jobControlWorkerTokenEnv)
	statePath := os.Getenv(jobControlStatePathEnv)
	store, err := frontendjobcontrol.NewPersistentStore(statePath)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize job control store")
	}
	if statePath == "" {
		log.Info("job control using in-memory store")
	} else {
		log.WithField("job_control_state_path", statePath).Info("job control using file-backed store")
	}
	controller := frontendjobcontrol.NewController(
		frontendjobcontrol.NewStaticBearerVerifierFromConfig(frontendjobcontrol.StaticBearerConfig{
			OperatorToken:   operatorToken,
			OperatorSubject: getenvDefault(jobControlOperatorSubjectEnv, "operator:static-bearer"),
			WorkerToken:     workerToken,
			WorkerIdentity: frontendjobcontrol.WorkerIdentity{
				WorkerID:        getenvDefault(jobControlWorkerIDEnv, "worker:static-bearer"),
				AuthMode:        frontendjobcontrol.AuthModeBearer,
				AuthSubject:     getenvDefault(jobControlWorkerSubjectEnv, getenvDefault(jobControlWorkerIDEnv, "worker:static-bearer")),
				AuthIssuer:      os.Getenv(jobControlWorkerIssuerEnv),
				MachineName:     os.Getenv(jobControlWorkerMachineEnv),
				WorkerVersion:   os.Getenv(jobControlWorkerVersionEnv),
				MaxLeaseSeconds: getenvInt(jobControlWorkerLeaseEnv, frontendjobcontrol.DefaultLeaseSeconds),
			},
		}),
		store,
	)
	log.Info("job control internal API enabled")
	return controller
}

func registerJobControlRoutes(r *mux.Router, baseURL string, fe *frontendServer) {
	if fe == nil || fe.jobControlController == nil {
		return
	}

	r.HandleFunc(baseURL+"/internal/jobs", fe.submitJobHandler).Methods(http.MethodPost)
	r.HandleFunc(baseURL+"/internal/jobs/{job_id}", fe.getJobHandler).Methods(http.MethodGet)
	r.HandleFunc(baseURL+"/internal/jobs/{job_id}/result", fe.getJobResultHandler).Methods(http.MethodGet)
	r.HandleFunc(baseURL+"/internal/jobs/{job_id}:cancel", fe.cancelJobHandler).Methods(http.MethodPost)
	if os.Getenv(jobControlWorkerTokenEnv) != "" {
		r.HandleFunc(baseURL+"/internal/worker/claim", fe.claimJobHandler).Methods(http.MethodPost)
		r.HandleFunc(baseURL+"/internal/jobs/{job_id}/lease:renew", fe.renewLeaseHandler).Methods(http.MethodPost)
		r.HandleFunc(baseURL+"/internal/jobs/{job_id}/status", fe.updateJobStatusHandler).Methods(http.MethodPost)
	}
}

func (fe *frontendServer) submitJobHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)

	var req frontendjobcontrol.SubmitJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJobControlError(log, w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON", false, nil)
		return
	}

	resp, err := fe.jobControlController.Submit(extractBearerToken(r.Header.Get("Authorization")), req, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusAccepted, resp)
}

func (fe *frontendServer) getJobHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)
	jobID := mux.Vars(r)["job_id"]

	resp, err := fe.jobControlController.Get(extractBearerToken(r.Header.Get("Authorization")), jobID)
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusOK, resp)
}

func (fe *frontendServer) cancelJobHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)
	jobID := mux.Vars(r)["job_id"]

	resp, err := fe.jobControlController.Cancel(extractBearerToken(r.Header.Get("Authorization")), jobID, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusAccepted, resp)
}

func (fe *frontendServer) claimJobHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)

	var requested frontendjobcontrol.WorkerIdentity
	if err := json.NewDecoder(r.Body).Decode(&requested); err != nil && !errors.Is(err, io.EOF) {
		writeJobControlError(log, w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON", false, nil)
		return
	}

	resp, err := fe.jobControlController.ClaimNext(extractBearerToken(r.Header.Get("Authorization")), requested, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusOK, resp)
}

func (fe *frontendServer) renewLeaseHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)
	jobID := mux.Vars(r)["job_id"]

	var requested frontendjobcontrol.WorkerIdentity
	if err := json.NewDecoder(r.Body).Decode(&requested); err != nil && !errors.Is(err, io.EOF) {
		writeJobControlError(log, w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON", false, nil)
		return
	}

	resp, err := fe.jobControlController.RenewLease(extractBearerToken(r.Header.Get("Authorization")), jobID, requested, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusOK, resp)
}

func (fe *frontendServer) updateJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)
	jobID := mux.Vars(r)["job_id"]

	var update frontendjobcontrol.StatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJobControlError(log, w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON", false, nil)
		return
	}

	resp, err := fe.jobControlController.UpdateStatus(extractBearerToken(r.Header.Get("Authorization")), jobID, update, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusOK, resp)
}

func (fe *frontendServer) getJobResultHandler(w http.ResponseWriter, r *http.Request) {
	log := jobControlRequestLogger(r)
	jobID := mux.Vars(r)["job_id"]

	resp, err := fe.jobControlController.ResultAccess(extractBearerToken(r.Header.Get("Authorization")), jobID, time.Now())
	if err != nil {
		writeJobControlControllerError(log, w, err)
		return
	}

	writeJobControlJSON(w, http.StatusOK, resp)
}

func jobControlRequestLogger(r *http.Request) logrus.FieldLogger {
	if log, ok := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger); ok && log != nil {
		return log
	}
	return logrus.New()
}

func writeJobControlJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJobControlControllerError(log logrus.FieldLogger, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, frontendjobcontrol.ErrUnauthorized):
		writeJobControlError(log, w, http.StatusUnauthorized, "unauthorized", "invalid or missing bearer token", false, nil)
	case errors.Is(err, frontendjobcontrol.ErrJobNotFound):
		writeJobControlError(log, w, http.StatusNotFound, "job_not_found", "job record was not found", false, nil)
	case errors.Is(err, frontendjobcontrol.ErrLeaseConflict):
		writeJobControlError(log, w, http.StatusConflict, "lease_conflict", "worker lease is invalid or expired", true, nil)
	case errors.Is(err, frontendjobcontrol.ErrResultUnavailable):
		writeJobControlError(log, w, http.StatusNotFound, "result_unavailable", "result is not available", false, nil)
	default:
		writeJobControlError(log, w, http.StatusInternalServerError, "internal_error", err.Error(), false, nil)
	}
}

func writeJobControlError(log logrus.FieldLogger, w http.ResponseWriter, status int, code, message string, retryable bool, details map[string]any) {
	log.WithFields(logrus.Fields{"error_code": code, "http_status": status}).Warn(message)
	writeJobControlJSON(w, status, jobControlErrorEnvelope{
		ErrorCode: code,
		Message:   message,
		Retryable: retryable,
		Details:   details,
	})
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
