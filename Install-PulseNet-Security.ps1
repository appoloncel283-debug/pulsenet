[CmdletBinding()]
param(
    [string]$InstallDirectory = (Join-Path $env:LOCALAPPDATA 'Programs\PulseNet'),
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repository = 'appoloncel283-debug/pulsenet'
$Headers = @{
    Accept = 'application/vnd.github+json'
    'User-Agent' = 'PulseNet-Security-Installer/2.5.0'
}

function Write-Step {
    param([string]$Message)
    Write-Host "[PulseNet Security] $Message" -ForegroundColor Cyan
}

function Download-File {
    param([string]$Uri, [string]$Destination)
    Invoke-WebRequest -Uri $Uri -OutFile $Destination -Headers $Headers -UseBasicParsing
    if (-not (Test-Path -LiteralPath $Destination -PathType Leaf)) {
        throw "Download did not create $Destination."
    }
}

function Get-ExpectedHash {
    param([string]$Manifest, [string]$FileName)

    foreach ($line in [regex]::Split($Manifest, '\r?\n')) {
        $match = [regex]::Match($line.Trim(), '^(?<hash>[A-Fa-f0-9]{64})\s+\*?(?<name>.+?)\s*$')
        if (-not $match.Success) { continue }
        $leaf = [IO.Path]::GetFileName($match.Groups['name'].Value.Replace('\', '/'))
        if ([string]::Equals($leaf, $FileName, [StringComparison]::OrdinalIgnoreCase)) {
            return $match.Groups['hash'].Value.ToUpperInvariant()
        }
    }
    throw "SHA-256 entry for $FileName was not found."
}

function Assert-Hash {
    param([string]$Path, [string]$Expected)
    $actual = (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToUpperInvariant()
    if ($actual -ne $Expected) {
        throw "SHA-256 verification failed for $(Split-Path $Path -Leaf)."
    }
}

Write-Step 'Resolving the latest release...'
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest" -Headers $Headers -UseBasicParsing
$tag = [string]$release.tag_name
if ([string]::IsNullOrWhiteSpace($tag)) {
    throw 'GitHub returned a release without a tag.'
}

$versionMatch = [regex]::Match($tag, '^v?(?<version>\d+\.\d+\.\d+)')
if (-not $versionMatch.Success -or [version]$versionMatch.Groups['version'].Value -lt [version]'2.5.0') {
    throw "The latest release ($tag) does not contain PulseNet Security Toolkit 2.5.0 or newer."
}

$base = "https://github.com/$Repository/releases/download/$([uri]::EscapeDataString($tag))"
$temp = Join-Path ([IO.Path]::GetTempPath()) ("PulseNet-Security-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp -Force | Out-Null

try {
    $manifestPath = Join-Path $temp 'SHA256SUMS.txt'
    $scriptPath = Join-Path $temp 'PulseNet-Security.ps1'
    $launcherPath = Join-Path $temp 'pnsec.cmd'

    Write-Step "Downloading assets from $tag..."
    Download-File "$base/SHA256SUMS.txt" $manifestPath
    Download-File "$base/PulseNet-Security.ps1" $scriptPath
    Download-File "$base/pnsec.cmd" $launcherPath

    $manifest = Get-Content -LiteralPath $manifestPath -Raw
    Write-Step 'Verifying SHA-256...'
    Assert-Hash $scriptPath (Get-ExpectedHash $manifest 'PulseNet-Security.ps1')
    Assert-Hash $launcherPath (Get-ExpectedHash $manifest 'pnsec.cmd')

    $pulseNetPath = Join-Path $InstallDirectory 'PulseNet.exe'
    if (-not (Test-Path -LiteralPath $pulseNetPath -PathType Leaf)) {
        throw "PulseNet.exe was not found in $InstallDirectory. Install PulseNet first."
    }

    $destinationScript = Join-Path $InstallDirectory 'PulseNet-Security.ps1'
    $destinationLauncher = Join-Path $InstallDirectory 'pnsec.cmd'

    if (-not $Force -and ((Test-Path $destinationScript) -or (Test-Path $destinationLauncher))) {
        $answer = (Read-Host 'Security Toolkit is already installed. Replace it? [Y/n]').Trim().ToLowerInvariant()
        if ($answer -in @('n', 'no')) {
            Write-Host 'Installation cancelled.' -ForegroundColor Yellow
            exit 0
        }
    }

    Copy-Item -LiteralPath $scriptPath -Destination $destinationScript -Force
    Copy-Item -LiteralPath $launcherPath -Destination $destinationLauncher -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @()
    if ($userPath) { $parts = $userPath.Split(';') | Where-Object { $_ } }
    $present = $parts | Where-Object {
        [string]::Equals($_.TrimEnd('\'), $InstallDirectory.TrimEnd('\'), [StringComparison]::OrdinalIgnoreCase)
    }
    if (-not $present) {
        [Environment]::SetEnvironmentVariable('Path', (($parts + $InstallDirectory) -join ';'), 'User')
    }

    Write-Host ''
    Write-Host 'PulseNet Security Toolkit installed successfully.' -ForegroundColor Green
    Write-Host "Location: $InstallDirectory"
    Write-Host 'Open a new PowerShell window and run: pnsec help'
}
finally {
    Remove-Item -LiteralPath $temp -Recurse -Force -ErrorAction SilentlyContinue
}
