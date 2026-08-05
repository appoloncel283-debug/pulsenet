package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

func RunDiagnosis(ctx context.Context, target Target, opts Options) Report {
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.TCPAttempts <= 0 {
		opts.TCPAttempts = 3
	}
	if len(opts.Ports) == 0 {
		opts.Ports = target.DefaultPorts()
	}
	started := time.Now()
	report := Report{GeneratedAt: started, Target: target}

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		report.DNS = ProbeDNS(ctx, target.Host, opts.Timeout)
	}()
	go func() {
		defer wg.Done()
		results := make([]TCPResult, len(opts.Ports))
		var portsWG sync.WaitGroup
		for i, port := range opts.Ports {
			portsWG.Add(1)
			go func(index int, value string) {
				defer portsWG.Done()
				results[index] = ProbeTCP(ctx, target.Host, value, opts.TCPAttempts, opts.Timeout)
			}(i, port)
		}
		portsWG.Wait()
		report.TCP = results
	}()
	go func() {
		defer wg.Done()
		result := ProbeHTTP(ctx, target.CandidateURLs, opts.Timeout*2, opts.InsecureTLS)
		report.HTTP = &result
		if result.Error == "" {
			audit := AuditHeaders(result)
			report.SecurityHeaders = &audit
		}
	}()
	go func() {
		defer wg.Done()
		port := "443"
		probe := true
		if target.ExplicitPort != "" {
			port = target.ExplicitPort
		}
		if target.ExplicitURL && len(target.CandidateURLs) > 0 {
			if parsed, err := url.Parse(target.CandidateURLs[0]); err == nil {
				probe = strings.EqualFold(parsed.Scheme, "https")
				if parsed.Port() != "" {
					port = parsed.Port()
				}
			}
		}
		if probe {
			result := ProbeTLS(ctx, target.Host, port, opts.Timeout, opts.InsecureTLS)
			report.TLS = &result
		}
	}()
	wg.Wait()
	report.TotalDurationMS = time.Since(started).Milliseconds()
	Analyze(&report)
	return report
}

func Analyze(report *Report) {
	score := 100
	recommendations := make([]string, 0)
	systemDNSOK := len(report.DNS) > 0 && report.DNS[0].Error == "" && len(report.DNS[0].Addresses) > 0
	publicDNSOK := false
	for _, result := range report.DNS[1:] {
		if result.Error == "" && len(result.Addresses) > 0 {
			publicDNSOK = true
			break
		}
	}
	if !systemDNSOK {
		score -= 30
		if publicDNSOK {
			recommendations = append(recommendations, "The system resolver failed while public resolvers succeeded. Review local DNS settings, VPN software, or the router DNS cache.")
		} else {
			recommendations = append(recommendations, "The hostname could not be resolved. Verify the name and check upstream DNS connectivity.")
		}
	}

	anyTCP := false
	for _, result := range report.TCP {
		if result.Successes == 0 {
			continue
		}
		anyTCP = true
		if result.LossPct > 0 {
			score -= 8
			recommendations = append(recommendations, fmt.Sprintf("TCP port %s was unstable (%.0f%% failed attempts). Check Wi-Fi quality, VPNs, firewalls, and the upstream route.", result.Port, result.LossPct))
		}
		if result.AvgMS > 300 {
			score -= 8
			recommendations = append(recommendations, fmt.Sprintf("TCP latency to port %s is high (%.0f ms average).", result.Port, result.AvgMS))
		}
	}
	if !anyTCP {
		score -= 35
		recommendations = append(recommendations, "None of the tested TCP ports accepted a connection. Confirm the service is running and that firewalls or security groups allow access.")
	}

	if report.TLS != nil {
		if report.TLS.Error != "" {
			score -= 18
			recommendations = append(recommendations, "TLS validation or negotiation failed. Check the certificate chain, hostname, system clock, and TLS interception software.")
		} else if report.TLS.DaysRemaining < 0 {
			score -= 30
			recommendations = append(recommendations, "The TLS certificate has expired.")
		} else if report.TLS.DaysRemaining <= 14 {
			score -= 10
			recommendations = append(recommendations, fmt.Sprintf("The TLS certificate expires in %d days.", report.TLS.DaysRemaining))
		}
	}

	if report.HTTP != nil {
		if report.HTTP.Error != "" {
			score -= 20
			recommendations = append(recommendations, "The HTTP request failed. Check proxy settings, TLS errors, redirects, and application availability.")
		} else {
			switch {
			case report.HTTP.StatusCode >= 500:
				score -= 25
				recommendations = append(recommendations, "The server returned a 5xx response, which points to an application or upstream service failure.")
			case report.HTTP.StatusCode >= 400:
				score -= 8
				recommendations = append(recommendations, fmt.Sprintf("The server is reachable but returned HTTP %d.", report.HTTP.StatusCode))
			}
			if report.HTTP.Timings.TTFBMS > 1500 {
				score -= 8
				recommendations = append(recommendations, fmt.Sprintf("Time to first byte is slow (%d ms). Investigate server processing time and upstream dependencies.", report.HTTP.Timings.TTFBMS))
			}
		}
	}
	if report.SecurityHeaders != nil && report.SecurityHeaders.Score < 50 {
		score -= 5
		recommendations = append(recommendations, "Several recommended browser security headers are missing. Run the headers command for details.")
	}
	if score < 0 {
		score = 0
	}
	report.Score = score
	report.Recommendations = dedupeStrings(recommendations)
	switch {
	case score >= 90:
		report.Verdict = "Healthy"
	case score >= 70:
		report.Verdict = "Reachable with warnings"
	case score >= 45:
		report.Verdict = "Degraded"
	default:
		report.Verdict = "Unavailable or severely degraded"
	}
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
