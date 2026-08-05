<p align="center">
  <img src="assets/pulsenet.svg" alt="PulseNet" width="560">
</p>

<p align="center">
  Practical network, website-maintenance, authorized database-backup, integrity, and local router tools in one Go binary.
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/raw/refs/heads/main/Install-PulseNet.ps1"><img alt="Download PulseNet for Windows" src="https://img.shields.io/badge/Download_for_Windows-Installer-0078D4?style=for-the-badge&logo=windows"></a>
  <a href="https://github.com/appoloncel283-debug/pulsenet/archive/refs/heads/main.zip"><img alt="Download source code" src="https://img.shields.io/badge/Download-Source_Code_ZIP-181717?style=for-the-badge&logo=github"></a>
</p>

<p align="center">
  <a href="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/appoloncel283-debug/pulsenet/actions/workflows/ci.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="Source-available no-modification license" src="https://img.shields.io/badge/license-source--available%20%7C%20no%20modifications-orange.svg"></a>
  <img alt="Go 1.23+" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8.svg">
</p>

## Install on Windows

Click **Download for Windows**, review the downloaded script, then run:

```powershell
Unblock-File -LiteralPath .\Install-PulseNet.ps1
.\Install-PulseNet.ps1
```

The download button points to the current installer source, not an installer copied into an older release. The installer refuses PulseNet releases older than 2.4.0, pins one exact immutable release, verifies SHA-256, installs the executable, writes an integrity manifest, and runs command smoke tests before reporting success.

Quick commands are enabled by default and include `pn`, `pncheck`, `pnlogs`, `pndump`, `pnwatch`, `pnrouter`, and `pnsha`. Disable them with:

```powershell
.\Install-PulseNet.ps1 -NoQuickCommands
```

Silent installation:

```powershell
.\Install-PulseNet.ps1 -Quiet -QuickCommands -Force
```

### Repair an older installation

First check the installed version:

```powershell
pulsenet version
```

If it is older than 2.4.0, download the installer again from **Download for Windows** and run:

```powershell
Unblock-File -LiteralPath .\Install-PulseNet.ps1
.\Install-PulseNet.ps1 -Force -QuickCommands
```

Close and reopen PowerShell after installation, then verify:

```powershell
pulsenet version
pulsenet integrity
pnsha
pnrouter
```

## Uninstall

Use **Windows Settings → Apps → Installed apps → PulseNet → Uninstall**.

Or run:

```powershell
& "$env:LOCALAPPDATA\Programs\PulseNet\Uninstall-PulseNet.ps1"
```

For the portable build, delete `PulseNet.exe`, `integrity.json`, and any reports or dumps you created.

## Startup integrity verification

PulseNet calculates its own SHA-256 every time it starts.

- Installed builds compare the executable with `integrity.json` and report `verified` or `mismatch`.
- Portable or development builds still display the current executable SHA-256, but report that no installed manifest is available.
- The integrity message is written to stderr, so JSON command output remains valid.

Show full details at any time:

```powershell
pulsenet integrity
pulsenet integrity --json
```

## Main capabilities

- DNS, TCP, TLS, HTTP, security-header, port, benchmark, trace, and uptime diagnostics.
- Public page snapshots with redacted sensitive response headers.
- Local Nginx, Apache, JSONL, and text log viewing with filters and follow mode.
- Site-dump comparison and local dump scanning for accidentally exposed secrets.
- Authorized PostgreSQL, MySQL/MariaDB, and SQLite schema exports, backups, manifests, and verification.
- Local router gateway detection and safe browser opening.
- JSON/text reports and an interactive terminal interface.

## Database toolkit

The interactive interface has one **Database toolkit** section for client detection, schema export, backup, and verification.

PulseNet wraps official database clients and never bypasses authentication:

```powershell
pulsenet db tools
pulsenet db schema --engine sqlite --database .\app.db --output .\backups\schema.sql
pulsenet db backup --engine postgres --database "postgresql://user@localhost/app" --output .\backups\app.dump
pulsenet db verify --engine postgres --file .\backups\app.dump
```

Use environment variables or official client configuration for credentials. Avoid passwords in command-line arguments because command history and process listings may expose them.

## Router assistant

PulseNet can detect the current default gateway, Wi-Fi name, gateway MAC address, and a reachable local router admin page:

```powershell
pulsenet router info
pulsenet router open
```

The router assistant opens the detected admin page in the default browser. It does **not** read, guess, extract, or submit router usernames or passwords. A browser may offer credentials that you previously saved in its own password manager.

For forgotten router credentials, use the label on the router, the ISP application or documentation, or the manufacturer's supported reset process.

## Website maintenance tools

```powershell
pulsenet dump https://example.com --output .\dumps\before
pulsenet dump https://example.com --output .\dumps\after
pulsenet site-diff --old .\dumps\before --new .\dumps\after
pulsenet site-secrets --dump .\dumps\after
```

`site-secrets` scans only local files and prints redacted previews. It does not retrieve private server data.

## Command overview

```text
pulsenet diagnose <target>                full DNS/TCP/TLS/HTTP diagnosis
pulsenet dns <domain-or-ip>               DNS records and resolver comparison
pulsenet tls <domain-or-host:port>         TLS and certificate inspection
pulsenet headers <url>                     security-header audit
pulsenet ports <host> --ports <list>       explicit TCP port check
pulsenet benchmark <url>                   controlled HTTP benchmark
pulsenet watch <url>                       availability monitor
pulsenet dump <url>                        save a public page snapshot
pulsenet logs <log-file>                   view and follow local website logs
pulsenet site-diff                         compare two local site dumps
pulsenet site-secrets                      scan one local dump
pulsenet db <subcommand>                   authorized database toolkit
pulsenet router <info|open>                local router assistant
pulsenet integrity                         executable SHA-256 verification
pulsenet trace <host>                      route trace
pulsenet support                           support address
```

Detailed options are documented in [docs/COMMANDS.md](docs/COMMANDS.md).

## Safety and privacy

Use port checks, benchmarks, database operations, and site-maintenance commands only on systems you own or are explicitly authorized to maintain.

PulseNet does not collect telemetry and does not silently upload reports, dumps, logs, credentials, or router information.

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
