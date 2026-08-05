package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareSiteDumps(t *testing.T) {
	base := t.TempDir()
	oldDir := filepath.Join(base, "old")
	newDir := filepath.Join(base, "new")
	for _, dir := range []string{oldDir, newDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(dir, hash, status string, headers map[string][]string) {
		meta, _ := json.Marshal(SiteDumpResult{SHA256: hash, Status: status})
		headerBytes, _ := json.Marshal(headers)
		if err := os.WriteFile(filepath.Join(dir, "metadata.json"), meta, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "headers.json"), headerBytes, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(oldDir, "a", "200 OK", map[string][]string{"Server": {"old"}})
	write(newDir, "b", "200 OK", map[string][]string{"Server": {"new"}})

	diff, err := CompareSiteDumps(oldDir, newDir)
	if err != nil {
		t.Fatal(err)
	}
	if !diff.BodyChanged || !diff.HeadersChanged || diff.StatusChanged {
		t.Fatalf("unexpected diff: %#v", diff)
	}
}
