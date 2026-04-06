package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

func TestJobControlInternalRoutesSubmitGetCancel(t *testing.T) {
	controller := frontendjobcontrol.NewController(
		frontendjobcontrol.NewStaticBearerVerifier("operator-secret", ""),
		frontendjobcontrol.NewStore(),
	)

	fe := &frontendServer{jobControlController: controller}
	router := mux.NewRouter()
	registerJobControlRoutes(router, "", fe)

	submitReq := httptest.NewRequest(http.MethodPost, "/internal/jobs", strings.NewReader(`{"job_type":"local-analysis","payload":{"input":"demo"}}`))
	submitReq.Header.Set("Authorization", "Bearer operator-secret")
	submitReq.Header.Set("Content-Type", "application/json")
	submitRec := httptest.NewRecorder()
	router.ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusAccepted {
		t.Fatalf("submit status = %d, want %d", submitRec.Code, http.StatusAccepted)
	}

	var submitResp frontendjobcontrol.SubmitJobResponse
	if err := json.NewDecoder(submitRec.Body).Decode(&submitResp); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if submitResp.JobID == "" {
		t.Fatal("submit response missing job id")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/internal/jobs/"+submitResp.JobID, nil)
	getReq.Header.Set("Authorization", "Bearer operator-secret")
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	var getResp frontendjobcontrol.JobStoreRecord
	if err := json.NewDecoder(getRec.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.Job.JobID != submitResp.JobID {
		t.Fatalf("get job id = %q, want %q", getResp.Job.JobID, submitResp.JobID)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/internal/jobs/"+submitResp.JobID+":cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer operator-secret")
	cancelRec := httptest.NewRecorder()
	router.ServeHTTP(cancelRec, cancelReq)

	if cancelRec.Code != http.StatusAccepted {
		t.Fatalf("cancel status = %d, want %d", cancelRec.Code, http.StatusAccepted)
	}

	var cancelResp frontendjobcontrol.CancelJobResponse
	if err := json.NewDecoder(cancelRec.Body).Decode(&cancelResp); err != nil {
		t.Fatalf("decode cancel response: %v", err)
	}
	if !cancelResp.CancelRequested {
		t.Fatal("cancel response did not request cancellation")
	}
}

func TestJobControlInternalRoutesRequireBearerToken(t *testing.T) {
	controller := frontendjobcontrol.NewController(
		frontendjobcontrol.NewStaticBearerVerifier("operator-secret", ""),
		frontendjobcontrol.NewStore(),
	)

	fe := &frontendServer{jobControlController: controller}
	router := mux.NewRouter()
	registerJobControlRoutes(router, "", fe)

	req := httptest.NewRequest(http.MethodGet, "/internal/jobs/job_0001", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Fatalf("body = %q, want unauthorized error payload", rec.Body.String())
	}
}

func TestJobControlInternalWorkerSimulationFlow(t *testing.T) {
	controller := frontendjobcontrol.NewController(
		frontendjobcontrol.NewStaticBearerVerifier("operator-secret", "worker-secret"),
		frontendjobcontrol.NewStore(),
	)

	fe := &frontendServer{jobControlController: controller}
	router := mux.NewRouter()
	registerJobControlRoutes(router, "", fe)

	submitReq := httptest.NewRequest(http.MethodPost, "/internal/jobs", strings.NewReader(`{"job_type":"local-analysis","payload":{"input":"demo"}}`))
	submitReq.Header.Set("Authorization", "Bearer operator-secret")
	submitReq.Header.Set("Content-Type", "application/json")
	submitRec := httptest.NewRecorder()
	router.ServeHTTP(submitRec, submitReq)

	var submitResp frontendjobcontrol.SubmitJobResponse
	if err := json.NewDecoder(submitRec.Body).Decode(&submitResp); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}

	claimReq := httptest.NewRequest(http.MethodPost, "/internal/worker/claim", nil)
	claimReq.Header.Set("Authorization", "Bearer worker-secret")
	claimRec := httptest.NewRecorder()
	router.ServeHTTP(claimRec, claimReq)

	if claimRec.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d", claimRec.Code, http.StatusOK)
	}

	var claimResp frontendjobcontrol.ClaimNextResponse
	if err := json.NewDecoder(claimRec.Body).Decode(&claimResp); err != nil {
		t.Fatalf("decode claim response: %v", err)
	}
	if claimResp.Job == nil || claimResp.Job.JobID != submitResp.JobID {
		t.Fatalf("claim response = %#v, want claimed job %q", claimResp, submitResp.JobID)
	}

	statusReq := httptest.NewRequest(http.MethodPost, "/internal/jobs/"+submitResp.JobID+"/status", strings.NewReader(`{"state":"succeeded","progress_percent":100,"status_message":"done","result_ref":{"kind":"local-file","uri":"file:///tmp/output.json","content_type":"application/json","size_bytes":42}}`))
	statusReq.Header.Set("Authorization", "Bearer worker-secret")
	statusReq.Header.Set("Content-Type", "application/json")
	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status update code = %d, want %d", statusRec.Code, http.StatusOK)
	}

	resultReq := httptest.NewRequest(http.MethodGet, "/internal/jobs/"+submitResp.JobID+"/result", nil)
	resultReq.Header.Set("Authorization", "Bearer operator-secret")
	resultRec := httptest.NewRecorder()
	router.ServeHTTP(resultRec, resultReq)

	if resultRec.Code != http.StatusOK {
		t.Fatalf("result code = %d, want %d", resultRec.Code, http.StatusOK)
	}

	var resultResp frontendjobcontrol.ResultAccessRecord
	if err := json.NewDecoder(resultRec.Body).Decode(&resultResp); err != nil {
		t.Fatalf("decode result response: %v", err)
	}
	if !resultResp.Available {
		t.Fatal("result access unavailable, want available")
	}
	if resultResp.ResultRef.URI != "file:///tmp/output.json" {
		t.Fatalf("result uri = %q, want %q", resultResp.ResultRef.URI, "file:///tmp/output.json")
	}
}
