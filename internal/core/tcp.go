package core

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var serviceNames = map[string]string{
	"20": "FTP data", "21": "FTP", "22": "SSH", "23": "Telnet", "25": "SMTP",
	"53": "DNS", "67": "DHCP", "68": "DHCP", "80": "HTTP", "110": "POP3",
	"123": "NTP", "143": "IMAP", "389": "LDAP", "443": "HTTPS", "445": "SMB",
	"465": "SMTPS", "587": "SMTP submission", "636": "LDAPS", "993": "IMAPS",
	"995": "POP3S", "1433": "MSSQL", "1521": "Oracle", "2049": "NFS",
	"2375": "Docker", "2376": "Docker TLS", "3000": "Web app", "3306": "MySQL",
	"3389": "RDP", "5432": "PostgreSQL", "5672": "AMQP", "6379": "Redis",
	"8000": "HTTP alt", "8080": "HTTP proxy", "8443": "HTTPS alt", "9000": "Web app",
	"9200": "Elasticsearch", "27017": "MongoDB",
}

func ServiceName(port string) string { return serviceNames[port] }

func ProbeTCP(parent context.Context, host, port string, attempts int, timeout time.Duration) TCPResult {
	if attempts < 1 {
		attempts = 1
	}
	result := TCPResult{Port: port, Service: ServiceName(port), Attempts: attempts}
	latencies := make([]float64, 0, attempts)
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(parent, timeout)
		started := time.Now()
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(host, port))
		elapsed := float64(time.Since(started).Microseconds()) / 1000
		cancel()
		if err != nil {
			result.LastError = err.Error()
			continue
		}
		_ = conn.Close()
		result.Successes++
		latencies = append(latencies, elapsed)
	}
	result.LossPct = float64(attempts-result.Successes) / float64(attempts) * 100
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		result.MinMS = latencies[0]
		result.MaxMS = latencies[len(latencies)-1]
		var sum float64
		for _, value := range latencies {
			sum += value
		}
		result.AvgMS = sum / float64(len(latencies))
		if len(latencies) > 1 {
			var variance float64
			for _, value := range latencies {
				delta := value - result.AvgMS
				variance += delta * delta
			}
			result.JitterMS = math.Sqrt(variance / float64(len(latencies)))
		}
	}
	return result
}

func ParsePorts(value string, maxPorts int) ([]string, error) {
	if maxPorts <= 0 {
		maxPorts = 128
	}
	seen := map[int]struct{}{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || start < 1 || end > 65535 || start > end {
				return nil, fmt.Errorf("invalid port range %q", part)
			}
			for port := start; port <= end; port++ {
				seen[port] = struct{}{}
				if len(seen) > maxPorts {
					return nil, fmt.Errorf("too many ports; maximum is %d", maxPorts)
				}
			}
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid port %q", part)
		}
		seen[port] = struct{}{}
		if len(seen) > maxPorts {
			return nil, fmt.Errorf("too many ports; maximum is %d", maxPorts)
		}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("no ports supplied")
	}
	ports := make([]int, 0, len(seen))
	for port := range seen {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	out := make([]string, len(ports))
	for i, port := range ports {
		out[i] = strconv.Itoa(port)
	}
	return out, nil
}

func ScanPorts(ctx context.Context, host string, ports []string, timeout time.Duration, concurrency int, includeClosed bool) PortScanResult {
	if concurrency < 1 {
		concurrency = 32
	}
	if concurrency > 64 {
		concurrency = 64
	}
	started := time.Now()
	result := PortScanResult{Host: host, StartedAt: started}
	type job struct{ port string }
	jobs := make(chan job)
	results := make(chan TCPResult, len(ports))
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results <- ProbeTCP(ctx, host, item.port, 1, timeout)
			}
		}()
	}
	go func() {
		for _, port := range ports {
			jobs <- job{port: port}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()
	for item := range results {
		if item.Successes > 0 {
			result.Open = append(result.Open, item)
		} else if includeClosed {
			result.Closed = append(result.Closed, item)
		}
	}
	sort.Slice(result.Open, func(i, j int) bool {
		a, _ := strconv.Atoi(result.Open[i].Port)
		b, _ := strconv.Atoi(result.Open[j].Port)
		return a < b
	})
	sort.Slice(result.Closed, func(i, j int) bool {
		a, _ := strconv.Atoi(result.Closed[i].Port)
		b, _ := strconv.Atoi(result.Closed[j].Port)
		return a < b
	})
	result.DurationMS = time.Since(started).Milliseconds()
	return result
}
