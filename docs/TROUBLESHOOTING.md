# Troubleshooting guide

## Public DNS works, system DNS fails

This usually points to a local resolver problem rather than a remote service outage.

Check:

- VPN or proxy software.
- Router DNS cache.
- Manually configured DNS servers.
- Split-DNS rules on managed networks.
- Local filtering or endpoint security software.

## DNS works, TCP fails

The hostname resolves, but the selected service port cannot be reached.

Check:

- Whether the service is running.
- Host firewall rules.
- Cloud security groups or network ACLs.
- Port forwarding on a router.
- ISP or corporate network filtering.

## TCP works, TLS fails

A connection is accepted, but TLS negotiation or certificate validation fails.

Check:

- Certificate hostname coverage.
- Missing intermediate certificates.
- Expired certificates.
- Incorrect system time.
- TLS inspection by antivirus, proxies, or corporate gateways.
- Unsupported TLS versions or cipher suites.

Use `--insecure` only to determine whether certificate validation is the specific failure. Do not treat it as a permanent fix.

## TLS works, HTTP fails

Check:

- Proxy environment variables.
- Redirect loops.
- Authentication requirements.
- Application crashes or upstream failures.
- Request filtering by a CDN or web application firewall.

## High TTFB

A high time to first byte often indicates server-side work rather than raw network latency.

Investigate:

- Database query time.
- Upstream API calls.
- Cold starts.
- Cache misses.
- Application thread or worker saturation.

## `traceroute` is not installed

On Debian or Ubuntu:

```bash
sudo apt install traceroute
```

On Arch Linux:

```bash
sudo pacman -S traceroute
```
