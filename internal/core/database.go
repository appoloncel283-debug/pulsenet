package core

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Database commands intentionally wrap official database client tools. They do
// not bypass authentication and never discover credentials on their own.

type DBEngine string

const (
	DBPostgres DBEngine = "postgres"
	DBMySQL    DBEngine = "mysql"
	DBSQLite   DBEngine = "sqlite"
)

type DBToolStatus struct {
	Engine    DBEngine `json:"engine"`
	Available bool     `json:"available"`
	Binary    string   `json:"binary,omitempty"`
	Version   string   `json:"version,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type DBOperationResult struct {
	Engine      DBEngine  `json:"engine"`
	Operation   string    `json:"operation"`
	OutputPath  string    `json:"output_path,omitempty"`
	SHA256      string    `json:"sha256,omitempty"`
	Bytes       int64     `json:"bytes,omitempty"`
	Verified    bool      `json:"verified"`
	StartedAt   time.Time `json:"started_at"`
	DurationMS  int64     `json:"duration_ms"`
	Tool        string    `json:"tool,omitempty"`
	ToolVersion string    `json:"tool_version,omitempty"`
	Details     []string  `json:"details,omitempty"`
}

func ParseDBEngine(value string) (DBEngine, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "postgres", "postgresql", "pg":
		return DBPostgres, nil
	case "mysql", "mariadb":
		return DBMySQL, nil
	case "sqlite", "sqlite3":
		return DBSQLite, nil
	default:
		return "", fmt.Errorf("unsupported database engine %q; use postgres, mysql, or sqlite", value)
	}
}

func CheckDatabaseTools(ctx context.Context) []DBToolStatus {
	checks := []struct {
		engine DBEngine
		binary string
	}{
		{DBPostgres, "pg_dump"},
		{DBMySQL, "mysqldump"},
		{DBSQLite, "sqlite3"},
	}

	results := make([]DBToolStatus, 0, len(checks))
	for _, check := range checks {
		status := DBToolStatus{Engine: check.engine, Binary: check.binary}
		path, err := exec.LookPath(check.binary)
		if err != nil {
			status.Error = err.Error()
			results = append(results, status)
			continue
		}
		status.Available = true
		status.Binary = path
		commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		output, commandErr := exec.CommandContext(commandCtx, path, "--version").CombinedOutput()
		cancel()
		status.Version = strings.TrimSpace(string(output))
		if commandErr != nil {
			status.Error = commandErr.Error()
		}
		results = append(results, status)
	}
	return results
}

func DatabaseSchema(ctx context.Context, engine DBEngine, database, output string, extraArgs []string, timeout time.Duration) (DBOperationResult, error) {
	return runDatabaseExport(ctx, engine, "schema", database, output, extraArgs, timeout)
}

func DatabaseBackup(ctx context.Context, engine DBEngine, database, output string, extraArgs []string, timeout time.Duration) (DBOperationResult, error) {
	return runDatabaseExport(ctx, engine, "backup", database, output, extraArgs, timeout)
}

func runDatabaseExport(parent context.Context, engine DBEngine, operation, database, output string, extraArgs []string, timeout time.Duration) (DBOperationResult, error) {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	if strings.TrimSpace(database) == "" {
		return DBOperationResult{}, fmt.Errorf("database is required")
	}
	if strings.TrimSpace(output) == "" {
		return DBOperationResult{}, fmt.Errorf("output path is required")
	}
	parentDir := filepath.Dir(output)
	if parentDir != "." {
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			return DBOperationResult{}, err
		}
	}

	binary, args, err := databaseExportCommand(engine, operation, database, output, extraArgs)
	if err != nil {
		return DBOperationResult{}, err
	}
	path, err := exec.LookPath(binary)
	if err != nil {
		return DBOperationResult{}, fmt.Errorf("%s is not installed or not available in PATH", binary)
	}

	result := DBOperationResult{
		Engine:     engine,
		Operation:  operation,
		OutputPath: output,
		StartedAt:  time.Now(),
		Tool:       path,
	}
	versionCtx, versionCancel := context.WithTimeout(parent, 5*time.Second)
	versionOutput, _ := exec.CommandContext(versionCtx, path, "--version").CombinedOutput()
	versionCancel()
	result.ToolVersion = strings.TrimSpace(string(versionOutput))

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)

	var outputFile *os.File
	if engine == DBSQLite && operation == "schema" {
		outputFile, err = os.Create(output)
		if err != nil {
			return result, err
		}
		defer outputFile.Close()
		cmd.Stdout = outputFile
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return result, err
	}
	if err := cmd.Start(); err != nil {
		return result, err
	}
	stderrBytes, _ := io.ReadAll(io.LimitReader(stderr, 1024*1024))
	err = cmd.Wait()
	result.DurationMS = time.Since(result.StartedAt).Milliseconds()
	if ctx.Err() != nil {
		_ = os.Remove(output)
		return result, fmt.Errorf("database %s timed out: %w", operation, ctx.Err())
	}
	if err != nil {
		_ = os.Remove(output)
		message := strings.TrimSpace(string(stderrBytes))
		if message == "" {
			message = err.Error()
		}
		return result, fmt.Errorf("database %s failed: %s", operation, message)
	}
	if outputFile != nil {
		if err := outputFile.Sync(); err != nil {
			return result, err
		}
	}

	info, err := os.Stat(output)
	if err != nil {
		return result, err
	}
	if info.Size() == 0 {
		return result, fmt.Errorf("database %s produced an empty file", operation)
	}
	result.Bytes = info.Size()
	result.SHA256, err = FileSHA256(output)
	if err != nil {
		return result, err
	}
	result.Verified = true
	result.Details = append(result.Details, "Authentication is handled only by the official database client and user-supplied environment variables or client configuration.")
	return result, nil
}

func databaseExportCommand(engine DBEngine, operation, database, output string, extraArgs []string) (string, []string, error) {
	switch engine {
	case DBPostgres:
		args := []string{"--file", output, "--no-owner", "--no-privileges"}
		if operation == "schema" {
			args = append(args, "--schema-only", "--format=plain")
		} else {
			args = append(args, "--format=custom")
		}
		args = append(args, extraArgs...)
		args = append(args, database)
		return "pg_dump", args, nil
	case DBMySQL:
		args := []string{"--single-transaction", "--quick", "--routines", "--events", "--triggers", "--result-file=" + output}
		if operation == "schema" {
			args = append(args, "--no-data")
		}
		args = append(args, extraArgs...)
		args = append(args, database)
		return "mysqldump", args, nil
	case DBSQLite:
		if _, err := os.Stat(database); err != nil {
			return "", nil, fmt.Errorf("sqlite database: %w", err)
		}
		if operation == "schema" {
			args := append([]string{database, ".schema"}, extraArgs...)
			return "sqlite3", args, nil
		}
		quotedOutput := strings.ReplaceAll(filepath.Clean(output), "'", "''")
		args := append([]string{database, ".backup '" + quotedOutput + "'"}, extraArgs...)
		return "sqlite3", args, nil
	default:
		return "", nil, fmt.Errorf("unsupported database engine %q", engine)
	}
}

func VerifyDatabaseArtifact(ctx context.Context, engine DBEngine, path string, timeout time.Duration) (DBOperationResult, error) {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	info, err := os.Stat(path)
	if err != nil {
		return DBOperationResult{}, err
	}
	if info.IsDir() {
		return DBOperationResult{}, fmt.Errorf("artifact path is a directory")
	}
	result := DBOperationResult{
		Engine:     engine,
		Operation:  "verify",
		OutputPath: path,
		StartedAt:  time.Now(),
		Bytes:      info.Size(),
	}
	result.SHA256, err = FileSHA256(path)
	if err != nil {
		return result, err
	}

	verifyCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	switch engine {
	case DBPostgres:
		binary, lookupErr := exec.LookPath("pg_restore")
		if lookupErr != nil {
			return result, fmt.Errorf("pg_restore is required to verify PostgreSQL custom archives")
		}
		result.Tool = binary
		output, commandErr := exec.CommandContext(verifyCtx, binary, "--list", path).CombinedOutput()
		if commandErr != nil {
			return result, fmt.Errorf("PostgreSQL archive verification failed: %s", strings.TrimSpace(string(output)))
		}
		result.Details = append(result.Details, fmt.Sprintf("Archive catalog contains %d lines.", countLines(output)))
	case DBSQLite:
		binary, lookupErr := exec.LookPath("sqlite3")
		if lookupErr != nil {
			return result, fmt.Errorf("sqlite3 is required to verify SQLite files")
		}
		result.Tool = binary
		output, commandErr := exec.CommandContext(verifyCtx, binary, path, "PRAGMA quick_check;").CombinedOutput()
		if commandErr != nil {
			return result, fmt.Errorf("SQLite integrity check failed: %s", strings.TrimSpace(string(output)))
		}
		if strings.TrimSpace(string(output)) != "ok" {
			return result, fmt.Errorf("SQLite integrity check returned: %s", strings.TrimSpace(string(output)))
		}
		result.Details = append(result.Details, "PRAGMA quick_check returned ok.")
	case DBMySQL:
		file, openErr := os.Open(path)
		if openErr != nil {
			return result, openErr
		}
		defer file.Close()
		scanner := bufio.NewScanner(io.LimitReader(file, 8*1024*1024))
		seenSQL := false
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "CREATE ") || strings.HasPrefix(line, "INSERT ") || strings.HasPrefix(line, "-- MySQL dump") {
				seenSQL = true
				break
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			return result, scanErr
		}
		if !seenSQL {
			return result, fmt.Errorf("file does not look like a MySQL dump")
		}
		result.Details = append(result.Details, "The file contains recognizable MySQL dump statements.")
	default:
		return result, fmt.Errorf("unsupported database engine %q", engine)
	}
	if verifyCtx.Err() != nil {
		return result, fmt.Errorf("verification timed out: %w", verifyCtx.Err())
	}
	result.DurationMS = time.Since(result.StartedAt).Milliseconds()
	result.Verified = true
	return result, nil
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func WriteDatabaseManifest(path string, result DBOperationResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, value := range data {
		if value == '\n' {
			count++
		}
	}
	return count
}
