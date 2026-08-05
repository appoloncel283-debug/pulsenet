package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const IntegrityManifestName = "integrity.json"

type IntegrityManifest struct {
	Product     string    `json:"product"`
	Version     string    `json:"version"`
	ReleaseTag  string    `json:"release_tag,omitempty"`
	Executable  string    `json:"executable"`
	SHA256      string    `json:"sha256"`
	GeneratedAt time.Time `json:"generated_at,omitempty"`
}

type IntegrityStatus struct {
	State          string             `json:"state"`
	ExecutablePath string             `json:"executable_path"`
	ManifestPath   string             `json:"manifest_path"`
	ActualSHA256   string             `json:"actual_sha256,omitempty"`
	ExpectedSHA256 string             `json:"expected_sha256,omitempty"`
	Manifest       *IntegrityManifest `json:"manifest,omitempty"`
	Error          string             `json:"error,omitempty"`
}

func VerifySelfIntegrity() IntegrityStatus {
	executable, err := os.Executable()
	if err != nil {
		return IntegrityStatus{State: "error", Error: err.Error()}
	}
	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}
	manifestPath := filepath.Join(filepath.Dir(executable), IntegrityManifestName)
	return VerifyExecutableIntegrity(executable, manifestPath)
}

func VerifyExecutableIntegrity(executablePath, manifestPath string) IntegrityStatus {
	status := IntegrityStatus{
		State:          "error",
		ExecutablePath: executablePath,
		ManifestPath:   manifestPath,
	}

	actualHash, err := FileSHA256(executablePath)
	if err != nil {
		status.Error = fmt.Sprintf("calculate executable SHA-256: %v", err)
		return status
	}
	status.ActualSHA256 = strings.ToUpper(actualHash)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			status.State = "unmanaged"
			status.Error = "integrity manifest was not found; this may be a portable or development build"
			return status
		}
		status.Error = fmt.Sprintf("read integrity manifest: %v", err)
		return status
	}

	var manifest IntegrityManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		status.Error = fmt.Sprintf("parse integrity manifest: %v", err)
		return status
	}
	status.Manifest = &manifest
	status.ExpectedSHA256 = strings.ToUpper(strings.TrimSpace(manifest.SHA256))
	if len(status.ExpectedSHA256) != 64 {
		status.Error = "integrity manifest contains an invalid SHA-256 value"
		return status
	}
	if manifest.Executable != "" && !strings.EqualFold(filepath.Base(executablePath), filepath.Base(manifest.Executable)) {
		status.Error = fmt.Sprintf("integrity manifest expects %s, running %s", manifest.Executable, filepath.Base(executablePath))
		return status
	}

	if status.ActualSHA256 != status.ExpectedSHA256 {
		status.State = "mismatch"
		status.Error = "the executable SHA-256 does not match the installed manifest"
		return status
	}

	status.State = "verified"
	status.Error = ""
	return status
}
