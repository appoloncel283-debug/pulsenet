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

var version = "2.4.0"

func main() {
	integrityStatus := core.VerifySelfIntegrity()
	ui.PrintStartupIntegrity(integrityStatus)

	if len(os.Args) == 1 {
		ui.RunMenu(version, integrityStatus)
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
	case "dump", "site-dump":
		err = dumpCommand(os.Args[2:])
	case "logs", "log-viewer":
		err = logsCommand(os.Args[2:])
	case "site-diff":
		err = siteDiffCommand(os.Args[2:])
	case "site-secrets":
		err = siteSecretsCommand(os.Args[2:])
	case "db", "database":
		err = databaseCommand(os.Args[2:])
	case "router":
		err = routerCommand(os.Args[2:])
	case "integrity", "sha256":
		err = integrityCommand(os.Args[2:], integrityStatus)
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

func dumpCommand(args []string) error {
	targetArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("dump", flag.ContinueOnError)
	output := fs.String("output", "", "output directory; default is generated from host and time")
	maxMB := fs.Int64("max-mb", 16, "maximum response body size in MiB (max 128)")
	timeout := fs.Duration("timeout", 15*time.Second, "request timeout")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	jsonOutput := fs.Bool("json", false, "print JSON result")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if targetArg == "" && fs.NArg() > 0 {
		targetArg = fs.Arg(0)
	}
	if targetArg == "" {
		return fmt.Errorf("usage: pulsenet dump <url> [options]")
	}
	if *maxMB < 1 || *maxMB > 128 {
		return fmt.Errorf("max-mb must be between 1 and 128")
	}
	return ui.RunSiteDump(targetArg, *output, *maxMB*1024*1024, *timeout, *insecure, *jsonOutput)
}

func logsCommand(args []string) error {
	fileArg, flagArgs := splitLeadingTarget(args)
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	lines := fs.Int("lines", 100, "number of recent matching lines to show")
	follow := fs.Bool("follow", false, "continue showing appended log lines")
	interval := fs.Duration("interval", 500*time.Millisecond, "poll interval in follow mode")
	contains := fs.String("contains", "", "case-insensitive text filter")
	level := fs.String("level", "", "log level filter, for example error or warn")
	status := fs.String("status", "", "HTTP status filter: 500, 500-599, or 5xx")
	ip := fs.String("ip", "", "client IP filter")
	method := fs.String("method", "", "HTTP method filter")
	requestPath := fs.String("request-path", "", "request path substring filter")
	jsonOutput := fs.Bool("json", false, "print JSON; unavailable with --follow")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if fileArg == "" && fs.NArg() > 0 {
		fileArg = fs.Arg(0)
	}
	if fileArg == "" {
		return fmt.Errorf("usage: pulsenet logs <log-file> [options]")
	}
	if *lines < 1 || *lines > 100000 {
		return fmt.Errorf("lines must be between 1 and 100000")
	}
	if *jsonOutput && *follow {
		return fmt.Errorf("--json cannot be combined with --follow")
	}
	filter := core.LogFilter{Contains: *contains, Level: *level, Status: *status, IP: *ip, Method: *method, RequestPath: *requestPath}
	return ui.RunLogs(fileArg, *lines, *follow, *interval, filter, *jsonOutput)
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
	fmt.Printf(`PulseNet %s — network, website, database, and router maintenance tools

Usage:
  pulsenet                                  interactive interface
  pulsenet diagnose <target>                complete DNS/TCP/TLS/HTTP diagnosis
  pulsenet dns <domain-or-ip>               DNS records and resolver comparison
  pulsenet tls <domain-or-host:port>         certificate and TLS inspection
  pulsenet headers <url>                     browser security header audit
  pulsenet ports <host> --ports <list>       explicit TCP port check
  pulsenet benchmark <url>                   HTTP latency and throughput benchmark
  pulsenet watch <url>                       availability monitor with optional CSV
  pulsenet dump <url>                        save a public page snapshot to disk
  pulsenet logs <log-file>                   view, filter, and follow local website logs
  pulsenet site-diff                         compare two local site dumps
  pulsenet site-secrets                      scan a local dump for exposed secrets
  pulsenet db <subcommand>                   authorized database backup toolkit
  pulsenet router info                       detect the local gateway and admin page
  pulsenet router open                       open the detected admin page in a browser
  pulsenet integrity                         show executable SHA-256 verification
  pulsenet trace <host>                      route trace using the platform tool
  pulsenet support                           project support address

Database examples:
  pulsenet db tools
  pulsenet db schema --engine sqlite --database app.db --output schema.sql
  pulsenet db backup --engine postgres --database postgresql://user@localhost/app --output app.dump
  pulsenet db verify --engine postgres --file app.dump

Router assistant:
  pulsenet router info
  pulsenet router open

PulseNet does not read, guess, extract, or submit router passwords. The browser may use credentials already saved in its own password manager.

Run "pulsenet <command> -h" for command-specific flags.
`, version)
}
