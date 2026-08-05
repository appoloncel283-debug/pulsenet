package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
)

func RunDNS(input string, timeout time.Duration, jsonOutput bool) {
	target, err := core.ParseTarget(input)
	if err != nil {
		PrintError(err)
		return
	}
	comparison := core.ProbeDNS(context.Background(), target.Host, timeout)
	records := core.LookupDNSRecords(context.Background(), target.Host, timeout)
	if jsonOutput {
		PrintJSON(map[string]any{"comparison": comparison, "records": records})
		return
	}
	fmt.Printf("\n%s %s\n", Color(bold, "Resolver comparison:"), target.Host)
	for _, item := range comparison {
		if item.Error != "" {
			fmt.Printf("  %s %-23s %s\n", Color(red, "✗"), item.Resolver, compact(item.Error, 100))
		} else {
			fmt.Printf("  %s %-23s %-45s %d ms\n", Color(green, "✓"), item.Resolver, strings.Join(item.Addresses, ", "), item.DurationMS)
		}
	}
	fmt.Println("\n" + Color(bold, "DNS records"))
	printRecord("A", records.A)
	printRecord("AAAA", records.AAAA)
	if records.CNAME != "" {
		printRecord("CNAME", []string{records.CNAME})
	}
	printRecord("MX", records.MX)
	printRecord("NS", records.NS)
	printRecord("TXT", records.TXT)
	printRecord("PTR", records.PTR)
	for _, message := range records.Errors {
		fmt.Println("  " + Color(yellow, "! "+message))
	}
}

func RunTLS(input string, timeout time.Duration, insecure bool) {
	target, err := core.ParseTarget(input)
	if err != nil {
		PrintError(err)
		return
	}
	port := target.ExplicitPort
	if port == "" {
		port = "443"
	}
	result := core.ProbeTLS(context.Background(), target.Host, port, timeout, insecure)
	if result.Error != "" {
		PrintError(fmt.Errorf("TLS: %s", result.Error))
		return
	}
	fmt.Printf("\n%s %s\n", Color(bold, "TLS certificate:"), result.Address)
	fmt.Printf("  Protocol:       %s\n", result.Protocol)
	fmt.Printf("  Cipher suite:   %s\n", result.CipherSuite)
	fmt.Printf("  ALPN:           %s\n", emptyAs(result.ALPN, "none"))
	fmt.Printf("  Subject:        %s\n", result.Subject)
	fmt.Printf("  Issuer:         %s\n", result.Issuer)
	fmt.Printf("  Serial:         %s\n", result.SerialNumber)
	fmt.Printf("  Valid from:     %s\n", result.NotBefore.Format(time.RFC1123))
	fmt.Printf("  Valid until:    %s (%d days)\n", result.NotAfter.Format(time.RFC1123), result.DaysRemaining)
	fmt.Printf("  Verified:       %v\n", result.Verified)
	fmt.Printf("  Chain length:   %d\n", result.ChainLength)
	fmt.Printf("  OCSP stapled:   %v\n", result.OCSPStapled)
	fmt.Printf("  DNS names:      %s\n", strings.Join(result.DNSNames, ", "))
}

func RunHeaders(input string, timeout time.Duration, insecure, jsonOutput bool) {
	target, err := core.ParseTarget(input)
	if err != nil {
		PrintError(err)
		return
	}
	result := core.ProbeHTTP(context.Background(), target.CandidateURLs, timeout, insecure)
	if result.Error != "" {
		PrintError(fmt.Errorf("HTTP: %s", result.Error))
		return
	}
	audit := core.AuditHeaders(result)
	if jsonOutput {
		PrintJSON(map[string]any{"url": result.FinalURL, "audit": audit})
		return
	}
	fmt.Printf("\n%s %s\n", Color(bold, "Security header audit:"), result.FinalURL)
	fmt.Printf("Grade %s · %d/100\n\n", audit.Grade, audit.Score)
	for _, check := range audit.Checks {
		mark := Color(red, "✗")
		if check.Passed {
			mark = Color(green, "✓")
		}
		fmt.Printf("  %s %-29s %s\n", mark, check.Name, check.Message)
		if check.Value != "" {
			fmt.Printf("      %s\n", Color(dim, compact(check.Value, 120)))
		}
	}
}

func RunPorts(host, portSpec string, timeout time.Duration, concurrency int, includeClosed, jsonOutput bool) {
	target, err := core.ParseTarget(host)
	if err != nil {
		PrintError(err)
		return
	}
	ports, err := core.ParsePorts(portSpec, 128)
	if err != nil {
		PrintError(err)
		return
	}
	result := core.ScanPorts(context.Background(), target.Host, ports, timeout, concurrency, includeClosed)
	if jsonOutput {
		PrintJSON(result)
		return
	}
	fmt.Printf("\n%s %s · %d ports · %d ms\n", Color(bold, "Port check:"), result.Host, len(ports), result.DurationMS)
	if len(result.Open) == 0 {
		fmt.Println("  " + Color(yellow, "No open ports found in the supplied list."))
	}
	for _, item := range result.Open {
		fmt.Printf("  %s %-5s %-20s %s\n", Color(green, "OPEN"), item.Port, emptyAs(item.Service, "unknown"), core.FormatMS(item.AvgMS))
	}
	if includeClosed {
		for _, item := range result.Closed {
			fmt.Printf("  %s %-5s %-20s\n", Color(dim, "closed"), item.Port, emptyAs(item.Service, "unknown"))
		}
	}
}

func RunBenchmark(rawURL, method string, requests, concurrency int, timeout time.Duration, insecure, jsonOutput bool) {
	result, err := core.BenchmarkHTTP(context.Background(), rawURL, method, requests, concurrency, timeout, insecure)
	if err != nil {
		PrintError(err)
		return
	}
	if jsonOutput {
		PrintJSON(result)
		return
	}
	fmt.Printf("\n%s %s\n", Color(bold, "HTTP benchmark:"), result.URL)
	fmt.Printf("  Requests:       %d (%d concurrent)\n", result.Requests, result.Concurrency)
	fmt.Printf("  Success:        %d/%d (%.1f%%)\n", result.Successes, result.Requests, result.SuccessRate)
	fmt.Printf("  Throughput:     %.2f req/s\n", result.RequestsPerSecond)
	fmt.Printf("  Latency min:    %.1f ms\n", result.MinMS)
	fmt.Printf("  Latency average:%.1f ms\n", result.AverageMS)
	fmt.Printf("  Latency p50:    %.1f ms\n", result.P50MS)
	fmt.Printf("  Latency p90:    %.1f ms\n", result.P90MS)
	fmt.Printf("  Latency p95:    %.1f ms\n", result.P95MS)
	fmt.Printf("  Latency p99:    %.1f ms\n", result.P99MS)
	fmt.Printf("  Latency max:    %.1f ms\n", result.MaxMS)
	fmt.Println("  Status codes:")
	for code, count := range result.StatusDistribution {
		fmt.Printf("    HTTP %d: %d\n", code, count)
	}
	for message, count := range result.Errors {
		fmt.Printf("    %s: %d\n", compact(message, 100), count)
	}
}
