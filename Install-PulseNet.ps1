[CmdletBinding()]
param(
    [string]$InstallDirectory = (Join-Path $env:LOCALAPPDATA 'Programs\PulseNet'),
    [switch]$QuickCommands,
    [switch]$NoPath,
    [switch]$NoShortcut,
    [switch]$Quiet,
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Repository = 'appoloncel283-debug/pulsenet'
$ReleaseBase = "https://github.com/$Repository/releases/download/latest"
$ExecutableUrl = "$ReleaseBase/PulseNet.exe"
$UninstallerUrl = "$ReleaseBase/Uninstall-PulseNet.ps1"
$ChecksumsUrl = "$ReleaseBase/SHA256SUMS.txt"
$ExecutablePath = Join-Path $InstallDirectory 'PulseNet.exe'
$UninstallerPath = Join-Path $InstallDirectory 'Uninstall-PulseNet.ps1'
$ShortcutPath = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\PulseNet.lnk'
$UninstallKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\PulseNet'

function Write-Step {
    param([string]$Message)
    Write-Host "[PulseNet] $Message" -ForegroundColor Cyan
}

function Ask-YesNo {
    param(
        [string]$Question,
        [bool]$DefaultYes
    )

    $suffix = if ($DefaultYes) { '[Y/n]' } else { '[y/N]' }
    while ($true) {
        $answer = (Read-Host "$Question $suffix").Trim().ToLowerInvariant()
        if ($answer -eq '') { return $DefaultYes }
        if ($answer -in @('y', 'yes')) { return $true }
        if ($answer -in @('n', 'no')) { return $false }
        Write-Host 'Please enter Y or N.' -ForegroundColor Yellow
    }
}

function Get-ExpectedHash {
    param(
        [string]$ChecksumText,
        [string]$FileName
    )

    foreach ($line in ($ChecksumText -split "`r?`n")) {
        if ($line -match "^([A-Fa-f0-9]{64})\s+\*?$([regex]::Escape($FileName))$") {
            return $Matches[1].ToUpperInvariant()
        }
    }
    throw "SHA-256 entry for $FileName was not found in SHA256SUMS.txt."
}

function Assert-Hash {
    param(
        [string]$Path,
        [string]$ExpectedHash
    )

    $actualHash = (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToUpperInvariant()
    if ($actualHash -ne $ExpectedHash) {
        throw "Checksum verification failed for $(Split-Path $Path -Leaf). Expected $ExpectedHash, received $actualHash."
    }
}

function Add-DirectoryToUserPath {
    param([string]$Directory)

    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @()
    if ($current) {
        $parts = $current.Split(';') | Where-Object { $_ -and $_.Trim() }
    }

    $alreadyPresent = $parts | Where-Object {
        [string]::Equals($_.TrimEnd('\'), $Directory.TrimEnd('\'), [System.StringComparison]::OrdinalIgnoreCase)
    }

    if (-not $alreadyPresent) {
        $newValue = (($parts + $Directory) -join ';')
        [Environment]::SetEnvironmentVariable('Path', $newValue, 'User')
    }

    if (-not (($env:Path -split ';') | Where-Object {
        [string]::Equals($_.TrimEnd('\'), $Directory.TrimEnd('\'), [System.StringComparison]::OrdinalIgnoreCase)
    })) {
        $env:Path = "$Directory;$env:Path"
    }
}

function Install-QuickCommands {
    param([string]$Directory)

    $commands = [ordered]@{
        'pn.cmd'       = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" %*' + "`r`n"
        'pncheck.cmd'  = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" diagnose %*' + "`r`n"
        'pnlogs.cmd'   = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" logs %*' + "`r`n"
        'pndump.cmd'   = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" dump %*' + "`r`n"
        'pnwatch.cmd'  = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" watch %*' + "`r`n"
    }

    foreach ($entry in $commands.GetEnumerator()) {
        Set-Content -LiteralPath (Join-Path $Directory $entry.Key) -Value $entry.Value -Encoding ASCII -NoNewline
    }
}

function Remove-QuickCommands {
    param([string]$Directory)

    foreach ($name in @('pn.cmd', 'pncheck.cmd', 'pnlogs.cmd', 'pndump.cmd', 'pnwatch.cmd')) {
        Remove-Item -LiteralPath (Join-Path $Directory $name) -Force -ErrorAction SilentlyContinue
    }
}

function Install-StartMenuShortcut {
    param(
        [string]$TargetPath,
        [string]$LinkPath
    )

    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($LinkPath)
    $shortcut.TargetPath = $TargetPath
    $shortcut.WorkingDirectory = Split-Path $TargetPath -Parent
    $shortcut.Description = 'PulseNet network diagnostics'
    $shortcut.IconLocation = "$TargetPath,0"
    $shortcut.Save()
}

$addToPath = -not $NoPath
$installQuickCommands = $QuickCommands.IsPresent
$createShortcut = -not $NoShortcut

if (-not $Quiet) {
    Write-Host ''
    Write-Host 'PulseNet installer' -ForegroundColor Cyan
    Write-Host 'The installer downloads the latest official release and verifies its SHA-256 checksum.'
    Write-Host ''
    $addToPath = Ask-YesNo 'Add PulseNet to your user PATH so the pulsenet command works everywhere?' $true
    $installQuickCommands = Ask-YesNo 'Install short commands (pn, pncheck, pnlogs, pndump, pnwatch)?' $false
    $createShortcut = Ask-YesNo 'Create a Start Menu shortcut?' $true
}

if ($installQuickCommands) {
    $addToPath = $true
}

$tempDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("PulseNet-Install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempDirectory -Force | Out-Null
$tempExecutable = Join-Path $tempDirectory 'PulseNet.exe'
$tempUninstaller = Join-Path $tempDirectory 'Uninstall-PulseNet.ps1'

try {
    Write-Step 'Downloading release checksums...'
    $checksumResponse = Invoke-WebRequest -Uri $ChecksumsUrl -UseBasicParsing
    $checksumText = [string]$checksumResponse.Content

    Write-Step 'Downloading PulseNet.exe...'
    Invoke-WebRequest -Uri $ExecutableUrl -OutFile $tempExecutable -UseBasicParsing

    Write-Step 'Downloading the uninstaller...'
    Invoke-WebRequest -Uri $UninstallerUrl -OutFile $tempUninstaller -UseBasicParsing

    Write-Step 'Verifying SHA-256 checksums...'
    Assert-Hash -Path $tempExecutable -ExpectedHash (Get-ExpectedHash -ChecksumText $checksumText -FileName 'PulseNet.exe')
    Assert-Hash -Path $tempUninstaller -ExpectedHash (Get-ExpectedHash -ChecksumText $checksumText -FileName 'Uninstall-PulseNet.ps1')

    $stream = [System.IO.File]::OpenRead($tempExecutable)
    try {
        if (($stream.ReadByte() -ne 0x4D) -or ($stream.ReadByte() -ne 0x5A)) {
            throw 'The downloaded executable does not contain a valid Windows PE header.'
        }
    }
    finally {
        $stream.Dispose()
    }

    if ((Test-Path -LiteralPath $ExecutablePath) -and -not $Force -and -not $Quiet) {
        if (-not (Ask-YesNo 'PulseNet is already installed. Replace the existing installation?' $true)) {
            Write-Host 'Installation cancelled.' -ForegroundColor Yellow
            exit 0
        }
    }

    Write-Step "Installing to $InstallDirectory..."
    New-Item -ItemType Directory -Path $InstallDirectory -Force | Out-Null
    Copy-Item -LiteralPath $tempExecutable -Destination $ExecutablePath -Force
    Copy-Item -LiteralPath $tempUninstaller -Destination $UninstallerPath -Force

    if ($installQuickCommands) {
        Write-Step 'Installing quick commands...'
        Install-QuickCommands -Directory $InstallDirectory
    }
    else {
        Remove-QuickCommands -Directory $InstallDirectory
    }

    if ($addToPath) {
        Write-Step 'Updating the user PATH...'
        Add-DirectoryToUserPath -Directory $InstallDirectory
    }

    if ($createShortcut) {
        Write-Step 'Creating the Start Menu shortcut...'
        Install-StartMenuShortcut -TargetPath $ExecutablePath -LinkPath $ShortcutPath
    }
    else {
        Remove-Item -LiteralPath $ShortcutPath -Force -ErrorAction SilentlyContinue
    }

    Write-Step 'Registering the uninstaller for the current user...'
    New-Item -Path $UninstallKey -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayName -Value 'PulseNet' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayVersion -Value '2.2.0' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name Publisher -Value 'PulseNet' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name InstallLocation -Value $InstallDirectory -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayIcon -Value $ExecutablePath -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name UninstallString -Value ('powershell.exe -NoProfile -ExecutionPolicy RemoteSigned -File "' + $UninstallerPath + '"') -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoModify -Value 1 -PropertyType DWord -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoRepair -Value 1 -PropertyType DWord -Force | Out-Null

    $manifest = [ordered]@{
        installed_at = (Get-Date).ToString('o')
        install_directory = $InstallDirectory
        added_to_path = $addToPath
        quick_commands = $installQuickCommands
        start_menu_shortcut = $createShortcut
    }
    $manifest | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $InstallDirectory 'installation.json') -Encoding UTF8

    Write-Host ''
    Write-Host 'PulseNet was installed successfully.' -ForegroundColor Green
    Write-Host "Executable: $ExecutablePath"
    if ($addToPath) {
        Write-Host 'Open a new terminal and run: pulsenet'
    }
    if ($installQuickCommands) {
        Write-Host 'Quick commands: pn, pncheck, pnlogs, pndump, pnwatch'
    }
    Write-Host "Uninstall from Windows Settings or run: $UninstallerPath"
}
finally {
    Remove-Item -LiteralPath $tempDirectory -Recurse -Force -ErrorAction SilentlyContinue
}
