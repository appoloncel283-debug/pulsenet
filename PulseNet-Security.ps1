[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'target', 'web', 'local', 'file', 'hash', 'token', 'encode', 'decode', 'router', 'integrity', 'db-tools')]
    [string]$Command = 'help',

    [Parameter(Position = 1)]
    [string]$Value,

    [string]$Ports = '21,22,25,53,80,110,143,443,445,587,993,995,1433,3306,3389,5432,6379,8080,8443',
    [string]$OutFile,
    [ValidateRange(16, 256)]
    [int]$Bytes = 32,
    [switch]$Authorized
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Banner {
    Write-Host ''
    Write-Host '╔══════════════════════════════════════════════╗' -ForegroundColor DarkCyan
    Write-Host '║        PULSENET SECURITY TOOLKIT            ║' -ForegroundColor Cyan
    Write-Host '║  authorized diagnostics • local defense    ║' -ForegroundColor DarkCyan
    Write-Host '╚══════════════════════════════════════════════╝' -ForegroundColor DarkCyan
    Write-Host ''
}

function Resolve-PulseNet {
    $localExecutable = Join-Path $PSScriptRoot 'PulseNet.exe'
    if (Test-Path -LiteralPath $localExecutable -PathType Leaf) {
        return $localExecutable
    }

    $commandInfo = Get-Command pulsenet -ErrorAction SilentlyContinue
    if ($commandInfo) {
        return $commandInfo.Source
    }

    throw 'PulseNet.exe was not found. Install PulseNet or place this script next to PulseNet.exe.'
}

function Assert-AuthorizedTarget {
    if (-not $Authorized) {
        throw 'Target auditing requires -Authorized. Use it only on systems you own or are explicitly permitted to test.'
    }
}

function Invoke-PulseNet {
    param([string[]]$Arguments)

    $executable = Resolve-PulseNet
    Write-Host ('> pulsenet ' + ($Arguments -join ' ')) -ForegroundColor DarkGray
    & $executable @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "PulseNet command failed with exit code $LASTEXITCODE."
    }
}

function Invoke-Recorded {
    param(
        [scriptblock]$Action,
        [string]$Path
    )

    if ([string]::IsNullOrWhiteSpace($Path)) {
        & $Action
        return
    }

    $parent = Split-Path -Parent $Path
    if ($parent) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }

    & $Action 2>&1 | Tee-Object -FilePath $Path
    Write-Host "`nReport saved to $Path" -ForegroundColor Green
}

function Show-Help {
    @'
Usage:
  .\PulseNet-Security.ps1 help
  .\PulseNet-Security.ps1 target <host-or-url> -Authorized [-Ports <list>] [-OutFile report.txt]
  .\PulseNet-Security.ps1 web <url> -Authorized [-OutFile report.txt]
  .\PulseNet-Security.ps1 local [-OutFile report.txt]
  .\PulseNet-Security.ps1 file <path>
  .\PulseNet-Security.ps1 hash <path>
  .\PulseNet-Security.ps1 token [-Bytes 32]
  .\PulseNet-Security.ps1 encode <text>
  .\PulseNet-Security.ps1 decode <base64>
  .\PulseNet-Security.ps1 router [address]
  .\PulseNet-Security.ps1 integrity
  .\PulseNet-Security.ps1 db-tools

Examples:
  .\PulseNet-Security.ps1 target example.com -Authorized
  .\PulseNet-Security.ps1 target 192.168.1.10 -Authorized -Ports 22,80,443,3389
  .\PulseNet-Security.ps1 web https://example.com -Authorized -OutFile .\reports\web.txt
  .\PulseNet-Security.ps1 local -OutFile .\reports\local.txt
  .\PulseNet-Security.ps1 file .\download.exe
  .\PulseNet-Security.ps1 router 192.168.1.1

This toolkit does not recover passwords, bypass authentication, exploit services, or dump private databases.
'@ | Write-Host
}

function Invoke-TargetAudit {
    Assert-AuthorizedTarget
    if ([string]::IsNullOrWhiteSpace($Value)) {
        throw 'Provide a host or URL.'
    }

    Invoke-Recorded -Path $OutFile -Action {
        Write-Host '=== TARGET OVERVIEW ===' -ForegroundColor Cyan
        Invoke-PulseNet @('diagnose', $Value, '--attempts', '3', '--timeout', '5s')

        Write-Host "`n=== DNS ===" -ForegroundColor Cyan
        Invoke-PulseNet @('dns', $Value, '--timeout', '5s')

        $hostValue = $Value
        try {
            if ($Value -match '^https?://') {
                $hostValue = ([uri]$Value).Host
            }
        }
        catch {
            $hostValue = $Value
        }

        Write-Host "`n=== AUTHORIZED PORT EXPOSURE ===" -ForegroundColor Cyan
        Invoke-PulseNet @('ports', $hostValue, '--ports', $Ports, '--timeout', '2s', '--concurrency', '24')
    }
}

