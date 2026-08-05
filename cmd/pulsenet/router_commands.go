package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
	"github.com/appoloncel283-debug/pulsenet/internal/ui"
)

func routerCommand(args []string) error {
	subcommand := "info"
	if len(args) > 0 && args[0] != "--json" && args[0] != "-h" && args[0] != "--help" {
		subcommand = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("router "+subcommand, flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "print JSON")
	address := fs.String("address", "", "router IP address or HTTP/HTTPS URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	switch subcommand {
	case "info", "detect":
		if *address != "" || fs.NArg() > 0 {
			return fmt.Errorf("router info detects the active gateway automatically; use router open for a specific address")
		}
		return ui.RunRouterAssistant(false, *jsonOutput)
	case "open":
		if *jsonOutput {
			return fmt.Errorf("--json cannot be combined with router open")
		}
		if fs.NArg() > 1 {
			return fmt.Errorf("usage: pulsenet router open [IP-or-URL]")
		}

		rawAddress := strings.TrimSpace(*address)
		if rawAddress == "" && fs.NArg() == 1 {
			rawAddress = strings.TrimSpace(fs.Arg(0))
		}
		if rawAddress == "" {
			return ui.RunRouterAssistant(true, false)
		}

		normalizedAddress, err := normalizeRouterAddress(rawAddress)
		if err != nil {
			return err
		}
		if err := core.OpenBrowser(normalizedAddress); err != nil {
			return err
		}
		fmt.Printf("Router admin page opened: %s\n", normalizedAddress)
		return nil
	default:
		return fmt.Errorf("usage: pulsenet router <info|open> [IP-or-URL] [--json]")
	}
}

func normalizeRouterAddress(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("router address cannot be empty")
	}
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid router address: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("router address must use http or https")
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("router address must include an IP address or hostname")
	}
	return parsed.String(), nil
}

func integrityCommand(args []string, status core.IntegrityStatus) error {
	fs := flag.NewFlagSet("integrity", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ui.PrintIntegrityStatus(status, *jsonOutput)
	if status.State == "mismatch" || status.State == "error" {
		return fmt.Errorf("integrity state is %s", status.State)
	}
	return nil
}
