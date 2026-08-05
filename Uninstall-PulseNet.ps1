[CmdletBinding()]
param(
    [switch]$Quiet
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$InstallDirectory = Split-Path -Parent $PSCommandPath
$ShortcutPath = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\PulseNet.lnk'
$UninstallKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\PulseNet'

function Remove-DirectoryFromUserPath {
    param([string]$Directory)

    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not $current) { return }

    $remaining = $current.Split(';') | Where-Object {
        $_ -and -not [string]::Equals(
            $_.TrimEnd('\'),
            $Directory.TrimEnd('\'),
            [System.StringComparison]::OrdinalIgnoreCase
        )
    }

    [Environment]::SetEnvironmentVariable('Path', ($remaining -join ';'), 'User')
}

if (-not $Quiet) {
    Write-Host ''
    Write-Host 'PulseNet uninstaller' -ForegroundColor Cyan
    $answer = (Read-Host "Remove PulseNet from $InstallDirectory? [y/N]").Trim().ToLowerInvariant()
    if ($answer -notin @('y', 'yes')) {
        Write-Host 'Uninstall cancelled.' -ForegroundColor Yellow
        exit 0
    }
}

Write-Host '[PulseNet] Removing the user PATH entry...' -ForegroundColor Cyan
Remove-DirectoryFromUserPath -Directory $InstallDirectory

Write-Host '[PulseNet] Removing the Start Menu shortcut...' -ForegroundColor Cyan
Remove-Item -LiteralPath $ShortcutPath -Force -ErrorAction SilentlyContinue

Write-Host '[PulseNet] Removing the Windows uninstall registration...' -ForegroundColor Cyan
Remove-Item -LiteralPath $UninstallKey -Recurse -Force -ErrorAction SilentlyContinue

Write-Host '[PulseNet] Removing installed files...' -ForegroundColor Cyan
$knownFiles = @(
    'PulseNet.exe',
    'integrity.json',
    'installation.json',
    'pn.cmd',
    'pncheck.cmd',
    'pnlogs.cmd',
    'pndump.cmd',
    'pnwatch.cmd',
    'pnrouter.cmd',
    'pnsha.cmd'
)

foreach ($name in $knownFiles) {
    Remove-Item -LiteralPath (Join-Path $InstallDirectory $name) -Force -ErrorAction SilentlyContinue
}

$selfPath = $PSCommandPath
try {
    Remove-Item -LiteralPath $selfPath -Force -ErrorAction Stop
}
catch {
    Write-Warning "Could not remove the uninstaller automatically: $selfPath"
}

try {
    Remove-Item -LiteralPath $InstallDirectory -Force -ErrorAction Stop
}
catch {
    Write-Warning "The installation folder contains additional files and was kept: $InstallDirectory"
}

Write-Host ''
Write-Host 'PulseNet was removed. Open terminals keep their old PATH until restarted.' -ForegroundColor Green