function Invoke-WebAudit {
    Assert-AuthorizedTarget
    if ([string]::IsNullOrWhiteSpace($Value)) {
        throw 'Provide an HTTP or HTTPS URL.'
    }

    Invoke-Recorded -Path $OutFile -Action {
        Write-Host '=== WEB DIAGNOSIS ===' -ForegroundColor Cyan
        Invoke-PulseNet @('diagnose', $Value, '--timeout', '6s')

        Write-Host "`n=== SECURITY HEADERS ===" -ForegroundColor Cyan
        Invoke-PulseNet @('headers', $Value, '--timeout', '8s')

        if ($Value -match '^https://') {
            Write-Host "`n=== TLS ===" -ForegroundColor Cyan
            Invoke-PulseNet @('tls', ([uri]$Value).Host, '--timeout', '6s')
        }
    }
}

function Show-LocalPosture {
    Invoke-Recorded -Path $OutFile -Action {
        Write-Host '=== SYSTEM ===' -ForegroundColor Cyan
        $os = Get-CimInstance Win32_OperatingSystem
        [pscustomobject]@{
            ComputerName = $env:COMPUTERNAME
            User         = [Environment]::UserName
            Windows      = $os.Caption
            Version      = $os.Version
            LastBoot     = $os.LastBootUpTime
        } | Format-List

        Write-Host '=== ACTIVE NETWORK ADAPTERS ===' -ForegroundColor Cyan
        Get-NetIPConfiguration -ErrorAction SilentlyContinue |
            Where-Object { $_.NetAdapter.Status -eq 'Up' } |
            Select-Object InterfaceAlias, IPv4Address, IPv4DefaultGateway, DNSServer |
            Format-List

        Write-Host '=== FIREWALL PROFILES ===' -ForegroundColor Cyan
        Get-NetFirewallProfile -ErrorAction SilentlyContinue |
            Select-Object Name, Enabled, DefaultInboundAction, DefaultOutboundAction |
            Format-Table -AutoSize

        Write-Host '=== MICROSOFT DEFENDER ===' -ForegroundColor Cyan
        if (Get-Command Get-MpComputerStatus -ErrorAction SilentlyContinue) {
            Get-MpComputerStatus |
                Select-Object AntivirusEnabled, RealTimeProtectionEnabled, BehaviorMonitorEnabled, AntivirusSignatureLastUpdated |
                Format-List
        }
        else {
            Write-Host 'Microsoft Defender status command is unavailable.' -ForegroundColor Yellow
        }

        Write-Host '=== LISTENING TCP PORTS ===' -ForegroundColor Cyan
        Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue |
            Sort-Object LocalPort -Unique |
            Select-Object LocalAddress, LocalPort, OwningProcess |
            Format-Table -AutoSize

        Write-Host '=== ROUTER ===' -ForegroundColor Cyan
        Invoke-PulseNet @('router', 'info')

        Write-Host '=== PULSENET INTEGRITY ===' -ForegroundColor Cyan
        Invoke-PulseNet @('integrity')
    }
}

function Inspect-File {
    if ([string]::IsNullOrWhiteSpace($Value)) {
        throw 'Provide a file path.'
    }

    $resolved = (Resolve-Path -LiteralPath $Value).Path
    $item = Get-Item -LiteralPath $resolved
    if ($item.PSIsContainer) {
        throw 'The supplied path is a directory, not a file.'
    }

    $hash = Get-FileHash -LiteralPath $resolved -Algorithm SHA256
    $signature = Get-AuthenticodeSignature -LiteralPath $resolved
    $zone = Get-Content -LiteralPath "$resolved`:Zone.Identifier" -ErrorAction SilentlyContinue

    [pscustomobject]@{
        Path              = $resolved
        SizeBytes         = $item.Length
        Modified          = $item.LastWriteTime
        SHA256            = $hash.Hash
        SignatureStatus   = $signature.Status
        SignerCertificate = if ($signature.SignerCertificate) { $signature.SignerCertificate.Subject } else { $null }
        HasMarkOfTheWeb    = [bool]$zone
    } | Format-List

    if ($zone) {
        Write-Host 'Mark-of-the-Web metadata:' -ForegroundColor Yellow
        $zone | ForEach-Object { Write-Host "  $_" }
    }
}

function Show-Hash {
    if ([string]::IsNullOrWhiteSpace($Value)) {
        throw 'Provide a file path.'
    }
    Get-FileHash -LiteralPath $Value -Algorithm SHA256 | Format-List
}

function New-SecureToken {
    $buffer = New-Object byte[] $Bytes
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($buffer)
    [Convert]::ToHexString($buffer).ToLowerInvariant() | Write-Output
}

function Encode-Text {
    if ($null -eq $Value) { throw 'Provide text to encode.' }
    [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Value)) | Write-Output
}

function Decode-Text {
    if ([string]::IsNullOrWhiteSpace($Value)) { throw 'Provide Base64 text to decode.' }
    [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($Value)) | Write-Output
}

function Open-Router {
    if ([string]::IsNullOrWhiteSpace($Value)) {
        Invoke-PulseNet @('router', 'open')
    }
    else {
        Invoke-PulseNet @('router', 'open', $Value)
    }
}

Write-Banner

switch ($Command) {
    'help'      { Show-Help }
    'target'    { Invoke-TargetAudit }
    'web'       { Invoke-WebAudit }
    'local'     { Show-LocalPosture }
    'file'      { Inspect-File }
    'hash'      { Show-Hash }
    'token'     { New-SecureToken }
    'encode'    { Encode-Text }
    'decode'    { Decode-Text }
    'router'    { Open-Router }
    'integrity' { Invoke-PulseNet @('integrity') }
    'db-tools'  { Invoke-PulseNet @('db', 'tools') }
}
