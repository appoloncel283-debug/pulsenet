package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunDiagnose(input string, timeout time.Duration, attempts int, ports []string, insecure bool, jsonPath, textPath string) error {
	target, err := core.ParseTarget(input)
	if err != nil {
		return err
	}
	fmt.Printf("\n%s %s\n", Color(cyan, "Diagnosing:"), target.Input)
	fmt.Println(Color(dim, "DNS → TCP → TLS → HTTP → security headers → recommendations"))
	report := core.RunDiagnosis(context.Background(), target, core.Options{Timeout: timeout, TCPAttempts: attempts, Ports: ports, InsecureTLS: insecure})
	PrintReport(report)
	if jsonPath != "" {
		if err := core.SaveJSON(jsonPath, report); err != nil {
			return err
		}
		fmt.Println(Color(green, "JSON report: "+jsonPath))
	}
	if textPath != "" {
		if err := core.SaveTextReport(textPath, report); err != nil {
			return err
		}
		fmt.Println(Color(green, "Text report: "+textPath))
	}
	return nil
}

func PrintReport(report core.Report) {
	fmt.Println("\n" + Color(bold, "DNS"))
	for _, item := range report.DNS {
		if item.Error != "" {
			fmt.Printf("  %s %-23s %s\n", Color(red, "✗"), item.Resolver, compact(item.Error, 120))
		} else {
			fmt.Printf("  %s %-23s %-45s %d ms\n", Color(green, "✓"), item.Resolver, strings.Join(item.Addresses, ", "), item.DurationMS)
		}
	}

	fmt.Println("\n" + Color(bold, "TCP"))
	for _, item := range report.TCP {
		service := item.Service
		if service == "" {
			service = "unknown"
		}
		if item.Successes == 0 {
			fmt.Printf("  %s %-5s %-18s unavailable  %s\n", Color(red, "✗"), item.Port, service, compact(item.LastError, 90))
		} else {
			status := green
			if item.Successes < item.Attempts {
				status = yellow
			}
			fmt.Printf("  %s %-5s %-18s %d/%d  min %s  avg %s  max %s  jitter %s\n", Color(status, "✓"), item.Port, service, item.Successes, item.Attempts, core.FormatMS(item.MinMS), core.FormatMS(item.AvgMS), core.FormatMS(item.MaxMS), core.FormatMS(item.JitterMS))
		}
	}

	fmt.Println("\n" + Color(bold, "TLS"))
	if report.TLS == nil {
		fmt.Println("  " + Color(dim, "– not applicable"))
	} else if report.TLS.Error != "" {
		fmt.Println("  " + Color(red, "✗ "+compact(report.TLS.Error, 130)))
	} else {
		status := green
		if report.TLS.DaysRemaining <= 14 {
			status = yellow
		}
		fmt.Printf("  %s %s · %s · ALPN %s\n", Color(status, "✓"), report.TLS.Protocol, report.TLS.CipherSuite, emptyAs(report.TLS.ALPN, "none"))
		fmt.Printf("    expires %s · %d days remaining · chain %d · OCSP stapled %v\n", report.TLS.NotAfter.Format("2006-01-02"), report.TLS.DaysRemaining, report.TLS.ChainLength, report.TLS.OCSPStapled)
		fmt.Printf("    issuer: %s\n", compact(report.TLS.Issuer, 110))
	}

	fmt.Println("\n" + Color(bold, "HTTP"))
	if report.HTTP == nil || report.HTTP.Error != "" {
		message := "no result"
		if report.HTTP != nil {
			message = report.HTTP.Error
		}
		fmt.Println("  " + Color(red, "✗ "+compact(message, 130)))
	} else {
		status := green
		if report.HTTP.StatusCode >= 400 {
			status = yellow
		}
		if report.HTTP.StatusCode >= 500 {
			status = red
		}
		fmt.Printf("  %s %s · %s · total %d ms · TTFB %d ms\n", Color(status, "✓"), report.HTTP.Status, report.HTTP.Protocol, report.HTTP.Timings.TotalMS, report.HTTP.Timings.TTFBMS)
		fmt.Printf("    %s\n", report.HTTP.FinalURL)
		if report.HTTP.RemoteAddress != "" {
			fmt.Printf("    remote: %s\n", report.HTTP.RemoteAddress)
		}
	}
	if report.SecurityHeaders != nil {
		fmt.Printf("\n%s grade %s · %d/100\n", Color(bold, "Security headers:"), report.SecurityHeaders.Grade, report.SecurityHeaders.Score)
	}

	verdictColor := green
	if report.Score < 90 {
		verdictColor = yellow
	}
	if report.Score < 45 {
		verdictColor = red
	}
	fmt.Printf("\n%s %s  %s\n", Color(bold, "RESULT:"), Color(verdictColor, report.Verdict), Color(bold, fmt.Sprintf("%d/100", report.Score)))
	for _, recommendation := range report.Recommendations {
		fmt.Printf("  %s %s\n", Color(yellow, "→"), recommendation)
	}
	fmt.Printf("\n%s\n", Color(dim, fmt.Sprintf("Completed in %d ms", report.TotalDurationMS)))
}
