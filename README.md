<p align="center">
  <img src="assets/pulsenet.svg" alt="PulseNet" width="560">
</p>

<p align="center">
  A focused network diagnostics toolkit for websites, servers, DNS, TLS, HTTP, page snapshots, and local web-server logs.
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

Click **Download for Windows** above to get `PulseNet.exe`. No Go installation is required. Open it to launch the interactive terminal interface, or run it from PowerShell:

```powershell
.\PulseNet.exe
.\PulseNet.exe diagnose example.com
```

PulseNet is currently unsigned, so Windows SmartScreen may display an unknown-publisher warning. Verify the SHA-256 checksum attached to the release before running the file.

### Source code

Click **Source Code ZIP** when you want to inspect, modify, or build PulseNet for your own tasks. Linux and macOS binaries are available on the [Releases](https://github.com/appoloncel283-debug/pulsenet/releases) page.

## Why PulseNet

PulseNet follows the same path a real request takes and reports where it breaks:

**DNS → TCP → TLS → HTTP → browser security headers → recommendations**

It also includes two practical website-maintenance tools:

- **Site dump** saves one public HTTP response to a local folder with the body, redacted headers, metadata, redirects, and SHA-256 hash.
- **Log viewer** reads local Nginx, Apache, JSONL, and plain-text logs with filters and live follow mode.

PulseNet does not collect telemetry and does not silently upload reports, dumps, or logs.

## Features

- Full diagnosis with a health score, verdict, and actionable recommendations.
- DNS comparison across the system resolver, Cloudflare, and Google.
- DNS records: A, AAAA, CNAME, MX, NS, TXT, and PTR.
- TCP stability checks with success rate, latency, loss, and jitter.
- TLS inspection with protocol, cipher, ALPN, certificate chain, expiry, hostname validation, and OCSP stapling status.
- HTTP phase timings including DNS, connect, TLS, TTFB, and total request time.
- Security-header audit with grades and remediation notes.
- Explicit bounded TCP port checks.
- Controlled HTTP benchmark with p50/p90/p95/p99 latency and throughput.
- Availability monitoring with optional CSV logging.
- Public page snapshots with size limits and sensitive-header redaction.
- Local website-log viewer with filters, JSON output, and follow mode.
- Route tracing through the operating system tool.
- JSON and text diagnostic reports.
- Interactive terminal interface.

## Quick start

### Windows

```powershell
.\PulseNet.exe
.\PulseNet.exe diagnose example.com
.\PulseNet.exe dump https://example.com
.\PulseNet.exe logs C:\nginx\logs\access.log --status 5xx --follow
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
pulsenet headers <url>                     security-header audit
pulsenet ports <host> --ports <list>       explicit TCP port check
pulsenet benchmark <url>                   HTTP benchmark
pulsenet watch <url>                       availability monitor
pulsenet dump <url>                        save a public page snapshot
pulsenet logs <log-file>                   view, filter, and follow website logs
pulsenet trace <host>                      route trace
pulsenet support                           support address
```

### Site dump examples

```bash
# Save page body, headers, metadata, redirects, and SHA-256
pulsenet dump https://example.com

# Select the output directory and maximum body size
pulsenet dump https://example.com --output snapshots/example --max-mb 32

# Machine-readable result
pulsenet dump https://example.com --json
```

A dump stores one public response. It is deliberately not a crawler and does not copy an entire website or bypass authentication. `Set-Cookie`, `WWW-Authenticate`, and proxy-authentication response headers are redacted.

### Website log examples

```bash
# Show the last 100 lines
pulsenet logs /var/log/nginx/access.log

# Show and follow 5xx requests
pulsenet logs /var/log/nginx/access.log --status 5xx --follow

# Filter errors containing a specific upstream name
pulsenet logs /var/log/nginx/error.log --level error --contains upstream

# Filter a Windows Nginx log
pulsenet logs C:\nginx\logs\access.log --method POST --request-path /api --lines 250

# Structured output for scripts
pulsenet logs /var/log/nginx/access.log --status 400-599 --json
```

The log viewer reads files available on the current computer. To inspect production logs, run PulseNet on the server or securely copy the log file first; it does not attempt to obtain private server logs from a public website.

Detailed options are documented in [docs/COMMANDS.md](docs/COMMANDS.md).

## Safety and scope

The `ports` command checks one host and requires an explicit port list. It accepts at most 128 unique ports per run and limits concurrency. Use it only on systems you own or are authorized to test.

The `benchmark` command generates real HTTP traffic. Start with a small request count and use it only against services you control or have permission to test.

The `dump` command retrieves the public response returned to a normal unauthenticated GET request. Respect site terms, robots policies where applicable, copyright, and local law.

## Privacy

PulseNet stores data only when you explicitly request a report, CSV, site dump, or other output file. DNS comparison sends the queried hostname to the selected public resolvers. HTTP and TLS checks connect directly to the target you provide. Environment proxy settings are honored.

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
