package core

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func ParseTarget(input string) (Target, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Target{}, fmt.Errorf("target cannot be empty")
	}

	target := Target{Input: input}
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err != nil || u.Hostname() == "" {
			return Target{}, fmt.Errorf("invalid URL: %q", input)
		}
		target.Host = u.Hostname()
		target.ExplicitPort = u.Port()
		target.ExplicitURL = true
		target.CandidateURLs = []string{u.String()}
		return target, nil
	}

	if host, port, err := net.SplitHostPort(input); err == nil {
		target.Host = strings.Trim(host, "[]")
		target.ExplicitPort = port
		target.CandidateURLs = candidateURLs(target.Host, port)
		return target, nil
	}

	trimmed := strings.Trim(input, "[]")
	if net.ParseIP(trimmed) != nil {
		target.Host = trimmed
		target.CandidateURLs = candidateURLs(trimmed, "")
		return target, nil
	}

	if strings.Contains(input, "/") || strings.ContainsAny(input, " \t\r\n") {
		return Target{}, fmt.Errorf("invalid hostname or URL: %q", input)
	}
	target.Host = strings.TrimSuffix(input, ".")
	if target.Host == "" {
		return Target{}, fmt.Errorf("invalid target")
	}
	target.CandidateURLs = candidateURLs(target.Host, "")
	return target, nil
}

func candidateURLs(host, port string) []string {
	hostForURL := host
	if strings.Contains(host, ":") {
		hostForURL = "[" + host + "]"
	}
	if port != "" {
		scheme := "http"
		if port == "443" || port == "8443" || port == "9443" {
			scheme = "https"
		}
		return []string{fmt.Sprintf("%s://%s:%s", scheme, hostForURL, port)}
	}
	return []string{"https://" + hostForURL, "http://" + hostForURL}
}

func (t Target) DefaultPorts() []string {
	if t.ExplicitPort != "" {
		return []string{t.ExplicitPort}
	}
	if t.ExplicitURL && len(t.CandidateURLs) > 0 {
		if u, err := url.Parse(t.CandidateURLs[0]); err == nil {
			if u.Port() != "" {
				return []string{u.Port()}
			}
			if strings.EqualFold(u.Scheme, "http") {
				return []string{"80"}
			}
		}
	}
	return []string{"443", "80"}
}
