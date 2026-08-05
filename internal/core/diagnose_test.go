package core

import "testing"

func TestAnalyzeHealthy(t *testing.T) {
	report := Report{
		DNS:  []DNSResult{{Resolver: "System resolver", Addresses: []string{"203.0.113.10"}}},
		TCP:  []TCPResult{{Port: "443", Attempts: 3, Successes: 3, AvgMS: 20}},
		TLS:  &TLSResult{DaysRemaining: 90},
		HTTP: &HTTPResult{StatusCode: 200, Timings: HTTPTimings{TTFBMS: 100}},
	}
	Analyze(&report)
	if report.Score < 90 || report.Verdict != "Healthy" {
		t.Fatalf("unexpected result: score=%d verdict=%q", report.Score, report.Verdict)
	}
}

func TestAnalyzeResolverMismatch(t *testing.T) {
	report := Report{
		DNS: []DNSResult{
			{Resolver: "System resolver", Error: "failed"},
			{Resolver: "Cloudflare", Addresses: []string{"203.0.113.10"}},
		},
		TCP:  []TCPResult{{Port: "443", Attempts: 1, Successes: 1}},
		TLS:  &TLSResult{DaysRemaining: 90},
		HTTP: &HTTPResult{StatusCode: 200},
	}
	Analyze(&report)
	if len(report.Recommendations) == 0 {
		t.Fatal("expected a recommendation")
	}
}
