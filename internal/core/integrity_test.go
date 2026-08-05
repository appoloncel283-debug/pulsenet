package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyExecutableIntegrity(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "PulseNet.exe")
	manifestPath := filepath.Join(directory, IntegrityManifestName)
	if err := os.WriteFile(executable, []byte("pulsenet-test-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	hash, err := FileSHA256(executable)
	if err != nil {
		t.Fatal(err)
	}
	manifest := IntegrityManifest{
		Product:    "PulseNet",
		Version:    "test",
		Executable: "PulseNet.exe",
		SHA256:     hash,
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	status := VerifyExecutableIntegrity(executable, manifestPath)
	if status.State != "verified" {
		t.Fatalf("state = %q, error = %q", status.State, status.Error)
	}

	if err := os.WriteFile(executable, []byte("modified"), 0o755); err != nil {
		t.Fatal(err)
	}
	status = VerifyExecutableIntegrity(executable, manifestPath)
	if status.State != "mismatch" {
		t.Fatalf("state = %q, want mismatch", status.State)
	}
}

func TestVerifyExecutableIntegrityWithoutManifest(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "pulsenet")
	if err := os.WriteFile(executable, []byte("portable"), 0o755); err != nil {
		t.Fatal(err)
	}
	status := VerifyExecutableIntegrity(executable, filepath.Join(directory, IntegrityManifestName))
	if status.State != "unmanaged" {
		t.Fatalf("state = %q, want unmanaged", status.State)
	}
}
