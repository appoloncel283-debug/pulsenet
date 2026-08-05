<p align="center">
  <img src="assets/pulsenet.svg" alt="PulseNet" width="560">
</p>

<p align="center">
  A focused network diagnostics toolkit for websites, servers, DNS, TLS, HTTP, and explicit TCP port checks.
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/releases/download/latest/PulseNet.exe"><img alt="Download PulseNet for Windows" src="https://img.shields.io/badge/Download_for_Windows-PulseNet.exe-0078D4?style=for-the-badge&logo=windows"></a>
  <a href="https://github.com/appoloncel283-debug/pulsenet/archive/refs/heads/main.zip"><img alt="Download source code" src="https://img.shields.io/badge/Download-Source_Code_ZIP-181717?style=for-the-badge&logo=github"></a>
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
  <img alt="Go 1.23+" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8.svg">
</p>

## Download

### Windows application

Click **Download for Windows** above to get `PulseNet.exe`. No Go installation is required. Open the downloaded file to launch the interactive terminal interface, or run it from PowerShell:

```powershell
.\PulseNet.exe
.\PulseNet.exe diagnose example.com
```

PulseNet is currently distributed as an unsigned executable, so Windows SmartScreen may show an unknown-publisher warning. Verify the file checksum from the release before running it.

### Source code

Click **Source Code ZIP** above when you want to inspect, modify, or build PulseNet for your own tasks. The repository contains the complete Go source, tests, documentation, and build workflows.

Linux and macOS binaries are available on the [Releases](https://github.com/appoloncel283-debug/pulsenet/releases) page.

## Why PulseNet

PulseNet follows the same path a real request takes and reports where it breaks:

**DNS → TCP → TLS → HTTP → browser security headers → recommendations**

It is designed for practical troubleshooting, support reports, lightweight monitoring, and targeted server checks. It does not collect telemetry and does not silently upload reports.

## Features

- **Full diagnosis** with a health score, verdict, and actionable recommendations.
- **DNS comparison** across the system resolver, Cloudflare, and Google.
- **DNS record lookup** for A, AAAA, CNAME, MX, NS, TXT, and PTR records.
- **TCP stability checks** with success rate, min/average/max latency, and jitter.
- **TLS inspection** with protocol, cipher suite, ALPN, certificate chain, expiry, hostname validation, and OCSP stapling status.
- **HTTP timing breakdown** including DNS, connect, TLS, TTFB, and total request time.
- **Security header audit** with a grade and focused remediation notes.
- **Explicit port checks** for a single host and a user-supplied port list or range.
- **HTTP benchmark** with throughput, success rate, p50/p90/p95/p99 latency, status distribution, and error grouping.
- **Availability monitor** with uptime statistics and optional CSV logging.
- **Route trace wrapper** using `tracert` on Windows or `traceroute` on Unix-like systems.
- **JSON and text reports** suitable for automation and support tickets.
- **Interactive terminal interface** when launched without arguments.

## Quick start

### Windows

Download `PulseNet.exe` using the button at the top of this page, or build it locally:

```powershell
build.bat
.\bin\pulsenet-windows-amd64.exe diagnose example.com
```

### Linux

Download the Linux binary from the GitHub Releases page, or build it locally:

```bash
./build.sh
./bin/pulsenet-linux-amd64 diagnose example.com
```

### Build from source

Go 1.23 or newer is required.

```bash
git clone https://github.com/appoloncel283-debug/pulsenet.git
cd pulsenet
go test ./...
go build -o pulsenet ./cmd/pulsenet
```

## Commands

```text
pulsenet                                  interactive interface
pulsenet diagnose <target>                full diagnosis
pulsenet dns <domain-or-ip>               DNS records and resolver comparison
pulsenet tls <domain-or-host:port>         TLS and certificate inspection
pulsenet headers <url>                     security header audit
pulsenet ports <host> --ports <list>       explicit TCP port check
pulsenet benchmark <url>                   HTTP benchmark
pulsenet watch <url>                       availability monitor
pulsenet trace <host>                      route trace
pulsenet support                           support address
```

### Useful examples

```bash
# Full report for a website
pulsenet diagnose https://example.com --report report.txt --json report.json

# Diagnose custom service ports
pulsenet diagnose api.example.com --ports 443,8443 --attempts 5 --timeout 4s

# Inspect DNS records and resolver differences
pulsenet dns example.com

# Audit browser-facing security headers
pulsenet headers https://example.com

# Check only ports you explicitly select
pulsenet ports 192.168.1.20 --ports 22,80,443,8000-8010

# Run a controlled HTTP benchmark
pulsenet benchmark https://example.com --requests 100 --concurrency 10

# Monitor availability and save samples
pulsenet watch https://example.com --interval 10s --csv uptime.csv
```

Detailed command documentation is available in [docs/COMMANDS.md](docs/COMMANDS.md).

## Safety and scope

The `ports` command checks one host and requires an explicit port list. It accepts at most 128 unique ports per run and limits concurrency. Use it only on systems you own or are authorized to test.

The `benchmark` command generates real HTTP traffic. Start with a small request count and use it only against services you control or have permission to test.

## Privacy

PulseNet stores data only when you explicitly request a report or CSV file. DNS comparison sends the queried hostname to the selected public resolvers. HTTP and TLS checks connect directly to the target you provide. Environment proxy settings are honored.

## Support the project

USDT on the **TRON network (TRC20)**:

```text
TGVDhCbDKEnWV5BVrUtMicjhwMiJVUYSSh
```

<p>
  <img src="assets/usdt-trc20-qr.svg" alt="USDT TRC20 support QR code" width="220">
</p>

Always verify that the selected transfer network is **TRC20** before sending.

## Contributing

Bug reports and focused pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting changes.

## License

PulseNet is released under the [MIT License](LICENSE).
