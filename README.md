<p align="center">
  <img src="assets/pulsenet.svg" alt="PulseNet" width="560">
</p>

<p align="center">
  Practical network, website-maintenance, log-analysis, and authorized database-backup tools in one Go binary.
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

Click **Download for Windows**, review the downloaded script, then run:

```powershell
Unblock-File -LiteralPath .\Install-PulseNet.ps1
.\Install-PulseNet.ps1
```

The installer pins one exact release, verifies SHA-256, installs for the current user, and can add `pulsenet` plus optional quick commands to PATH.

Silent installation with quick commands:

```powershell
.\Install-PulseNet.ps1 -Quiet -QuickCommands
```

## Uninstall

Use **Windows Settings → Apps → Installed apps → PulseNet → Uninstall**.

Or run:

```powershell
& "$env:LOCALAPPDATA\Programs\PulseNet\Uninstall-PulseNet.ps1"
```

For the portable build, simply delete `PulseNet.exe` and any reports or dumps you created.

## Main capabilities

- DNS, TCP, TLS, HTTP, security-header, port, benchmark, trace, and uptime diagnostics.
- Public page snapshots with redacted sensitive response headers.
- Local Nginx, Apache, JSONL, and text log viewing with filters and follow mode.
- Site-dump comparison by body hash, HTTP status, and headers.
- Local dump scanning for accidentally exposed keys or tokens; previews are redacted.
- Authorized PostgreSQL, MySQL/MariaDB, and SQLite schema exports and backups.
- Backup verification with SHA-256 plus PostgreSQL archive, SQLite integrity, or MySQL dump checks.
- JSON/text reports and an interactive terminal interface.

## Database toolkit

PulseNet wraps official client tools and never bypasses authentication. Install the database client for the engine you use, then check availability:

```powershell
pulsenet db tools
```

SQLite:

```powershell
pulsenet db schema --engine sqlite --database .\app.db --output .\backups\schema.sql
pulsenet db backup --engine sqlite --database .\app.db --output .\backups\app.db
pulsenet db verify --engine sqlite --file .\backups\app.db
```

PostgreSQL:

```powershell
pulsenet db schema --engine postgres --database "postgresql://user@localhost/app" --output .\backups\schema.sql
pulsenet db backup --engine postgres --database "postgresql://user@localhost/app" --output .\backups\app.dump
pulsenet db verify --engine postgres --file .\backups\app.dump
```

MySQL/MariaDB:

```powershell
pulsenet db schema --engine mysql --database app --output .\backups\schema.sql --arg=--host=127.0.0.1 --arg=--user=backup
pulsenet db backup --engine mysql --database app --output .\backups\app.sql --arg=--host=127.0.0.1 --arg=--user=backup
pulsenet db verify --engine mysql --file .\backups\app.sql
```

Use environment variables or the official client configuration for credentials. Do not put passwords directly in command history.

## Website maintenance tools

Create two snapshots and compare them:

```powershell
pulsenet dump https://example.com --output .\dumps\before
pulsenet dump https://example.com --output .\dumps\after
pulsenet site-diff --old .\dumps\before --new .\dumps\after
```

Check a local dump before publishing or sharing it:

```powershell
pulsenet site-secrets --dump .\dumps\after
```

This scans only local files and reports redacted previews. It does not retrieve private server data.

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
pulsenet logs <log-file>                   view/filter/follow local website logs
pulsenet site-diff                         compare two local dump directories
pulsenet site-secrets                      scan a local dump for exposed secrets
pulsenet db tools                          detect official database clients
pulsenet db schema                         authorized schema-only export
pulsenet db backup                         authorized database backup
pulsenet db verify                         verify a backup or dump
pulsenet trace <host>                      route trace
pulsenet support                           support address
```

Detailed options are in [docs/COMMANDS.md](docs/COMMANDS.md).

## Safety and privacy

Use network checks, benchmarks, site tools, and database operations only on systems you own or are explicitly authorized to maintain. Database commands require a local database file or credentials accepted by the official client. PulseNet does not contain authentication bypasses or credential-discovery features.

PulseNet does not collect telemetry or silently upload reports, logs, dumps, or backups.

## Source availability and license

The source is available for inspection, security review, education, evaluation, and compilation of an unmodified personal or internal copy. Modification, derivative works, redistribution, repackaging, sublicensing, and commercial exploitation require prior written permission. See [LICENSE](LICENSE).

## Support the project

USDT on the **TRON network (TRC20)**:

```text
TGVDhCbDKEnWV5BVrUtMicjhwMiJVUYSSh
```

<p>
  <img src="assets/usdt-trc20-qr.svg" alt="USDT TRC20 support QR code" width="220">
</p>

Always verify that the selected transfer network is **TRC20** before sending.
