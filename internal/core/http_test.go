package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	result := ProbeHTTP(context.Background(), []string{server.URL}, 2*time.Second, false)
	if result.Error != "" {
		t.Fatal(result.Error)
	}
	if result.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d", result.StatusCode)
	}
	if result.ResponseHeaders["X-Content-Type-Options"] != "nosniff" {
		t.Fatalf("header missing: %#v", result.ResponseHeaders)
	}
}

func TestBenchmarkHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	result, err := BenchmarkHTTP(context.Background(), server.URL, "GET", 10, 2, 2*time.Second, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Successes != 10 || result.Failures != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
}
