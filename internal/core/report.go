package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func SaveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := ensureParent(path); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func SaveTextReport(path string, report Report) error {
	if err := ensureParent(path); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintln(&b, "PulseNet diagnostic report")
	fmt.Fprintf(&b, "Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "Target: %s\n", report.Target.Input)
	fmt.Fprintf(&b, "Verdict: %s (%d/100)\n\n", report.Verdict, report.Score)
	fmt.Fprintln(&b, "DNS")
	for _, item := range report.DNS {
		if item.Error != "" {
			fmt.Fprintf(&b, "- %s: ERROR: %s\n", item.Resolver, item.Error)
		} else {
			fmt.Fprintf(&b, "- %s: %s (%d ms)\n", item.Resolver, strings.Join(item.Addresses, ", "), item.DurationMS)
		}
	}
	fmt.Fprintln(&b, "\nTCP")
	for _, item := range report.TCP {
		fmt.Fprintf(&b, "- %s/%s: %d/%d successful, avg %.1f ms, jitter %.1f ms\n", item.Port, item.Service, item.Successes, item.Attempts, item.AvgMS, item.JitterMS)
	}
	if report.TLS != nil {
		fmt.Fprintln(&b, "\nTLS")
		if report.TLS.Error != "" {
			fmt.Fprintf(&b, "- ERROR: %s\n", report.TLS.Error)
		} else {
			fmt.Fprintf(&b, "- %s, %s, expires %s (%d days)\n", report.TLS.Protocol, report.TLS.CipherSuite, report.TLS.NotAfter.Format("2006-01-02"), report.TLS.DaysRemaining)
		}
	}
	if report.HTTP != nil {
		fmt.Fprintln(&b, "\nHTTP")
		if report.HTTP.Error != "" {
			fmt.Fprintf(&b, "- ERROR: %s\n", report.HTTP.Error)
		} else {
			fmt.Fprintf(&b, "- %s, %s, TTFB %d ms, total %d ms\n", report.HTTP.Status, report.HTTP.Protocol, report.HTTP.Timings.TTFBMS, report.HTTP.Timings.TotalMS)
			fmt.Fprintf(&b, "- Final URL: %s\n", report.HTTP.FinalURL)
		}
	}
	if report.SecurityHeaders != nil {
		fmt.Fprintf(&b, "\nSecurity headers: %s (%d/100)\n", report.SecurityHeaders.Grade, report.SecurityHeaders.Score)
	}
	if len(report.Recommendations) > 0 {
		fmt.Fprintln(&b, "\nRecommendations")
		for _, recommendation := range report.Recommendations {
			fmt.Fprintf(&b, "- %s\n", recommendation)
		}
	}
	fmt.Fprintf(&b, "\nTotal duration: %d ms\n", report.TotalDurationMS)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func OpenWatchCSV(path string) (*os.File, *csv.Writer, error) {
	if err := ensureParent(path); err != nil {
		return nil, nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"timestamp", "up", "status_code", "latency_ms", "error"}); err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	writer.Flush()
	return file, writer, writer.Error()
}

func WriteWatchCSV(writer *csv.Writer, sample WatchSample) error {
	if writer == nil {
		return nil
	}
	err := writer.Write([]string{
		sample.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		strconv.FormatBool(sample.Up),
		strconv.Itoa(sample.StatusCode),
		strconv.FormatInt(sample.LatencyMS, 10),
		sample.Error,
	})
	writer.Flush()
	if err != nil {
		return err
	}
	return writer.Error()
}

func ensureParent(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}
