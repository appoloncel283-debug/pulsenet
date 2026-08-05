# Changelog

## 2.2.0 — 2026-08-05

- Added a guided PowerShell installer for current-user Windows installations.
- Added SHA-256 verification before the installer accepts downloaded release files.
- Added optional user PATH setup, Start Menu shortcut creation, and quick commands: `pn`, `pncheck`, `pnlogs`, `pndump`, and `pnwatch`.
- Added a Windows Settings uninstaller and a standalone PowerShell uninstaller.
- Removed `publish-github.bat` and `publish-github.sh` from the repository.
- Replaced the MIT license with a source-available license that does not permit modification or derivative works without prior written permission.
- Updated release packaging to publish installer scripts and their checksums.

## 2.1.0 — 2026-08-05

- Added `dump` / `site-dump` for saving a public page body, redacted response headers, metadata, redirects, and SHA-256.
- Added response-size limits, generated output directories, content-aware file names, and truncation reporting.
- Added `logs` / `log-viewer` for local Nginx, Apache, JSONL, and plain-text logs.
- Added log tailing, live follow mode, JSON output, and filters for text, level, HTTP status, IP, method, and request path.
- Added interactive-menu entries and documentation for site dumps and website logs.

## 2.0.0 — 2026-08-05

- Reworked the entire interface and documentation in English.
- Added DNS record lookup and multi-resolver comparison.
- Added HTTP phase timings and selected response metadata.
- Expanded TLS details with ALPN, chain length, serial number, and OCSP stapling status.
- Added a security-header audit with grades and remediation notes.
- Added explicit bounded TCP port checks.
- Added an HTTP benchmark with percentiles and throughput.
- Added availability monitoring with CSV output.
- Added a platform route-trace wrapper.
- Added CI, automated release builds, issue templates, and project documentation.

## 1.0.0

- Initial DNS, TCP, TLS, HTTP, watch, and report functionality.
