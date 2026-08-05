package ui

import (
	"fmt"
	"os"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func PrintStartupIntegrity(status core.IntegrityStatus) {
	switch status.State {
	case "verified":
		fmt.Fprintf(os.Stderr, "%s SHA-256 %s\n", Color(green, "[Integrity] verified"), status.ActualSHA256)
	case "mismatch":
		fmt.Fprintf(os.Stderr, "%s\n", Color(red+bold, "[Integrity] WARNING: executable hash mismatch"))
		fmt.Fprintf(os.Stderr, "  expected: %s\n  actual:   %s\n", status.ExpectedSHA256, status.ActualSHA256)
	case "unmanaged":
		fmt.Fprintf(os.Stderr, "%s SHA-256 %s\n", Color(yellow, "[Integrity] manifest unavailable"), status.ActualSHA256)
	default:
		fmt.Fprintf(os.Stderr, "%s %s\n", Color(red, "[Integrity] check failed:"), status.Error)
	}
}

func PrintIntegritySummary(status core.IntegrityStatus) {
	switch status.State {
	case "verified":
		fmt.Printf("  %s %s\n\n", Color(green, "Integrity verified"), compact(status.ActualSHA256, 20))
	case "mismatch":
		fmt.Printf("  %s\n\n", Color(red+bold, "Integrity mismatch — reinstall from the official release"))
	case "unmanaged":
		fmt.Printf("  %s\n\n", Color(yellow, "Portable/development build — no integrity manifest"))
	default:
		fmt.Printf("  %s\n\n", Color(red, "Integrity check failed"))
	}
}

func PrintIntegrityStatus(status core.IntegrityStatus, jsonOutput bool) {
	if jsonOutput {
		PrintJSON(status)
		return
	}

	fmt.Println("\n" + Color(bold, "Executable integrity"))
	fmt.Printf("  State:         %s\n", status.State)
	fmt.Printf("  Executable:    %s\n", status.ExecutablePath)
	fmt.Printf("  Manifest:      %s\n", status.ManifestPath)
	if status.ActualSHA256 != "" {
		fmt.Printf("  Actual SHA-256:   %s\n", status.ActualSHA256)
	}
	if status.ExpectedSHA256 != "" {
		fmt.Printf("  Expected SHA-256: %s\n", status.ExpectedSHA256)
	}
	if status.Manifest != nil {
		fmt.Printf("  Release tag:   %s\n", emptyAs(status.Manifest.ReleaseTag, "unknown"))
		fmt.Printf("  Version:       %s\n", emptyAs(status.Manifest.Version, "unknown"))
	}
	if status.Error != "" {
		fmt.Printf("  Note:          %s\n", status.Error)
	}
}
