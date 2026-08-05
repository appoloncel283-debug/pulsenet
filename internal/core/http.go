package core

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

func ProbeHTTP(ctx context.Context, candidates []string, timeout time.Duration, insecure bool) HTTPResult {
	if len(candidates) == 0 {
		return HTTPResult{Error: "no URL candidates available"}
	}
	type resultItem struct {
		index  int
		result HTTPResult
	}
	ch := make(chan resultItem, len(candidates))
	for i, candidate := range candidates {
		go func(index int, rawURL string) {
			ch <- resultItem{index: index, result: probeSingleHTTP(ctx, rawURL, timeout, insecure)}
		}(i, candidate)
	}
	results := make([]HTTPResult, len(candidates))
	for range candidates {
		item := <-ch
		results[item.index] = item.result
	}
	for _, result := range results {
		if result.Error == "" {
			return result
		}
	}
	return results[0]
}

func probeSingleHTTP(parent context.Context, rawURL string, timeout time.Duration, insecure bool) HTTPResult {
	result := HTTPResult{URL: rawURL}
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecure}, // #nosec G402 -- controlled by an explicit diagnostic flag.
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ForceAttemptHTTP2:     true,
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			result.Redirects = append(result.Redirects, req.URL.String())
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "PulseNet/2.0 (+https://github.com/appoloncel283-debug/pulsenet)")
	req.Header.Set("Accept", "*/*")

	var dnsStart, connectStart, tlsStart time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				result.Timings.DNSMS = time.Since(dnsStart).Milliseconds()
			}
		},
		ConnectStart: func(_, _ string) { connectStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) {
			if !connectStart.IsZero() {
				result.Timings.ConnectMS = time.Since(connectStart).Milliseconds()
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tlsStart.IsZero() {
				result.Timings.TLSMS = time.Since(tlsStart).Milliseconds()
			}
		},
		GotConn:              func(info httptrace.GotConnInfo) { result.RemoteAddress = info.Conn.RemoteAddr().String() },
		GotFirstResponseByte: func() { result.Timings.TTFBMS = time.Since(started).Milliseconds() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := client.Do(req)
	result.Timings.TotalMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	_, _ = io.CopyN(io.Discard, resp.Body, 64*1024)
	result.Timings.TotalMS = time.Since(started).Milliseconds()
	result.FinalURL = resp.Request.URL.String()
	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.Protocol = resp.Proto
	result.Server = resp.Header.Get("Server")
	result.ContentType = resp.Header.Get("Content-Type")
	result.ContentLength = resp.ContentLength
	result.ResponseHeaders = selectedHeaders(resp.Header)
	return result
}

func selectedHeaders(headers http.Header) map[string]string {
	keys := []string{
		"Strict-Transport-Security", "Content-Security-Policy", "X-Content-Type-Options",
		"X-Frame-Options", "Referrer-Policy", "Permissions-Policy",
		"Cross-Origin-Opener-Policy", "Cross-Origin-Resource-Policy", "Cache-Control",
		"Access-Control-Allow-Origin", "Set-Cookie",
	}
	out := make(map[string]string)
	for _, key := range keys {
		values := headers.Values(key)
		if len(values) > 0 {
			out[key] = strings.Join(values, " | ")
		}
	}
	return out
}
