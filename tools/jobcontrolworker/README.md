# Job Control Worker

This is the first real local-worker process for the hybrid job-control model.

It talks to the internal job-control API exposed by the frontend and performs a
small one-shot processing cycle:

1. claim the next available job
2. write a synthetic local-file result artifact
3. report a succeeded status update back to the control plane

## Scope

- one-shot or simple looping local worker process
- no queue
- no scheduler
- no advanced retry loop or supervisor
- no heavy business logic yet

## Environment variables

- `JOB_CONTROL_BASE_URL`
  - default: `http://127.0.0.1:8080`
- `JOB_CONTROL_WORKER_TOKEN`
  - required
- `JOB_CONTROL_WORKER_ID`
  - default: `local-worker-1`
- `JOB_CONTROL_WORKER_SUBJECT`
  - default: `worker:${JOB_CONTROL_WORKER_ID}`
- `JOB_CONTROL_WORKER_ISSUER`
  - default: `jobcontrolworker-local`
- `JOB_CONTROL_WORKER_MACHINE_NAME`
  - default: current hostname or `local-machine`
- `JOB_CONTROL_WORKER_VERSION`
  - default: `0.1.0`
- `JOB_CONTROL_WORKER_MAX_LEASE_SECONDS`
  - default: `300`
- `JOB_CONTROL_PROCESS_DELAY_SECONDS`
  - default: `0`
  - when greater than zero, simulates longer-running work and can trigger lease renewal
- `JOB_CONTROL_RUN_MODE`
  - default: `once`
  - supported values: `once`, `loop`
- `JOB_CONTROL_IDLE_POLL_SECONDS`
  - default: `5`
  - only used in `loop` mode when no job is available
- `JOB_CONTROL_MAX_JOBS`
  - default: `0`
  - when greater than zero in `loop` mode, exits after that many completed jobs
- `JOB_CONTROL_RESULT_DIR`
  - default: `/tmp/jobcontrol-results`

## Example

```bash
JOB_CONTROL_BASE_URL=http://127.0.0.1:18081 \
JOB_CONTROL_WORKER_TOKEN=worker-secret \
JOB_CONTROL_RESULT_DIR=/tmp/jobcontrol-results \
go run ./tools/jobcontrolworker
```

This is still a deliberately small worker. Its purpose is to replace the purely
in-process simulation path with a real second process while keeping the rest of
the system simple.

In `loop` mode it becomes a simple poller, but it is still not a full service
manager or daemon supervisor.
