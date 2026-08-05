package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunDBTools(jsonOutput bool) {
	results := core.CheckDatabaseTools(context.Background())
	if jsonOutput {
		PrintJSON(results)
		return
	}
	fmt.Println("\n" + Color(bold, "Database client tools"))
	for _, result := range results {
		if result.Available {
			fmt.Printf("  %s %-10s %s\n", Color(green, "✓"), result.Engine, result.Version)
		} else {
			fmt.Printf("  %s %-10s not installed\n", Color(yellow, "–"), result.Engine)
		}
	}
}

func RunDBExport(engineValue, operation, database, output string, extraArgs []string, timeout time.Duration, jsonOutput bool) error {
	engine, err := core.ParseDBEngine(engineValue)
	if err != nil {
		return err
	}
	var result core.DBOperationResult
	if operation == "schema" {
		result, err = core.DatabaseSchema(context.Background(), engine, database, output, extraArgs, timeout)
	} else {
		result, err = core.DatabaseBackup(context.Background(), engine, database, output, extraArgs, timeout)
	}
	if err != nil {
		return err
	}
	manifest := output + ".manifest.json"
	if err := core.WriteDatabaseManifest(manifest, result); err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(result)
		return nil
	}
	label := "Backup created:"
	if operation == "schema" {
		label = "Schema dump created:"
	}
	fmt.Printf("\n%s %s\n", Color(bold, label), result.OutputPath)
	fmt.Printf("  Engine:       %s\n", result.Engine)
	fmt.Printf("  Size:         %d bytes\n", result.Bytes)
	fmt.Printf("  SHA-256:      %s\n", result.SHA256)
	fmt.Printf("  Tool:         %s\n", result.ToolVersion)
	fmt.Printf("  Manifest:     %s\n", manifest)
	fmt.Printf("  Duration:     %d ms\n", result.DurationMS)
	return nil
}

func RunDBVerify(engineValue, path string, timeout time.Duration, jsonOutput bool) error {
	engine, err := core.ParseDBEngine(engineValue)
	if err != nil {
		return err
	}
	result, err := core.VerifyDatabaseArtifact(context.Background(), engine, path, timeout)
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(result)
		return nil
	}
	fmt.Printf("\n%s %s\n", Color(green+bold, "Database artifact verified:"), path)
	fmt.Printf("  Engine:       %s\n", result.Engine)
	fmt.Printf("  Size:         %d bytes\n", result.Bytes)
	fmt.Printf("  SHA-256:      %s\n", result.SHA256)
	for _, detail := range result.Details {
		fmt.Printf("  %s\n", detail)
	}
	return nil
}

func RunSiteDiff(oldDirectory, newDirectory string, jsonOutput bool) error {
	result, err := core.CompareSiteDumps(oldDirectory, newDirectory)
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(result)
		return nil
	}
	fmt.Println("\n" + Color(bold, "Site dump comparison"))
	fmt.Printf("  Body changed:     %v\n", result.BodyChanged)
	fmt.Printf("  Status changed:   %v (%s -> %s)\n", result.StatusChanged, result.OldStatus, result.NewStatus)
	fmt.Printf("  Headers changed:  %v\n", result.HeadersChanged)
	fmt.Printf("  Old SHA-256:      %s\n", result.OldSHA256)
	fmt.Printf("  New SHA-256:      %s\n", result.NewSHA256)
	for key, value := range result.AddedHeaders {
		fmt.Printf("  + %s: %s\n", key, value)
	}
	for key, value := range result.RemovedHeaders {
		fmt.Printf("  - %s: %s\n", key, value)
	}
	for key, value := range result.ChangedHeaders {
		fmt.Printf("  ~ %s: %s\n", key, value)
	}
	return nil
}

func RunSiteSecretScan(directory string, maxMB int64, jsonOutput bool) error {
	findings, err := core.ScanLocalDumpForSecrets(directory, maxMB*1024*1024)
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(findings)
		return nil
	}
	fmt.Printf("\n%s %s\n", Color(bold, "Local dump secret scan:"), directory)
	if len(findings) == 0 {
		fmt.Println(Color(green, "  No obvious secret patterns were found."))
		return nil
	}
	for _, finding := range findings {
		fmt.Printf("  %s %s:%d  %s\n", Color(yellow, strings.ToUpper(finding.Kind)), finding.File, finding.Line, finding.Preview)
	}
	fmt.Printf("\n  %d potential finding(s). Review manually before publishing the dump.\n", len(findings))
	return nil
}
