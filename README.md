<p align="center">
  <img src="assets/pulsenet.svg" alt="PulseNet" width="560">
</p>

<p align="center">
  Practical network diagnostics for websites, servers, DNS, TLS, HTTP, page snapshots, and local web-server logs.
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/releases/latest/download/Install-PulseNet.ps1"><img alt="Download PulseNet for Windows" src="https://img.shields.io/badge/Download_for_Windows-Installer-0078D4?style=for-the-badge&logo=windows"></a>
  <a href="https://github.com/appoloncel283-debug/pulsenet/archive/refs/heads/main.zip"><img alt="Download source code" src="https://img.shields.io/badge/Download-Source_Code_ZIP-181717?style=for-the-badge&logo=github"></a>
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="Source-available no-modification license" src="https://img.shields.io/badge/license-source--available%20%7C%20no%20modifications-orange.svg"></a>
  <img alt="Go 1.23+" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8.svg">
</p>

## Install on Windows

Click **Download for Windows** above. The button downloads the official guided installer.

Review the downloaded script, then run it from PowerShell:

```powershell
Unblock-File -LiteralPath .\Install-PulseNet.ps1
.\Install-PulseNet.ps1
```

The installer:

- resolves one exact GitHub release tag before downloading anything;
- downloads the executable, uninstaller, and checksum manifest from that same release;
- verifies SHA-256 before installing;
- installs only for the current Windows user;
- optionally adds `pulsenet` to the user PATH;
- optionally installs `pn`, `pncheck`, `pnlogs`, `pndump`, and `pnwatch`;
- optionally creates a Start Menu shortcut;
- registers an uninstaller in Windows Settings.

Silent installation with quick commands:

```powershell
.\Install-PulseNet.ps1 -Quiet -QuickCommands
```

Custom installation folder:

```powershell
.\Install-PulseNet.ps1 -InstallDirectory D:\Tools\PulseNet
```

A portable `PulseNet.exe`, packaged Windows builds, Linux builds, macOS builds, and `SHA256SUMS.txt` are available on the [Releases](https://github.com/appoloncel283-debug/pulsenet/releases) page.

PulseNet is currently unsigned, so Windows may display an unknown-publisher warning.

## Features

- Full DNS → TCP → TLS → HTTP diagnosis with a score and recommendations.
- DNS records and comparison between the system resolver, Cloudflare, and Google.
- TCP latency, loss, and jitter checks.
- TLS protocol, cipher, ALPN, certificate chain, expiry, hostname, and OCSP details.
- HTTP DNS/connect/TLS/TTFB/total timing breakdown.
- Browser security-header audit.
- Explicit bounded TCP port checks.
- Controlled HTTP benchmark with latency percentiles and throughput.
- Availability monitoring with optional CSV logging.
- Public page snapshots with size limits and sensitive-header redaction.
- Local Nginx, Apache, JSONL, and plain-text log viewing with filters and follow mode.
- Route tracing, JSON reports, text reports, and an interactive terminal interface.

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

Examples:

```powershell
pulsenet diagnose example.com --report report.txt --json report.json
pulsenet dump https://example.com --output snapshots\example
pulsenet logs C:\nginx\logs\access.log --status 5xx --follow
pulsenet ports 192.168.1.20 --ports 22,80,443,8000-8010
pulsenet benchmark https://example.com --requests 100 --concurrency 10
```

Detailed options are documented in [docs/COMMANDS.md](docs/COMMANDS.md).

## Build an unmodified copy

Go 1.23 or newer is required.

```bash
git clone https://github.com/appoloncel283-debug/pulsenet.git
cd pulsenet
go test ./...
go vet ./...
go build -o pulsenet ./cmd/pulsenet
```

## Safety and privacy

Use port checks and benchmarks only on systems you own or are authorized to test. The log viewer reads files already available on the current computer; it does not obtain private logs from a public website.

PulseNet does not collect telemetry and does not silently upload reports, dumps, or logs. DNS comparison sends the queried hostname to the selected public resolvers. HTTP and TLS checks connect to the target supplied by the user.

## Source availability and license

The source code is available for inspection, security review, education, evaluation, and compilation of an unmodified personal or internal copy.

Modification, derivative works, redistribution, repackaging, sublicensing, and commercial exploitation are not permitted without prior written permission. See the [PulseNet Source-Available No-Modification License 1.1](LICENSE).

## Support the project

USDT on the **TRON network (TRC20)**:

```text
TGVDhCbDKEnWV5BVrUtMicjhwMiJVUYSSh
```

<p>
  <img src="assets/usdt-trc20-qr.svg" alt="USDT TRC20 support QR code" width="220">
</p>

Always verify that the selected transfer network is **TRC20** before sending.
