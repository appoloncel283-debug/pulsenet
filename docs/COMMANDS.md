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

Compares address resolution across the system resolver, Cloudflare, and Google. It also requests common records through the system resolver.

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

The output includes TLS version, cipher suite, ALPN, certificate subject and issuer, validity dates, DNS names, chain length, and OCSP stapling presence.

## `headers`

Audits browser-facing response headers and assigns a grade.

```text
pulsenet headers https://example.com
pulsenet headers https://example.com --json
```

The audit checks HSTS, Content Security Policy, MIME sniffing protection, framing restrictions, Referrer Policy, Permissions Policy, Cross-Origin Opener Policy, and observed cookie flags.

## `ports`

Checks an explicit list of TCP ports on one host.

```text
pulsenet ports server.example.com --ports 22,80,443
pulsenet ports 192.168.1.20 --ports 8000-8010 --show-closed
```

Limits:

- One host per run.
- Maximum 128 unique ports.
- Maximum concurrency of 64.

## `benchmark`

Runs a controlled HTTP benchmark.

```text
pulsenet benchmark https://example.com --requests 100 --concurrency 10
```

Flags:

- `--requests` — total requests, maximum 10,000.
- `--concurrency` — parallel requests, maximum 100.
- `--method GET|HEAD` — request method.
- `--timeout` — per-request timeout.
- `--json` — structured output.
- `--insecure` — skip certificate verification.

Results include request rate, success rate, latency percentiles, status codes, and grouped errors.

## `watch`

Monitors a website repeatedly.

```text
pulsenet watch https://example.com --interval 10s
pulsenet watch https://example.com --count 60 --csv uptime.csv
```

A response is considered up when the request succeeds and the status code is below 500.

## `trace`

Runs the operating system route trace utility.

```text
pulsenet trace example.com --max-hops 20 --timeout 45s
```

Windows uses `tracert`. Linux and macOS use `traceroute`, which may need to be installed separately.

## `support`

Prints the project support address.

```text
pulsenet support
```
