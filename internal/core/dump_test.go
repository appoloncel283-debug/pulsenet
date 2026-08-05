package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDumpSite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Set-Cookie", "session=secret")
		_, _ = w.Write([]byte("<html><body>snapshot</body></html>"))
	}))
	defer server.Close()

	output := filepath.Join(t.TempDir(), "dump")
	result, err := DumpSite(context.Background(), server.URL, SiteDumpOptions{OutputDir: output, Timeout: 2 * time.Second, MaxBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if result.BodyFile != "page.html" || result.BytesSaved == 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(output, "page.html")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(output, "headers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var headers map[string][]string
	if err := json.Unmarshal(data, &headers); err != nil {
		t.Fatal(err)
	}
	if got := headers["Set-Cookie"]; len(got) != 1 || got[0] != "<redacted>" {
		t.Fatalf("cookie was not redacted: %#v", got)
	}
}

func TestDumpSiteTruncates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer server.Close()

	result, err := DumpSite(context.Background(), server.URL, SiteDumpOptions{OutputDir: t.TempDir(), Timeout: 2 * time.Second, MaxBytes: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Truncated || result.BytesSaved != 4 {
		t.Fatalf("unexpected truncation result: %#v", result)
	}
}
