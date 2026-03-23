package main

import "testing"

func TestTracingEnabled(t *testing.T) {
	t.Setenv("ENABLE_TRACING", "")
	if tracingEnabled() {
		t.Fatal("expected tracing to be disabled by default")
	}

	t.Setenv("ENABLE_TRACING", "1")
	if !tracingEnabled() {
		t.Fatal("expected tracing to be enabled when ENABLE_TRACING=1")
	}
}

func TestProfilerEnabled(t *testing.T) {
	t.Setenv("ENABLE_PROFILER", "")
	if profilerEnabled() {
		t.Fatal("expected profiler to be disabled by default")
	}

	t.Setenv("ENABLE_PROFILER", "1")
	if !profilerEnabled() {
		t.Fatal("expected profiler to be enabled when ENABLE_PROFILER=1")
	}
}

func TestClientOptions(t *testing.T) {
	t.Setenv("ENABLE_TRACING", "")
	if got := len(clientOptions()); got != 1 {
		t.Fatalf("expected transport-only client options when tracing is disabled, got %d", got)
	}

	t.Setenv("ENABLE_TRACING", "1")
	if got := len(clientOptions()); got != 2 {
		t.Fatalf("expected tracing client handler when tracing is enabled, got %d options", got)
	}
}

func TestServerOptions(t *testing.T) {
	t.Setenv("ENABLE_TRACING", "")
	if got := len(serverOptions()); got != 0 {
		t.Fatalf("expected no grpc server options when tracing is disabled, got %d", got)
	}

	t.Setenv("ENABLE_TRACING", "1")
	if got := len(serverOptions()); got != 1 {
		t.Fatalf("expected otel grpc server handler when tracing is enabled, got %d options", got)
	}
}
