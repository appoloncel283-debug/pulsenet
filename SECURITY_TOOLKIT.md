# PulseNet Security Toolkit

`PulseNet-Security.ps1` is a Windows security-assessment layer built around PulseNet. It focuses on authorized diagnostics, local defense, file verification, and evidence collection.

It does **not** recover passwords, bypass authentication, exploit services, brute-force accounts, dump private databases, or hide activity.

## Quick start

Place these files next to `PulseNet.exe`:

- `PulseNet-Security.ps1`
- `pnsec.cmd`

Then run:

```powershell
pnsec help
```

## Commands

### Authorized target audit

```powershell
pnsec target example.com -Authorized
pnsec target 192.168.1.10 -Authorized -Ports 22,80,443,3389
pnsec target example.com -Authorized -OutFile .\reports\target.txt
```

The target audit combines:

- full PulseNet diagnosis;
- DNS inspection;
- an explicit, bounded TCP port check;
- optional transcript output.

The `-Authorized` switch is mandatory.

### Authorized web audit

```powershell
pnsec web https://example.com -Authorized
pnsec web https://example.com -Authorized -OutFile .\reports\web.txt
```

The web audit runs:

- full HTTP diagnosis;
- browser security-header review;
- TLS inspection for HTTPS targets.

### Local Windows security posture

```powershell
pnsec local
pnsec local -OutFile .\reports\local.txt
```

This displays:

- Windows version and boot time;
- active network adapters, gateways, and DNS servers;
- Windows Firewall profile status;
- Microsoft Defender status when available;
- local listening TCP ports;
- router information;
- PulseNet executable integrity.

### Inspect a downloaded file

```powershell
pnsec file .\download.exe
```

The output includes:

- file size and modification time;
- SHA-256;
- Authenticode signature status;
- signer certificate when present;
- Mark-of-the-Web presence.

### Hash a file

```powershell
pnsec hash .\archive.zip
```

### Generate a cryptographically secure token

```powershell
pnsec token
pnsec token -Bytes 64
```

### Base64 utilities

```powershell
pnsec encode "hello"
pnsec decode "aGVsbG8="
```

These are data-conversion helpers, not encryption.

### Router assistant

```powershell
pnsec router
pnsec router 192.168.1.1
```

This opens the detected or explicitly supplied local router page. Credentials remain controlled by the router and browser.

### Integrity and database client checks

```powershell
pnsec integrity
pnsec db-tools
```

## Safety model

PulseNet Security Toolkit is for systems you own or are explicitly authorized to assess. The script requires `-Authorized` before target and web audit workflows run.

Port checks remain explicit and bounded. The toolkit does not include stealth, evasion, persistence, credential theft, malware delivery, destructive actions, or automatic exploitation.
