# Changelog

## 2.4.0 — 2026-08-05

- Added one interactive Database Toolkit tab for client detection, schema exports, backups, and verification.
- Added executable SHA-256 verification on every startup.
- Added installed integrity manifests and an `integrity` command with text and JSON output.
- Added a local router assistant that detects the default gateway, Wi-Fi SSID, gateway MAC, and reachable admin page.
- Added safe browser opening through `router open` without reading, guessing, extracting, or submitting credentials.
- Added `pnrouter` and `pnsha` optional quick commands.
- Added integrity and router parser tests.
- Updated the Windows installer, uninstaller, release packaging, and documentation.

## 2.3.0 — 2026-08-05

- Added authorized PostgreSQL, MySQL/MariaDB, and SQLite database tooling.
- Added official-client detection with version reporting.
- Added schema-only exports and full backups with SHA-256 manifests.
- Added PostgreSQL archive, SQLite integrity, and MySQL dump verification.
- Added site-dump comparison for body, status, and header changes.
- Added local dump scanning for accidentally exposed keys and tokens with redacted previews.
- Added focused unit tests for database engine parsing, file hashing, and site-dump comparison.
- Added clearer uninstall and credential-handling documentation.

## 2.2.1 — 2026-08-05

- Fixed SHA-256 installation failures by resolving one exact release tag and downloading every verified asset from that immutable release.
- Reworked CI and release validation.
- Clarified the source-available no-modification license.

## 2.2.0 — 2026-08-05

- Added a guided current-user PowerShell installer and uninstaller.
- Added optional PATH, Start Menu, and quick-command setup.
- Removed repository publishing helper scripts.

## 2.1.0 — 2026-08-05

- Added public page dumps and local website-log viewing.

## 2.0.0 — 2026-08-05

- Added the expanded network diagnostics toolkit and release automation.

## 1.0.0

- Initial DNS, TCP, TLS, HTTP, watch, and report functionality.
