package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxLogReadBytes int64 = 64 * 1024 * 1024

type LogFilter struct {
	Contains    string `json:"contains,omitempty"`
	Level       string `json:"level,omitempty"`
	Status      string `json:"status,omitempty"`
	IP          string `json:"ip,omitempty"`
	Method      string `json:"method,omitempty"`
	RequestPath string `json:"request_path,omitempty"`
}

type LogEntry struct {
	Raw       string `json:"raw"`
	Timestamp string `json:"timestamp,omitempty"`
	Level     string `json:"level,omitempty"`
	IP        string `json:"ip,omitempty"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Status    int    `json:"status,omitempty"`
	Bytes     int64  `json:"bytes,omitempty"`
	Message   string `json:"message,omitempty"`
	Format    string `json:"format"`
}

type LogViewResult struct {
	File          string         `json:"file"`
	ScannedLines  int            `json:"scanned_lines"`
	MatchedLines  int            `json:"matched_lines"`
	Entries       []LogEntry     `json:"entries"`
	StatusCounts  map[int]int    `json:"status_counts,omitempty"`
	LevelCounts   map[string]int `json:"level_counts,omitempty"`
	TruncatedScan bool           `json:"truncated_scan"`
}

var (
	combinedLogPattern = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([A-Z]+) ([^\s"]+)(?: HTTP/[^"]+)?" (\d{3}) (\S+)`)
	levelPattern       = regexp.MustCompile(`(?i)(?:^|[\s\[])\b(trace|debug|info|notice|warn|warning|error|err|critical|crit|fatal|panic)\b`)
	timestampPattern   = regexp.MustCompile(`^(\d{4}[-/]\d{2}[-/]\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?(?:Z|[+-]\d{2}:?\d{2})?)`)
)

