package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCombinedLog(t *testing.T) {
	entry := ParseLogLine(`203.0.113.9 - - [05/Aug/2026:12:00:00 +0000] "GET /health HTTP/1.1" 503 21 "-" "curl"`)
	if entry.IP != "203.0.113.9" || entry.Method != "GET" || entry.Path != "/health" || entry.Status != 503 {
		t.Fatalf("unexpected entry: %#v", entry)
	}
	if !MatchLogEntry(entry, LogFilter{Status: "5xx", RequestPath: "health"}) {
		t.Fatal("expected entry to match")
	}
}

func TestParseJSONLog(t *testing.T) {
	entry := ParseLogLine(`{"timestamp":"2026-08-05T12:00:00Z","level":"error","status":500,"path":"/api","message":"upstream failed"}`)
	if entry.Format != "json" || entry.Level != "error" || entry.Status != 500 || entry.Path != "/api" {
		t.Fatalf("unexpected entry: %#v", entry)
	}
}

func TestReadLogEntriesFiltersAndTails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	content := "203.0.113.1 - - [05/Aug/2026:12:00:00 +0000] \"GET / HTTP/1.1\" 200 10\n" +
		"203.0.113.2 - - [05/Aug/2026:12:00:01 +0000] \"GET /api HTTP/1.1\" 500 20\n" +
		"203.0.113.3 - - [05/Aug/2026:12:00:02 +0000] \"POST /api HTTP/1.1\" 502 30\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ReadLogEntries(path, 1, LogFilter{Status: "5xx"})
	if err != nil {
		t.Fatal(err)
	}
	if result.MatchedLines != 2 || len(result.Entries) != 1 || result.Entries[0].Status != 502 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestValidateLogFilter(t *testing.T) {
	if err := ValidateLogFilter(LogFilter{Status: "900"}); err == nil {
		t.Fatal("expected invalid status filter")
	}
}
