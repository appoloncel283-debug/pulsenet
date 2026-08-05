# PulseNet command reference

## Network and website commands

### `diagnose`

```text
pulsenet diagnose <target> [--timeout 5s] [--attempts 3] [--ports 80,443] [--report report.txt] [--json report.json]
```

Runs DNS, TCP, TLS, HTTP, and security-header checks and generates a health score.

### `dump`

```text
pulsenet dump <url> [--output directory] [--max-mb 16] [--timeout 15s] [--json]
```

Saves one public HTTP response, redacted headers, metadata, redirects, and body SHA-256. It does not log in or obtain private server data.

### `logs`

```text
pulsenet logs <local-file> [--lines 100] [--follow] [--status 5xx] [--level error] [--contains text]
```

Reads a local Nginx, Apache, JSONL, or text log. Production logs must already be accessible on the computer running PulseNet.

### `site-diff`

```text
pulsenet site-diff --old <older-dump-directory> --new <newer-dump-directory> [--json]
```

Compares two PulseNet dump directories by body SHA-256, HTTP status, and saved headers.

### `site-secrets`

```text
pulsenet site-secrets --dump <directory> [--max-mb 32] [--json]
```

Scans local dump files for obvious private-key, cloud-key, GitHub-token, and generic secret markers. Output previews are redacted. Findings are hints and require manual review.

## Database commands

All database commands wrap installed official client tools. They require authorization and do not discover or bypass credentials.

### `db tools`

```text
pulsenet db tools [--json]
```

Detects `pg_dump`, `mysqldump`, and `sqlite3` and prints their versions.

### `db schema`

```text
pulsenet db schema --engine <postgres|mysql|sqlite> --database <name-or-file> --output <file> [--timeout 30m] [--arg <client-argument>] [--json]
```

Creates a schema-only export. `--arg` may be repeated to pass non-secret options to the official client.

### `db backup`

```text
pulsenet db backup --engine <postgres|mysql|sqlite> --database <name-or-file> --output <file> [--timeout 30m] [--arg <client-argument>] [--json]
```

- PostgreSQL uses `pg_dump --format=custom`.
- MySQL/MariaDB uses `mysqldump --single-transaction --quick` with routines, events, and triggers.
- SQLite uses the official `.backup` command.

Every successful export also writes `<output>.manifest.json` with size, SHA-256, client version, and timing.

### `db verify`

```text
pulsenet db verify --engine <postgres|mysql|sqlite> --file <backup> [--timeout 2m] [--json]
```

- PostgreSQL: validates the custom archive catalog with `pg_restore --list`.
- SQLite: runs `PRAGMA quick_check`.
- MySQL/MariaDB: checks for recognizable dump statements.
- All engines: calculate SHA-256 and report file size.

## Credential handling

Use the official client configuration, operating-system secret storage, or environment variables. Avoid passwords in command-line arguments because command history and process listings may expose them.

## Authorization

Run database, benchmark, port, and site-maintenance commands only on systems you own or are explicitly authorized to maintain.
