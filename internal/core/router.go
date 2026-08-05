package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type RouterInfo struct {
	GatewayIP     string    `json:"gateway_ip"`
	AdminURL      string    `json:"admin_url,omitempty"`
	SSID          string    `json:"ssid,omitempty"`
	GatewayMAC    string    `json:"gateway_mac,omitempty"`
	HTTPStatus    string    `json:"http_status,omitempty"`
	PageTitle     string    `json:"page_title,omitempty"`
	ServerHeader  string    `json:"server_header,omitempty"`
	CheckedAt     time.Time `json:"checked_at"`
	HTTPS         bool      `json:"https"`
	Reachable     bool      `json:"reachable"`
	Notes         []string  `json:"notes,omitempty"`
	ProbeErrors   []string  `json:"probe_errors,omitempty"`
}

var (
	windowsRoutePattern = regexp.MustCompile(`(?m)^\s*0\.0\.0\.0\s+0\.0\.0\.0\s+(\d{1,3}(?:\.\d{1,3}){3})\s+`)
	ipRoutePattern      = regexp.MustCompile(`(?m)^default(?:\s+via)?\s+(\d{1,3}(?:\.\d{1,3}){3})\b`)
	unixRoutePattern    = regexp.MustCompile(`(?m)^\s*gateway:\s*([^\s]+)`)
	netstatRoutePattern = regexp.MustCompile(`(?m)^default\s+(\d{1,3}(?:\.\d{1,3}){3})\b`)
	macAddressPattern   = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b`)
	titlePattern        = regexp.MustCompile(`(?is)<title[^>]*>\s*(.*?)\s*</title>`)
)

func InspectRouter(ctx context.Context, timeout time.Duration) (RouterInfo, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	gateway, err := DetectDefaultGateway(ctx, timeout)
	if err != nil {
		return RouterInfo{}, err
	}
	info := RouterInfo{
		GatewayIP: gateway,
		SSID:      DetectCurrentSSID(ctx, timeout),
		GatewayMAC: DetectGatewayMAC(ctx, gateway, timeout),
		CheckedAt: time.Now(),
		Notes: []string{
			"PulseNet never reads, guesses, extracts, or submits router credentials.",
			"Your browser may offer credentials that you previously saved in its own password manager.",
		},
	}

	candidates := []string{"https://" + gateway, "http://" + gateway}
	for _, candidate := range candidates {
		probe, probeErr := probeRouterPage(ctx, candidate, timeout)
		if probeErr != nil {
			info.ProbeErrors = append(info.ProbeErrors, fmt.Sprintf("%s: %v", candidate, probeErr))
			continue
		}
		info.AdminURL = probe.AdminURL
		info.HTTPStatus = probe.HTTPStatus
		info.PageTitle = probe.PageTitle
		info.ServerHeader = probe.ServerHeader
		info.HTTPS = strings.HasPrefix(strings.ToLower(probe.AdminURL), "https://")
		info.Reachable = true
		return info, nil
	}

	// The browser may still reach a page that rejected the lightweight probe.
	info.AdminURL = "http://" + gateway
	info.Notes = append(info.Notes, "No admin page answered the lightweight probe; the detected gateway can still be opened manually.")
	return info, nil
}

func DetectDefaultGateway(ctx context.Context, timeout time.Duration) (string, error) {
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var attempts []struct {
		name string
		args []string
		parse func(string) string
	}
	switch runtime.GOOS {
	case "windows":
		attempts = []struct {
			name string
			args []string
			parse func(string) string
		}{
			{"route", []string{"print", "-4"}, parseWindowsGateway},
			{"netstat", []string{"-rn"}, parseNetstatGateway},
		}
	case "darwin":
		attempts = []struct {
			name string
			args []string
			parse func(string) string
		}{
			{"route", []string{"-n", "get", "default"}, parseUnixRouteGateway},
			{"netstat", []string{"-rn", "-f", "inet"}, parseNetstatGateway},
		}
	default:
		attempts = []struct {
			name string
			args []string
			parse func(string) string
		}{
			{"ip", []string{"route", "show", "default"}, parseIPRouteGateway},
			{"netstat", []string{"-rn"}, parseNetstatGateway},
		}
	}

	for _, attempt := range attempts {
		output, err := exec.CommandContext(commandCtx, attempt.name, attempt.args...).CombinedOutput()
		if err != nil && len(output) == 0 {
			continue
		}
		gateway := attempt.parse(string(output))
		if ip := net.ParseIP(gateway); ip != nil && ip.To4() != nil {
			return gateway, nil
		}
	}
	return "", fmt.Errorf("default IPv4 gateway could not be detected")
}

func DetectCurrentSSID(ctx context.Context, timeout time.Duration) string {
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		output, _ := exec.CommandContext(commandCtx, "netsh", "wlan", "show", "interfaces").CombinedOutput()
		re := regexp.MustCompile(`(?mi)^\s*SSID\s*:\s*(.+?)\s*$`)
		if match := re.FindStringSubmatch(string(output)); len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
	case "darwin":
		airport := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
		output, _ := exec.CommandContext(commandCtx, airport, "-I").CombinedOutput()
		re := regexp.MustCompile(`(?mi)^\s*SSID:\s*(.+?)\s*$`)
		if match := re.FindStringSubmatch(string(output)); len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
	default:
		if output, err := exec.CommandContext(commandCtx, "iwgetid", "-r").CombinedOutput(); err == nil {
			if value := strings.TrimSpace(string(output)); value != "" {
				return value
			}
		if output, err := exec.CommandContext(commandCtx, "nmcli", "-t", "-f", "active,ssid", "dev", "wifi").CombinedOutput(); err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				if strings.HasPrefix(line, "yes:") {
					return strings.TrimSpace(strings.TrimPrefix(line, "yes:"))
				}
			}
		}
	}
	return ""
}

func DetectGatewayMAC(ctx context.Context, gateway string, timeout time.Duration) string {
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var output []byte
	if runtime.GOOS == "windows" {
		output, _ = exec.CommandContext(commandCtx, "arp", "-a", gateway).CombinedOutput()
	} else {
		output, _ = exec.CommandContext(commandCtx, "arp", "-n", gateway).CombinedOutput()
	}
	return strings.ToUpper(macAddressPattern.FindString(string(output)))
}

type routerPageProbe struct {
	AdminURL     string
	HTTPStatus   string
	PageTitle    string
	ServerHeader string
}

func probeRouterPage(parent context.Context, rawURL string, timeout time.Duration) (routerPageProbe, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, // Router admin pages commonly use self-signed certificates; no credentials are sent.
		},
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return routerPageProbe{}, err
	}
	req.Header.Set("User-Agent", "PulseNet/2.4 router-assistant")
	resp, err := client.Do(req)
	if err != nil {
		return routerPageProbe{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	title := ""
	if match := titlePattern.FindSubmatch(body); len(match) == 2 {
		title = strings.Join(strings.Fields(html.UnescapeString(string(match[1]))), " ")
	}
	return routerPageProbe{
		AdminURL:     resp.Request.URL.String(),
		HTTPStatus:   resp.Status,
		PageTitle:    title,
		ServerHeader: resp.Header.Get("Server"),
	}, nil
}

func OpenBrowser(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid router admin URL")
	}

	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	case "darwin":
		command = exec.Command("open", rawURL)
	default:
		command = exec.Command("xdg-open", rawURL)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}

func parseWindowsGateway(output string) string {
	if match := windowsRoutePattern.FindStringSubmatch(output); len(match) == 2 {
		return match[1]
	}
	return ""
}

func parseIPRouteGateway(output string) string {
	if match := ipRoutePattern.FindStringSubmatch(output); len(match) == 2 {
		return match[1]
	}
	return ""
}

func parseUnixRouteGateway(output string) string {
	if match := unixRoutePattern.FindStringSubmatch(output); len(match) == 2 {
		return match[1]
	}
	return ""
}

func parseNetstatGateway(output string) string {
	if match := netstatRoutePattern.FindStringSubmatch(output); len(match) == 2 {
		return match[1]
	}
	return ""
}