func ReadLogEntries(path string, limit int, filter LogFilter) (LogViewResult, error) {
	if limit < 1 {
		limit = 100
	}
	if err := ValidateLogFilter(filter); err != nil {
		return LogViewResult{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return LogViewResult{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return LogViewResult{}, err
	}
	truncated := false
	if info.Size() > maxLogReadBytes {
		start := info.Size() - maxLogReadBytes
		truncated = true
		if _, err := file.Seek(start, io.SeekStart); err != nil {
			return LogViewResult{}, err
		}
		reader := bufio.NewReader(file)
		_, _ = reader.ReadString('\n')
		return scanLogReader(path, reader, limit, filter, truncated)
	}
	return scanLogReader(path, bufio.NewReader(file), limit, filter, truncated)
}

func scanLogReader(path string, reader io.Reader, limit int, filter LogFilter, truncated bool) (LogViewResult, error) {
	result := LogViewResult{
		File:          path,
		Entries:       make([]LogEntry, 0, limit),
		StatusCounts:  map[int]int{},
		LevelCounts:   map[string]int{},
		TruncatedScan: truncated,
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		result.ScannedLines++
		entry := ParseLogLine(scanner.Text())
		if !MatchLogEntry(entry, filter) {
			continue
		}
		result.MatchedLines++
		if entry.Status != 0 {
			result.StatusCounts[entry.Status]++
		}
		if entry.Level != "" {
			result.LevelCounts[entry.Level]++
		}
		if len(result.Entries) == limit {
			copy(result.Entries, result.Entries[1:])
			result.Entries[len(result.Entries)-1] = entry
		} else {
			result.Entries = append(result.Entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return LogViewResult{}, err
	}
	if len(result.StatusCounts) == 0 {
		result.StatusCounts = nil
	}
	if len(result.LevelCounts) == 0 {
		result.LevelCounts = nil
	}
	return result, nil
}

func ParseLogLine(line string) LogEntry {
	entry := LogEntry{Raw: line, Format: "text"}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return entry
	}

	var object map[string]any
	if strings.HasPrefix(trimmed, "{") && json.Unmarshal([]byte(trimmed), &object) == nil {
		entry.Format = "json"
		entry.Timestamp = firstJSONText(object, "timestamp", "time", "ts", "date", "datetime")
		entry.Level = normalizeLevel(firstJSONText(object, "level", "severity", "log_level"))
		entry.IP = firstJSONText(object, "ip", "client_ip", "remote_addr", "remote_ip")
		entry.Method = strings.ToUpper(firstJSONText(object, "method", "http_method", "request_method"))
		entry.Path = firstJSONText(object, "path", "url", "uri", "request_uri")
		entry.Status = firstJSONInt(object, "status", "status_code", "http_status", "response_status")
		entry.Bytes = int64(firstJSONInt(object, "bytes", "body_bytes_sent", "response_bytes", "size"))
		entry.Message = firstJSONText(object, "message", "msg", "error", "event")
		return entry
	}

	if matches := combinedLogPattern.FindStringSubmatch(trimmed); len(matches) == 7 {
		entry.Format = "combined"
		entry.IP = matches[1]
		entry.Timestamp = matches[2]
		entry.Method = matches[3]
		entry.Path = matches[4]
		entry.Status, _ = strconv.Atoi(matches[5])
		if matches[6] != "-" {
			entry.Bytes, _ = strconv.ParseInt(matches[6], 10, 64)
		}
		return entry
	}

	if matches := timestampPattern.FindStringSubmatch(trimmed); len(matches) > 1 {
		entry.Timestamp = matches[1]
	}
	if matches := levelPattern.FindStringSubmatch(trimmed); len(matches) > 1 {
		entry.Level = normalizeLevel(matches[1])
	}
	entry.Message = trimmed
	return entry
}

func ValidateLogFilter(filter LogFilter) error {
	if filter.Status == "" {
		return nil
	}
	_, _, err := parseStatusFilter(filter.Status)
	return err
}

func MatchLogEntry(entry LogEntry, filter LogFilter) bool {
	if filter.Contains != "" && !strings.Contains(strings.ToLower(entry.Raw), strings.ToLower(filter.Contains)) {
		return false
	}
	if filter.Level != "" && normalizeLevel(entry.Level) != normalizeLevel(filter.Level) {
		return false
	}
	if filter.IP != "" && !strings.Contains(strings.ToLower(entry.IP), strings.ToLower(filter.IP)) {
		return false
	}
	if filter.Method != "" && !strings.EqualFold(entry.Method, filter.Method) {
		return false
	}
	if filter.RequestPath != "" && !strings.Contains(strings.ToLower(entry.Path), strings.ToLower(filter.RequestPath)) {
		return false
	}
	if filter.Status != "" {
		min, max, err := parseStatusFilter(filter.Status)
		if err != nil || entry.Status < min || entry.Status > max {
			return false
		}
	}
	return true
}

func FollowLog(ctx context.Context, path string, interval time.Duration, filter LogFilter, onEntry func(LogEntry)) error {
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	if err := ValidateLogFilter(filter); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	offset := info.Size()
	pending := ""
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			info, err := file.Stat()
			if err != nil {
				return err
			}
			if info.Size() < offset {
				offset = 0
				pending = ""
			}
			if info.Size() == offset {
				continue
			}
			remaining := info.Size() - offset
			if remaining > 8*1024*1024 {
				offset = info.Size() - 8*1024*1024
				remaining = 8 * 1024 * 1024
				pending = ""
			}
			buffer := make([]byte, remaining)
			n, err := file.ReadAt(buffer, offset)
			if err != nil && err != io.EOF {
				return err
			}
			offset += int64(n)
			text := pending + string(buffer[:n])
			parts := strings.Split(text, "\n")
			pending = parts[len(parts)-1]
			for _, line := range parts[:len(parts)-1] {
				entry := ParseLogLine(strings.TrimSuffix(line, "\r"))
				if MatchLogEntry(entry, filter) {
					onEntry(entry)
				}
			}
		}
	}
}

func parseStatusFilter(spec string) (int, int, error) {
	spec = strings.ToLower(strings.TrimSpace(spec))
	if len(spec) == 3 && spec[1:] == "xx" && spec[0] >= '1' && spec[0] <= '5' {
		base := int(spec[0]-'0') * 100
		return base, base + 99, nil
	}
	if strings.Contains(spec, "-") {
		parts := strings.SplitN(spec, "-", 2)
		min, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		max, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil || min < 100 || max > 599 || min > max {
			return 0, 0, fmt.Errorf("invalid status filter %q", spec)
		}
		return min, max, nil
	}
	status, err := strconv.Atoi(spec)
	if err != nil || status < 100 || status > 599 {
		return 0, 0, fmt.Errorf("invalid status filter %q; use 500, 500-599, or 5xx", spec)
	}
	return status, status, nil
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "warning":
		return "warn"
	case "err":
		return "error"
	case "critical":
		return "crit"
	default:
		return strings.ToLower(strings.TrimSpace(level))
	}
}

func firstJSONText(object map[string]any, keys ...string) string {
	for _, key := range keys {
		for actual, value := range object {
			if !strings.EqualFold(actual, key) {
				continue
			}
			switch typed := value.(type) {
			case string:
				return typed
			case json.Number:
				return typed.String()
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			}
		}
	}
	return ""
}

func firstJSONInt(object map[string]any, keys ...string) int {
	for _, key := range keys {
		for actual, value := range object {
			if !strings.EqualFold(actual, key) {
				continue
			}
			switch typed := value.(type) {
			case float64:
				return int(typed)
			case string:
				parsed, _ := strconv.Atoi(typed)
				return parsed
			case json.Number:
				parsed, _ := strconv.Atoi(typed.String())
				return parsed
			}
		}
	}
	return 0
}
