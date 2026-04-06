package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	frontendjobcontrol "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/jobcontrol"
)

func main() {
	baseURL := getenvDefault("JOB_CONTROL_BASE_URL", "http://127.0.0.1:8080")
	workerToken := os.Getenv("JOB_CONTROL_WORKER_TOKEN")
	if workerToken == "" {
		fmt.Fprintln(os.Stderr, "error: JOB_CONTROL_WORKER_TOKEN must be set")
		os.Exit(1)
	}
	resultDir := getenvDefault("JOB_CONTROL_RESULT_DIR", "/tmp/jobcontrol-results")
	workerID := getenvDefault("JOB_CONTROL_WORKER_ID", "local-worker-1")
	workerSubject := getenvDefault("JOB_CONTROL_WORKER_SUBJECT", "worker:"+workerID)
	workerIssuer := getenvDefault("JOB_CONTROL_WORKER_ISSUER", "jobcontrolworker-local")
	workerMachine := getenvDefault("JOB_CONTROL_WORKER_MACHINE_NAME", hostnameOr("local-machine"))
	workerVersion := getenvDefault("JOB_CONTROL_WORKER_VERSION", "0.1.0")
	workerLease := getenvIntDefault("JOB_CONTROL_WORKER_MAX_LEASE_SECONDS", frontendjobcontrol.DefaultLeaseSeconds)
	processDelay := getenvIntDefault("JOB_CONTROL_PROCESS_DELAY_SECONDS", 0)
	runMode := strings.ToLower(getenvDefault("JOB_CONTROL_RUN_MODE", "once"))
	idlePoll := getenvIntDefault("JOB_CONTROL_IDLE_POLL_SECONDS", 5)
	maxJobs := getenvIntDefaultAllowZero("JOB_CONTROL_MAX_JOBS", 0)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := &Worker{
		Client: &Client{
			BaseURL:     baseURL,
			WorkerToken: workerToken,
			HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		},
		Identity: frontendjobcontrol.WorkerIdentity{
			WorkerID:         workerID,
			AuthMode:         frontendjobcontrol.AuthModeBearer,
			AuthSubject:      workerSubject,
			AuthIssuer:       workerIssuer,
			MachineName:      workerMachine,
			WorkerVersion:    workerVersion,
			SupportsCancel:   true,
			AcceptedJobTypes: []string{"local-analysis"},
			MaxLeaseSeconds:  workerLease,
		},
		ResultDir:    resultDir,
		Now:          time.Now,
		ProcessDelay: time.Duration(processDelay) * time.Second,
	}

	switch runMode {
	case "once":
		jobID, processed, err := worker.RunOnce(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !processed {
			fmt.Println("no job available")
			return
		}
		fmt.Printf("completed job %s\n", jobID)
	case "loop":
		err := worker.RunLoop(ctx, LoopConfig{
			IdlePollInterval: time.Duration(idlePoll) * time.Second,
			MaxJobs:          maxJobs,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("worker loop complete")
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported JOB_CONTROL_RUN_MODE %q\n", runMode)
		os.Exit(1)
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvIntDefault(key string, fallback int) int {
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

func getenvIntDefaultAllowZero(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func hostnameOr(fallback string) string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return fallback
	}
	return host
}
