package ui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunSiteDump(input, output string, maxBytes int64, timeout time.Duration, insecure, jsonOutput bool) error {
	result, err := core.DumpSite(context.Background(), input, core.SiteDumpOptions{
		OutputDir:   output,
		Timeout:     timeout,
		MaxBytes:    maxBytes,
		InsecureTLS: insecure,
	})
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(result)
		return nil
	}
	absolute, _ := filepath.Abs(result.OutputDir)
	fmt.Printf("\n%s %s\n", Color(bold, "Site dump saved:"), absolute)
	fmt.Printf("  Final URL:     %s\n", result.FinalURL)
	fmt.Printf("  Response:      %s · %s\n", result.Status, result.Protocol)
	fmt.Printf("  Content type:  %s\n", emptyAs(result.ContentType, "unknown"))
	fmt.Printf("  Body:          %s (%d bytes)\n", result.BodyFile, result.BytesSaved)
	fmt.Printf("  Headers:       %s\n", result.HeadersFile)
	fmt.Printf("  Metadata:      %s\n", result.MetadataFile)
	fmt.Printf("  SHA-256:       %s\n", result.SHA256)
	if len(result.Redirects) > 0 {
		fmt.Printf("  Redirects:     %d\n", len(result.Redirects))
	}
	if result.Truncated {
		fmt.Println(Color(yellow, "  Warning: the response body reached the configured size limit and was truncated."))
	}
	fmt.Println(Color(dim, "  Sensitive response headers such as Set-Cookie are redacted in the saved header file."))
	return nil
}

func RunLogs(path string, lines int, follow bool, interval time.Duration, filter core.LogFilter, jsonOutput bool) error {
	result, err := core.ReadLogEntries(path, lines, filter)
	if err != nil {
		return err
	}
	if jsonOutput {
		PrintJSON(result)
		return nil
	}
	absolute, _ := filepath.Abs(path)
	fmt.Printf("\n%s %s\n", Color(bold, "Website log:"), absolute)
	fmt.Printf("  Scanned: %d lines · matched: %d · showing: %d\n", result.ScannedLines, result.MatchedLines, len(result.Entries))
	if result.TruncatedScan {
		fmt.Println(Color(yellow, "  Note: only the most recent 64 MiB of this large log file was scanned."))
	}
	fmt.Println()
	for _, entry := range result.Entries {
		printLogEntry(entry)
	}
	if !follow {
		printLogSummary(result)
		return nil
	}

	fmt.Println(Color(dim, "\nFollowing new lines. Press Ctrl+C to stop."))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return core.FollowLog(ctx, path, interval, filter, printLogEntry)
}

func printLogEntry(entry core.LogEntry) {
	code := ""
	color := reset
	switch {
	case entry.Level == "fatal" || entry.Level == "panic" || entry.Level == "crit":
		code, color = strings.ToUpper(entry.Level), red+bold
	case entry.Level == "error":
		code, color = "ERROR", red
	case entry.Level == "warn":
		code, color = "WARN", yellow
	case entry.Status >= 500:
		code, color = fmt.Sprintf("HTTP %d", entry.Status), red
	case entry.Status >= 400:
		code, color = fmt.Sprintf("HTTP %d", entry.Status), yellow
	case entry.Status > 0:
		code, color = fmt.Sprintf("HTTP %d", entry.Status), green
	case entry.Level != "":
		code, color = strings.ToUpper(entry.Level), cyan
	}
	if code == "" {
		fmt.Println(entry.Raw)
		return
	}
	fmt.Printf("%s %s\n", Color(color, fmt.Sprintf("%-8s", code)), entry.Raw)
}

func printLogSummary(result core.LogViewResult) {
	if len(result.StatusCounts) == 0 && len(result.LevelCounts) == 0 {
		return
	}
	fmt.Println("\n" + Color(bold, "Matched summary"))
	if len(result.LevelCounts) > 0 {
		fmt.Print("  Levels: ")
		first := true
		for level, count := range result.LevelCounts {
			if !first {
				fmt.Print(" · ")
			}
			fmt.Printf("%s %d", level, count)
			first = false
		}
		fmt.Println()
	}
	if len(result.StatusCounts) > 0 {
		classes := map[int]int{}
		for status, count := range result.StatusCounts {
			classes[status/100] += count
		}
		fmt.Print("  HTTP:   ")
		first := true
		for class := 1; class <= 5; class++ {
			if classes[class] == 0 {
				continue
			}
			if !first {
				fmt.Print(" · ")
			}
			fmt.Printf("%dxx %d", class, classes[class])
			first = false
		}
		fmt.Println()
	}
}
