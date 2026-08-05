package core

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const defaultDumpMaxBytes int64 = 16 * 1024 * 1024

type SiteDumpOptions struct {
	OutputDir   string
	Timeout     time.Duration
	MaxBytes    int64
	InsecureTLS bool
}

type SiteDumpResult struct {
	CapturedAt      time.Time           `json:"captured_at"`
	RequestedURL    string              `json:"requested_url"`
	FinalURL        string              `json:"final_url"`
	StatusCode      int                 `json:"status_code"`
	Status          string              `json:"status"`
	Protocol        string              `json:"protocol"`
	ContentType     string              `json:"content_type,omitempty"`
	ContentLength   int64               `json:"content_length,omitempty"`
	BytesSaved      int64               `json:"bytes_saved"`
	Truncated       bool                `json:"truncated"`
	SHA256          string              `json:"sha256"`
	Redirects       []string            `json:"redirects,omitempty"`
	OutputDir       string              `json:"output_dir"`
	BodyFile        string              `json:"body_file"`
	HeadersFile     string              `json:"headers_file"`
	MetadataFile    string              `json:"metadata_file"`
	ResponseHeaders map[string][]string `json:"response_headers"`
}

func DumpSite(ctx context.Context, input string, opts SiteDumpOptions) (SiteDumpResult, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 15 * time.Second
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = defaultDumpMaxBytes
	}
	if opts.MaxBytes > 128*1024*1024 {
		return SiteDumpResult{}, fmt.Errorf("maximum dump size is 128 MiB")
	}

	target, err := ParseTarget(input)
	if err != nil {
		return SiteDumpResult{}, err
	}
	if len(target.CandidateURLs) == 0 {
		return SiteDumpResult{}, fmt.Errorf("no URL candidate available")
	}

	var lastErr error
	for _, candidate := range target.CandidateURLs {
		result, err := dumpSingleURL(ctx, candidate, opts)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return SiteDumpResult{}, lastErr
}

func dumpSingleURL(parent context.Context, rawURL string, opts SiteDumpOptions) (SiteDumpResult, error) {
	ctx, cancel := context.WithTimeout(parent, opts.Timeout)
	defer cancel()

	redirects := make([]string, 0, 4)
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: opts.Timeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: opts.InsecureTLS}, // #nosec G402 -- explicit troubleshooting option.
		TLSHandshakeTimeout:   opts.Timeout,
		ResponseHeaderTimeout: opts.Timeout,
		ForceAttemptHTTP2:     true,
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirects = append(redirects, req.URL.String())
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return SiteDumpResult{}, err
	}
	req.Header.Set("User-Agent", "PulseNet/2.1 site-dump")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return SiteDumpResult{}, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, opts.MaxBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return SiteDumpResult{}, err
	}
	truncated := int64(len(body)) > opts.MaxBytes
	if truncated {
		body = body[:opts.MaxBytes]
	}

	finalURL := resp.Request.URL.String()
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = defaultDumpDirectory(finalURL)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return SiteDumpResult{}, err
	}

	contentType := resp.Header.Get("Content-Type")
	bodyName := dumpBodyName(contentType)
	bodyPath := filepath.Join(outputDir, bodyName)
	if err := os.WriteFile(bodyPath, body, 0o644); err != nil {
		return SiteDumpResult{}, err
	}

	headers := sanitizeDumpHeaders(resp.Header)
	headersPath := filepath.Join(outputDir, "headers.json")
	if err := writeIndentedJSON(headersPath, headers); err != nil {
		return SiteDumpResult{}, err
	}

	digest := sha256.Sum256(body)
	result := SiteDumpResult{
		CapturedAt:      time.Now(),
		RequestedURL:    rawURL,
		FinalURL:        finalURL,
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		Protocol:        resp.Proto,
		ContentType:     contentType,
		ContentLength:   resp.ContentLength,
		BytesSaved:      int64(len(body)),
		Truncated:       truncated,
		SHA256:          hex.EncodeToString(digest[:]),
		Redirects:       redirects,
		OutputDir:       outputDir,
		BodyFile:        bodyName,
		HeadersFile:     "headers.json",
		MetadataFile:    "metadata.json",
		ResponseHeaders: headers,
	}
	metadataPath := filepath.Join(outputDir, result.MetadataFile)
	if err := writeIndentedJSON(metadataPath, result); err != nil {
		return SiteDumpResult{}, err
	}
	return result, nil
}

func writeIndentedJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func sanitizeDumpHeaders(headers http.Header) map[string][]string {
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		switch strings.ToLower(key) {
		case "set-cookie", "proxy-authenticate", "www-authenticate":
			out[key] = []string{"<redacted>"}
		default:
			out[key] = append([]string(nil), values...)
		}
	}
	return out
}

func dumpBodyName(contentType string) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch mediaType {
	case "text/html", "application/xhtml+xml":
		return "page.html"
	case "application/json", "application/ld+json":
		return "body.json"
	case "text/css":
		return "body.css"
	case "application/javascript", "text/javascript":
		return "body.js"
	case "text/plain", "text/xml", "application/xml":
		return "body.txt"
	default:
		return "body.bin"
	}
}

var dumpNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func defaultDumpDirectory(rawURL string) string {
	host := "site"
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Hostname() != "" {
		host = parsed.Hostname()
	}
	host = strings.Trim(dumpNameCleaner.ReplaceAllString(host, "-"), "-.")
	if host == "" {
		host = "site"
	}
	return filepath.Join("site-dumps", host+"-"+time.Now().Format("20060102-150405"))
}
