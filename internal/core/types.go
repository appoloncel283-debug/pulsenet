package core

import "time"

type Target struct {
	Input         string   `json:"input"`
	Host          string   `json:"host"`
	ExplicitPort  string   `json:"explicit_port,omitempty"`
	ExplicitURL   bool     `json:"explicit_url"`
	CandidateURLs []string `json:"candidate_urls,omitempty"`
}

type DNSResult struct {
	Resolver   string   `json:"resolver"`
	Addresses  []string `json:"addresses,omitempty"`
	DurationMS int64    `json:"duration_ms"`
	Error      string   `json:"error,omitempty"`
}

type DNSRecords struct {
	Host       string   `json:"host"`
	A          []string `json:"a,omitempty"`
	AAAA       []string `json:"aaaa,omitempty"`
	CNAME      string   `json:"cname,omitempty"`
	MX         []string `json:"mx,omitempty"`
	NS         []string `json:"ns,omitempty"`
	TXT        []string `json:"txt,omitempty"`
	PTR        []string `json:"ptr,omitempty"`
	Errors     []string `json:"errors,omitempty"`
	DurationMS int64    `json:"duration_ms"`
}

type TCPResult struct {
	Port      string  `json:"port"`
	Service   string  `json:"service,omitempty"`
	Attempts  int     `json:"attempts"`
	Successes int     `json:"successes"`
	LossPct   float64 `json:"loss_percent"`
	MinMS     float64 `json:"min_ms,omitempty"`
	AvgMS     float64 `json:"avg_ms,omitempty"`
	MaxMS     float64 `json:"max_ms,omitempty"`
	JitterMS  float64 `json:"jitter_ms,omitempty"`
	LastError string  `json:"last_error,omitempty"`
}

type TLSResult struct {
	Address       string    `json:"address"`
	Protocol      string    `json:"protocol,omitempty"`
	CipherSuite   string    `json:"cipher_suite,omitempty"`
	ALPN          string    `json:"alpn,omitempty"`
	Subject       string    `json:"subject,omitempty"`
	Issuer        string    `json:"issuer,omitempty"`
	SerialNumber  string    `json:"serial_number,omitempty"`
	DNSNames      []string  `json:"dns_names,omitempty"`
	NotBefore     time.Time `json:"not_before,omitempty"`
	NotAfter      time.Time `json:"not_after,omitempty"`
	DaysRemaining int       `json:"days_remaining,omitempty"`
	Verified      bool      `json:"verified"`
	ChainLength   int       `json:"chain_length,omitempty"`
	OCSPStapled   bool      `json:"ocsp_stapled"`
	DurationMS    int64     `json:"duration_ms"`
	Error         string    `json:"error,omitempty"`
}

type HTTPTimings struct {
	DNSMS     int64 `json:"dns_ms,omitempty"`
	ConnectMS int64 `json:"connect_ms,omitempty"`
	TLSMS     int64 `json:"tls_ms,omitempty"`
	TTFBMS    int64 `json:"ttfb_ms,omitempty"`
	TotalMS   int64 `json:"total_ms"`
}

type HTTPResult struct {
	URL             string            `json:"url"`
	FinalURL        string            `json:"final_url,omitempty"`
	StatusCode      int               `json:"status_code,omitempty"`
	Status          string            `json:"status,omitempty"`
	Protocol        string            `json:"protocol,omitempty"`
	RemoteAddress   string            `json:"remote_address,omitempty"`
	Server          string            `json:"server,omitempty"`
	ContentType     string            `json:"content_type,omitempty"`
	ContentLength   int64             `json:"content_length,omitempty"`
	Redirects       []string          `json:"redirects,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	Timings         HTTPTimings       `json:"timings"`
	Error           string            `json:"error,omitempty"`
}

type HeaderCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Weight  int    `json:"weight"`
}

type HeaderAudit struct {
	Score  int           `json:"score"`
	Grade  string        `json:"grade"`
	Checks []HeaderCheck `json:"checks"`
}

type Report struct {
	GeneratedAt     time.Time    `json:"generated_at"`
	Target          Target       `json:"target"`
	DNS             []DNSResult  `json:"dns"`
	TCP             []TCPResult  `json:"tcp"`
	TLS             *TLSResult   `json:"tls,omitempty"`
	HTTP            *HTTPResult  `json:"http,omitempty"`
	SecurityHeaders *HeaderAudit `json:"security_headers,omitempty"`
	Verdict         string       `json:"verdict"`
	Score           int          `json:"score"`
	Recommendations []string     `json:"recommendations,omitempty"`
	TotalDurationMS int64        `json:"total_duration_ms"`
}

type Options struct {
	Timeout     time.Duration
	TCPAttempts int
	Ports       []string
	InsecureTLS bool
}

type PortScanResult struct {
	Host       string      `json:"host"`
	StartedAt  time.Time   `json:"started_at"`
	DurationMS int64       `json:"duration_ms"`
	Open       []TCPResult `json:"open"`
	Closed     []TCPResult `json:"closed,omitempty"`
}

type BenchmarkResult struct {
	URL                string         `json:"url"`
	Method             string         `json:"method"`
	Requests           int            `json:"requests"`
	Concurrency        int            `json:"concurrency"`
	Successes          int            `json:"successes"`
	Failures           int            `json:"failures"`
	SuccessRate        float64        `json:"success_rate"`
	RequestsPerSecond  float64        `json:"requests_per_second"`
	MinMS              float64        `json:"min_ms"`
	AverageMS          float64        `json:"average_ms"`
	P50MS              float64        `json:"p50_ms"`
	P90MS              float64        `json:"p90_ms"`
	P95MS              float64        `json:"p95_ms"`
	P99MS              float64        `json:"p99_ms"`
	MaxMS              float64        `json:"max_ms"`
	StatusDistribution map[int]int    `json:"status_distribution"`
	Errors             map[string]int `json:"errors,omitempty"`
	DurationMS         int64          `json:"duration_ms"`
}

type WatchSample struct {
	Timestamp  time.Time `json:"timestamp"`
	Up         bool      `json:"up"`
	StatusCode int       `json:"status_code,omitempty"`
	LatencyMS  int64     `json:"latency_ms,omitempty"`
	Error      string    `json:"error,omitempty"`
}
