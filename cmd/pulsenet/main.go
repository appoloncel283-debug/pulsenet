package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/appoloncel283-debug/pulsenet/internal/core"
	"github.com/appoloncel283-debug/pulsenet/internal/ui"
)

var version = "2.0.0"

func main() {
	if len(os.Args) == 1 {
		ui.RunMenu(version)
		return
	}

	var err error
	switch os.Args[1] {
	case "diagnose", "check":
		err = diagnoseCommand(os.Args[2:])
	case "dns":
		err = dnsCommand(os.Args[2:])
	case "tls", "cert":
		err = tlsCommand(os.Args[2:])
	case "headers", "security":
		err = headersCommand(os.Args[2:])
	case "ports", "scan":
		err = portsCommand(os.Args[2:])
	case "benchmark", "bench":
		err = benchmarkCommand(os.Args[2:])
	case "watch", "monitor":
		err = watchCommand(os.Args[2:])
	case "trace":
		err = traceCommand(os.Args[2:])
	case "support", "donate":
		ui.PrintSupport()
	case "version", "--version", "-v":
		fmt.Println("PulseNet", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		ui.PrintError(err)
		os.Exit(1)
	}
}

func diagnoseCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("diagnose", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 5*time.Second, "timeout per probe")
	attempts := fs.Int("attempts", 3, "TCP connection attempts per port")
	portsRaw := fs.String("ports", "", "comma-separated ports or ranges, for example 80,443,8000-8010")
	jsonPath := fs.String("json", "", "save a JSON report")
	textPath := fs.String("report", "", "save a text report")
	jsonOnly := fs.Bool("json-only", false, "print only JSON")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet diagnose <target> [options]")
	}
	var ports []string
	if *portsRaw != "" {
		parsed, err := core.ParsePorts(*portsRaw, 128)
		if err != nil {
			return err
		}
		ports = parsed
	}
	if *jsonOnly {
		target, err := core.ParseTarget(targetArg)
		if err != nil {
			return err
		}
		report := core.RunDiagnosis(context.Background(), target, core.Options{Timeout: *timeout, TCPAttempts: *attempts, Ports: ports, InsecureTLS: *insecure})
		ui.PrintJSON(report)
		return nil
	}
	return ui.RunDiagnose(targetArg, *timeout, *attempts, ports, *insecure, *jsonPath, *textPath)
}

func dnsCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("dns", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 5*time.Second, "DNS timeout")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet dns <domain-or-ip>")
	}
	ui.RunDNS(targetArg, *timeout, *jsonOutput)
	return nil
}

func tlsCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("tls", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 5*time.Second, "TLS timeout")
	insecure := fs.Bool("insecure", false, "skip certificate verification")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet tls <domain-or-host:port>")
	}
	ui.RunTLS(targetArg, *timeout, *insecure)
	return nil
}

func headersCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("headers", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 8*time.Second, "HTTP timeout")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet headers <url>")
	}
	ui.RunHeaders(targetArg, *timeout, *insecure, *jsonOutput)
	return nil
}

func portsCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("ports", flag.ContinueOnError)
	portsRaw := fs.String("ports", "", "required comma-separated ports or ranges")
	timeout := fs.Duration("timeout", 2*time.Second, "connection timeout")
	concurrency := fs.Int("concurrency", 32, "parallel connection limit (max 64)")
	showClosed := fs.Bool("show-closed", false, "include closed ports")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" || *portsRaw == "" {
		return fmt.Errorf("usage: pulsenet ports <host> --ports 22,80,443")
	}
	ui.RunPorts(targetArg, *portsRaw, *timeout, *concurrency, *showClosed, *jsonOutput)
	return nil
}

func benchmarkCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	requests := fs.Int("requests", 50, "number of requests (max 10000)")
	concurrency := fs.Int("concurrency", 5, "parallel requests (max 100)")
	method := fs.String("method", "GET", "GET or HEAD")
	timeout := fs.Duration("timeout", 10*time.Second, "request timeout")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet benchmark <url> [options]")
	}
	ui.RunBenchmark(targetArg, *method, *requests, *concurrency, *timeout, *insecure, *jsonOutput)
	return nil
}

func watchCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	interval := fs.Duration("interval", 5*time.Second, "time between checks")
	timeout := fs.Duration("timeout", 5*time.Second, "request timeout")
	count := fs.Int("count", 0, "stop after this many checks; zero means unlimited")
	csvPath := fs.String("csv", "", "write samples to a CSV file")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet watch <url> [options]")
	}
	if *interval < time.Second {
		return fmt.Errorf("interval must be at least 1 second")
	}
	ui.RunWatch(targetArg, *interval, *timeout, *count, *csvPath, *insecure)
	return nil
}

func traceCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("trace", flag.ContinueOnError)
	maxHops := fs.Int("max-hops", 20, "maximum number of hops")
	timeout := fs.Duration("timeout", 45*time.Second, "overall trace timeout")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet trace <host> [options]")
	}
	target, err := core.ParseTarget(targetArg)
	if err != nil {
		return err
	}
	output, err := core.TraceRoute(context.Background(), target.Host, *maxHops, *timeout)
	if output != "" {
		fmt.Print(output)
		if !strings.HasSuffix(output, "\n") {
			fmt.Println()
		}
	}
	return err
}

func splitLeadingTarget(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

func printHelp() {
	fmt.Printf(`PulseNet %s — practical network diagnostics in one binary

Usage:
  pulsenet                                  interactive interface
  pulsenet diagnose <target>                complete DNS/TCP/TLS/HTTP diagnosis
  pulsenet dns <domain-or-ip>               DNS records and resolver comparison
  pulsenet tls <domain-or-host:port>         certificate and TLS inspection
  pulsenet headers <url>                     browser security header audit
  pulsenet ports <host> --ports <list>       explicit TCP port check
  pulsenet benchmark <url>                   HTTP latency and throughput benchmark
  pulsenet watch <url>                       availability monitor with optional CSV
  pulsenet trace <host>                      route trace using the platform tool
  pulsenet support                           project support address

Examples:
  pulsenet diagnose example.com --report report.txt --json report.json
  pulsenet diagnose example.com --ports 80,443,8443 --attempts 5
  pulsenet dns example.com --json
  pulsenet headers https://example.com
  pulsenet ports 192.168.1.10 --ports 22,80,443,8000-8010
  pulsenet benchmark https://example.com --requests 100 --concurrency 10
  pulsenet watch https://example.com --interval 10s --csv uptime.csv

Run "pulsenet <command> -h" for command-specific flags.
`, version)
}
