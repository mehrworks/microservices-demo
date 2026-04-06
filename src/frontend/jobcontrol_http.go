package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

const (
	jobControlOperatorTokenEnv = "JOB_CONTROL_OPERATOR_TOKEN"
	jobControlWorkerTokenEnv   = "JOB_CONTROL_WORKER_TOKEN"
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
	controller := frontendjobcontrol.NewController(
		frontendjobcontrol.NewStaticBearerVerifier(operatorToken, workerToken),
		frontendjobcontrol.NewStore(),
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
	r.HandleFunc(baseURL+"/internal/worker/claim", fe.claimJobHandler).Methods(http.MethodPost)
	r.HandleFunc(baseURL+"/internal/jobs/{job_id}/status", fe.updateJobStatusHandler).Methods(http.MethodPost)
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

	resp, err := fe.jobControlController.ClaimNext(extractBearerToken(r.Header.Get("Authorization")), time.Now())
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
