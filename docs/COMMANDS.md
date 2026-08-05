# PulseNet command reference

## `diagnose`

Runs DNS, TCP, TLS, HTTP, and security-header checks and combines the results into a health score.

```text
pulsenet diagnose <target> [flags]
```

Flags:

- `--timeout 5s` — timeout for each probe.
- `--attempts 3` — TCP connection attempts per port.
- `--ports 80,443,8443` — override the default ports.
- `--report report.txt` — save a human-readable report.
- `--json report.json` — save a structured JSON report.
- `--json-only` — print only JSON to standard output.
- `--insecure` — skip certificate verification for troubleshooting only.

## `dns`

Compares address resolution across the system resolver, Cloudflare, and Google and requests common records through the system resolver.

```text
pulsenet dns example.com
pulsenet dns 1.1.1.1 --json
```

## `tls`

Inspects a TLS endpoint and its leaf certificate.

```text
pulsenet tls example.com
pulsenet tls example.com:8443
```

## `headers`

Audits browser-facing response headers and assigns a grade.

```text
pulsenet headers https://example.com
pulsenet headers https://example.com --json
```

## `ports`

Checks an explicit list of TCP ports on one host.

```text
pulsenet ports server.example.com --ports 22,80,443
pulsenet ports 192.168.1.20 --ports 8000-8010 --show-closed
```

Limits: one host, 128 unique ports, and concurrency capped at 64.

## `benchmark`

Runs a controlled HTTP benchmark.

```text
pulsenet benchmark https://example.com --requests 100 --concurrency 10
```

Results include request rate, success rate, latency percentiles, status codes, and grouped errors.

## `watch`

Monitors a website repeatedly.

```text
pulsenet watch https://example.com --interval 10s
pulsenet watch https://example.com --count 60 --csv uptime.csv
```

A response is considered up when the request succeeds and the status code is below 500.

## `dump`

Saves a single public HTTP response to a local directory.

```text
pulsenet dump <url> [flags]
```

Files produced:

- `page.html`, `body.json`, `body.txt`, or `body.bin` — response body.
- `headers.json` — response headers with sensitive authentication and cookie headers redacted.
- `metadata.json` — capture time, URLs, status, protocol, content type, body size, redirects, truncation state, and SHA-256.

Flags:

- `--output <directory>` — destination directory. The default is `site-dumps/<host>-<timestamp>`.
- `--max-mb 16` — maximum saved response-body size, from 1 to 128 MiB.
- `--timeout 15s` — request timeout.
- `--json` — print the result as JSON.
- `--insecure` — skip TLS certificate validation for troubleshooting only.

Examples:

```text
pulsenet dump https://example.com
pulsenet dump https://example.com --output snapshots/example --max-mb 32
```

The command performs a normal unauthenticated GET request. It is not a crawler, does not download an entire website, and does not bypass authentication.

## `logs`

Views and filters a local website or application log file. Common Nginx/Apache combined logs, JSONL logs, and plain text are recognized.

```text
pulsenet logs <log-file> [flags]
```

Flags:

- `--lines 100` — number of recent matching entries to display.
- `--follow` — continue displaying appended lines until Ctrl+C.
- `--interval 500ms` — follow-mode polling interval.
- `--contains <text>` — case-insensitive raw-line search.
- `--level error` — level filter such as `error`, `warn`, or `info`.
- `--status 5xx` — exact status, range, or class: `500`, `500-599`, or `5xx`.
- `--ip <value>` — client-IP substring filter.
- `--method POST` — HTTP method filter.
- `--request-path /api` — request-path substring filter.
- `--json` — structured output; cannot be combined with `--follow`.

Examples:

```text
pulsenet logs /var/log/nginx/access.log --status 5xx --follow
pulsenet logs /var/log/nginx/error.log --level error --contains upstream --lines 200
pulsenet logs C:\nginx\logs\access.log --method POST --request-path /api
```

For large files, the viewer scans the most recent 64 MiB. It reads only files available on the current machine. Run PulseNet on the server or securely copy the logs first when diagnosing a remote deployment.

## `trace`

Runs the operating system route-trace utility.

```text
pulsenet trace example.com --max-hops 20 --timeout 45s
```

Windows uses `tracert`. Linux and macOS use `traceroute`, which may need to be installed separately.

## `support`

Prints the project support address.

```text
pulsenet support
```
