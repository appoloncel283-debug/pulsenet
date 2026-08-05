package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SiteDumpDiff struct {
	OldDirectory   string            `json:"old_directory"`
	NewDirectory   string            `json:"new_directory"`
	BodyChanged    bool              `json:"body_changed"`
	StatusChanged  bool              `json:"status_changed"`
	HeadersChanged bool              `json:"headers_changed"`
	OldSHA256      string            `json:"old_sha256,omitempty"`
	NewSHA256      string            `json:"new_sha256,omitempty"`
	OldStatus      string            `json:"old_status,omitempty"`
	NewStatus      string            `json:"new_status,omitempty"`
	AddedHeaders   map[string]string `json:"added_headers,omitempty"`
	RemovedHeaders map[string]string `json:"removed_headers,omitempty"`
	ChangedHeaders map[string]string `json:"changed_headers,omitempty"`
}

type LocalSecretFinding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Kind    string `json:"kind"`
	Preview string `json:"preview"`
}

func CompareSiteDumps(oldDirectory, newDirectory string) (SiteDumpDiff, error) {
	oldMeta, oldHeaders, err := readDumpMetadata(oldDirectory)
	if err != nil {
		return SiteDumpDiff{}, fmt.Errorf("old dump: %w", err)
	}
	newMeta, newHeaders, err := readDumpMetadata(newDirectory)
	if err != nil {
		return SiteDumpDiff{}, fmt.Errorf("new dump: %w", err)
	}

	diff := SiteDumpDiff{
		OldDirectory:   oldDirectory,
		NewDirectory:   newDirectory,
		OldSHA256:      oldMeta.SHA256,
		NewSHA256:      newMeta.SHA256,
		OldStatus:      oldMeta.Status,
		NewStatus:      newMeta.Status,
		BodyChanged:    oldMeta.SHA256 != newMeta.SHA256,
		StatusChanged:  oldMeta.Status != newMeta.Status,
		AddedHeaders:   map[string]string{},
		RemovedHeaders: map[string]string{},
		ChangedHeaders: map[string]string{},
	}
	for key, oldValue := range oldHeaders {
		newValue, ok := newHeaders[key]
		if !ok {
			diff.RemovedHeaders[key] = oldValue
			continue
		}
		if oldValue != newValue {
			diff.ChangedHeaders[key] = oldValue + " -> " + newValue
		}
	}
	for key, newValue := range newHeaders {
		if _, ok := oldHeaders[key]; !ok {
			diff.AddedHeaders[key] = newValue
		}
	}
	diff.HeadersChanged = len(diff.AddedHeaders)+len(diff.RemovedHeaders)+len(diff.ChangedHeaders) > 0
	if len(diff.AddedHeaders) == 0 {
		diff.AddedHeaders = nil
	}
	if len(diff.RemovedHeaders) == 0 {
		diff.RemovedHeaders = nil
	}
	if len(diff.ChangedHeaders) == 0 {
		diff.ChangedHeaders = nil
	}
	return diff, nil
}

func readDumpMetadata(directory string) (SiteDumpResult, map[string]string, error) {
	metadataBytes, err := os.ReadFile(filepath.Join(directory, "metadata.json"))
	if err != nil {
		return SiteDumpResult{}, nil, err
	}
	var metadata SiteDumpResult
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return SiteDumpResult{}, nil, err
	}

	headersBytes, err := os.ReadFile(filepath.Join(directory, "headers.json"))
	if err != nil {
		return SiteDumpResult{}, nil, err
	}
	var raw map[string][]string
	if err := json.Unmarshal(headersBytes, &raw); err != nil {
		return SiteDumpResult{}, nil, err
	}
	headers := make(map[string]string, len(raw))
	for key, values := range raw {
		headers[strings.ToLower(key)] = strings.Join(values, " | ")
	}
	return metadata, headers, nil
}

func ScanLocalDumpForSecrets(directory string, maxBytes int64) ([]LocalSecretFinding, error) {
	if maxBytes <= 0 {
		maxBytes = 32 * 1024 * 1024
	}
	patterns := []struct {
		kind    string
		needles []string
	}{
		{"private-key", []string{"-----BEGIN PRIVATE KEY-----", "-----BEGIN RSA PRIVATE KEY-----", "-----BEGIN OPENSSH PRIVATE KEY-----"}},
		{"aws-access-key", []string{"AKIA", "ASIA"}},
		{"github-token", []string{"ghp_", "github_pat_"}},
		{"generic-secret", []string{"api_key", "apikey", "secret_key", "client_secret", "access_token", "authorization: bearer"}},
	}

	files := make([]string, 0)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if name == "headers.json" || name == "metadata.json" || strings.HasPrefix(name, "body.") || name == "page.html" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	findings := make([]LocalSecretFinding, 0)
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.Size() > maxBytes {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			lower := strings.ToLower(line)
			for _, pattern := range patterns {
				matched := false
				for _, needle := range pattern.needles {
					if pattern.kind == "private-key" || pattern.kind == "aws-access-key" || pattern.kind == "github-token" {
						matched = strings.Contains(line, needle)
					} else {
						matched = strings.Contains(lower, needle)
					}
					if matched {
						break
					}
				}
				if matched {
					findings = append(findings, LocalSecretFinding{
						File:    path,
						Line:    lineNumber,
						Kind:    pattern.kind,
						Preview: redactPreview(line),
					})
				}
			}
		}
		scanErr := scanner.Err()
		closeErr := file.Close()
		if scanErr != nil {
			return nil, scanErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	return findings, nil
}

func redactPreview(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 120 {
		value = value[:120] + "…"
	}
	if len(value) <= 16 {
		return "<redacted>"
	}
	return value[:8] + "…<redacted>…" + value[len(value)-4:]
}
