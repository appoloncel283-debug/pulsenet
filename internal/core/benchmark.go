package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type benchSample struct {
	duration float64
	status   int
	err      string
}

func BenchmarkHTTP(ctx context.Context, rawURL, method string, requests, concurrency int, timeout time.Duration, insecure bool) (BenchmarkResult, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodGet && method != http.MethodHead {
		return BenchmarkResult{}, fmt.Errorf("method must be GET or HEAD")
	}
	if requests < 1 || requests > 10000 {
		return BenchmarkResult{}, fmt.Errorf("requests must be between 1 and 10000")
	}
	if concurrency < 1 || concurrency > 100 {
		return BenchmarkResult{}, fmt.Errorf("concurrency must be between 1 and 100")
	}
	if concurrency > requests {
		concurrency = requests
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecure}, // #nosec G402 -- controlled by an explicit diagnostic flag.
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		MaxIdleConns:          concurrency * 2,
		MaxIdleConnsPerHost:   concurrency,
		ForceAttemptHTTP2:     true,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: timeout}

	jobs := make(chan struct{})
	samples := make(chan benchSample, requests)
	var wg sync.WaitGroup
	started := time.Now()
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
				if err != nil {
					samples <- benchSample{err: err.Error()}
					continue
				}
				req.Header.Set("User-Agent", "PulseNet/2.0 benchmark")
				t0 := time.Now()
				resp, err := client.Do(req)
				elapsed := float64(time.Since(t0).Microseconds()) / 1000
				if err != nil {
					samples <- benchSample{duration: elapsed, err: compactNetworkError(err.Error())}
					continue
				}
				if method != http.MethodHead {
					_, _ = io.CopyN(io.Discard, resp.Body, 128*1024)
				}
				_ = resp.Body.Close()
				samples <- benchSample{duration: elapsed, status: resp.StatusCode}
			}
		}()
	}
	go func() {
		for i := 0; i < requests; i++ {
			jobs <- struct{}{}
		}
		close(jobs)
		wg.Wait()
		close(samples)
	}()

	result := BenchmarkResult{URL: rawURL, Method: method, Requests: requests, Concurrency: concurrency, StatusDistribution: map[int]int{}, Errors: map[string]int{}}
	latencies := make([]float64, 0, requests)
	for sample := range samples {
		if sample.err != "" {
			result.Failures++
			result.Errors[sample.err]++
			continue
		}
		result.Successes++
		result.StatusDistribution[sample.status]++
		latencies = append(latencies, sample.duration)
	}
	result.DurationMS = time.Since(started).Milliseconds()
	result.SuccessRate = float64(result.Successes) / float64(requests) * 100
	if result.DurationMS > 0 {
		result.RequestsPerSecond = float64(requests) / (float64(result.DurationMS) / 1000)
	}
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		result.MinMS = latencies[0]
		result.MaxMS = latencies[len(latencies)-1]
		var sum float64
		for _, latency := range latencies {
			sum += latency
		}
		result.AverageMS = sum / float64(len(latencies))
		result.P50MS = percentile(latencies, 0.50)
		result.P90MS = percentile(latencies, 0.90)
		result.P95MS = percentile(latencies, 0.95)
		result.P99MS = percentile(latencies, 0.99)
	}
	if len(result.Errors) == 0 {
		result.Errors = nil
	}
	return result, nil
}

func percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	index := int(math.Ceil(p*float64(len(sortedValues)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sortedValues) {
		index = len(sortedValues) - 1
	}
	return sortedValues[index]
}

func compactNetworkError(value string) string {
	for _, marker := range []string{" dial tcp ", " request canceled", " context deadline exceeded"} {
		if strings.Contains(value, marker) {
			return strings.TrimSpace(value)
		}
	}
	if len(value) > 160 {
		return value[:157] + "..."
	}
	return value
}
