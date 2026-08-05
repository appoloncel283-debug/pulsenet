[CmdletBinding()]
param(
    [string]$InstallDirectory = (Join-Path $env:LOCALAPPDATA 'Programs\PulseNet'),
    [string]$ReleaseTag,
    [switch]$QuickCommands,
    [switch]$NoQuickCommands,
    [switch]$NoPath,
    [switch]$NoShortcut,
    [switch]$Quiet,
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
}
catch {
    # PowerShell 7 negotiates TLS without this compatibility setting.
}

$Repository = 'appoloncel283-debug/pulsenet'
$MinimumVersion = [version]'2.4.0'
$ApiHeaders = @{
    Accept = 'application/vnd.github+json'
    'User-Agent' = 'PulseNet-Installer/2.4.0'
}
$ExecutablePath = Join-Path $InstallDirectory 'PulseNet.exe'
$UninstallerPath = Join-Path $InstallDirectory 'Uninstall-PulseNet.ps1'
$IntegrityPath = Join-Path $InstallDirectory 'integrity.json'
$InstallationPath = Join-Path $InstallDirectory 'installation.json'
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

function Get-ReleaseVersion {
    param([string]$Tag)

    $match = [regex]::Match($Tag, '^v?(?<version>\d+\.\d+\.\d+)')
    if (-not $match.Success) {
        throw "Release tag '$Tag' does not contain a semantic version."
    }
    return [version]$match.Groups['version'].Value
}

function Resolve-Release {
    param([string]$RequestedTag)

    if ([string]::IsNullOrWhiteSpace($RequestedTag)) {
        $uri = "https://api.github.com/repos/$Repository/releases/latest"
        Write-Step 'Resolving the latest stable release...'
    }
    else {
        $encodedTag = [uri]::EscapeDataString($RequestedTag.Trim())
        $uri = "https://api.github.com/repos/$Repository/releases/tags/$encodedTag"
        Write-Step "Resolving release $($RequestedTag.Trim())..."
    }

    try {
        return Invoke-RestMethod -Uri $uri -Headers $ApiHeaders
    }
    catch {
        throw "Could not resolve a PulseNet release: $($_.Exception.Message)"
    }
}

function Get-AssetUrl {
    param(
        [object]$Release,
        [string]$Name
    )

    $asset = @($Release.assets) | Where-Object {
        [string]::Equals([string]$_.name, $Name, [System.StringComparison]::OrdinalIgnoreCase)
    } | Select-Object -First 1

    if ($null -eq $asset -or [string]::IsNullOrWhiteSpace([string]$asset.browser_download_url)) {
        throw "Release $($Release.tag_name) does not contain required asset $Name."
    }
    return [string]$asset.browser_download_url
}

function Download-File {
    param(
        [string]$Uri,
        [string]$Destination
    )

    try {
        Invoke-WebRequest -Uri $Uri -OutFile $Destination -Headers $ApiHeaders -UseBasicParsing
    }
    catch {
        throw "Download failed for $Uri`: $($_.Exception.Message)"
    }

    if (-not (Test-Path -LiteralPath $Destination -PathType Leaf)) {
        throw "The download did not create $Destination."
    }
    if ((Get-Item -LiteralPath $Destination).Length -le 0) {
        throw "The downloaded file is empty: $Destination"
    }
}

function Get-ExpectedHash {
    param(
        [string]$ChecksumText,
        [string]$FileName
    )

    foreach ($line in [regex]::Split($ChecksumText, '\r?\n')) {
        $trimmed = $line.Trim()
        if ($trimmed -eq '') { continue }

        $match = [regex]::Match($trimmed, '^(?<hash>[A-Fa-f0-9]{64})\s+\*?(?<name>.+?)\s*$')
        if (-not $match.Success) { continue }

        $listedName = $match.Groups['name'].Value.Replace('\', '/')
        $listedLeaf = [System.IO.Path]::GetFileName($listedName)
        if ([string]::Equals($listedLeaf, $FileName, [System.StringComparison]::OrdinalIgnoreCase)) {
            return $match.Groups['hash'].Value.ToUpperInvariant()
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

function Write-Utf8NoBom {
    param(
        [string]$Path,
        [string]$Content
    )

    $encoding = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Content, $encoding)
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
        [Environment]::SetEnvironmentVariable('Path', (($parts + $Directory) -join ';'), 'User')
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
        'pnrouter.cmd' = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" router open %*' + "`r`n"
        'pnsha.cmd'    = '@echo off' + "`r`n" + '"%~dp0PulseNet.exe" integrity %*' + "`r`n"
    }

    foreach ($entry in $commands.GetEnumerator()) {
        Set-Content -LiteralPath (Join-Path $Directory $entry.Key) -Value $entry.Value -Encoding ASCII -NoNewline
    }
}

function Remove-QuickCommands {
    param([string]$Directory)

    foreach ($name in @('pn.cmd', 'pncheck.cmd', 'pnlogs.cmd', 'pndump.cmd', 'pnwatch.cmd', 'pnrouter.cmd', 'pnsha.cmd')) {
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

if ($QuickCommands -and $NoQuickCommands) {
    throw 'Use either -QuickCommands or -NoQuickCommands, not both.'
}

$addToPath = -not $NoPath
$installQuickCommands = -not $NoQuickCommands
$createShortcut = -not $NoShortcut

if (-not $Quiet) {
    Write-Host ''
    Write-Host 'PulseNet installer' -ForegroundColor Cyan
    Write-Host 'The installer refuses stale releases and verifies the installed executable before reporting success.'
    Write-Host ''
    $addToPath = Ask-YesNo 'Add PulseNet to your user PATH so the pulsenet command works everywhere?' $true
    $installQuickCommands = Ask-YesNo 'Install short commands (pn, pncheck, pnlogs, pndump, pnwatch, pnrouter, pnsha)?' $true
    $createShortcut = Ask-YesNo 'Create a Start Menu shortcut?' $true
}

if ($QuickCommands) {
    $installQuickCommands = $true
}
if ($installQuickCommands) {
    $addToPath = $true
}

$release = Resolve-Release -RequestedTag $ReleaseTag
$resolvedTag = [string]$release.tag_name
$releaseVersion = Get-ReleaseVersion -Tag $resolvedTag
if ($releaseVersion -lt $MinimumVersion) {
    throw "GitHub's latest published PulseNet release is $releaseVersion, but this installer requires $MinimumVersion or newer. The new release is not ready yet; no old binary was installed."
}

$executableUrl = Get-AssetUrl -Release $release -Name 'PulseNet.exe'
$uninstallerUrl = Get-AssetUrl -Release $release -Name 'Uninstall-PulseNet.ps1'
$checksumsUrl = Get-AssetUrl -Release $release -Name 'SHA256SUMS.txt'

$tempDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("PulseNet-Install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempDirectory -Force | Out-Null
$tempExecutable = Join-Path $tempDirectory 'PulseNet.exe'
$tempUninstaller = Join-Path $tempDirectory 'Uninstall-PulseNet.ps1'
$tempChecksums = Join-Path $tempDirectory 'SHA256SUMS.txt'

try {
    Write-Step "Using immutable release $resolvedTag."
    Write-Step 'Downloading checksums...'
    Download-File -Uri $checksumsUrl -Destination $tempChecksums

    Write-Step 'Downloading PulseNet.exe...'
    Download-File -Uri $executableUrl -Destination $tempExecutable

    Write-Step 'Downloading the uninstaller...'
    Download-File -Uri $uninstallerUrl -Destination $tempUninstaller

    $checksumText = Get-Content -LiteralPath $tempChecksums -Raw
    $expectedExecutableHash = Get-ExpectedHash -ChecksumText $checksumText -FileName 'PulseNet.exe'
    $expectedUninstallerHash = Get-ExpectedHash -ChecksumText $checksumText -FileName 'Uninstall-PulseNet.ps1'

    Write-Step 'Verifying downloaded SHA-256 checksums...'
    Assert-Hash -Path $tempExecutable -ExpectedHash $expectedExecutableHash
    Assert-Hash -Path $tempUninstaller -ExpectedHash $expectedUninstallerHash

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

    Write-Step "Installing PulseNet $releaseVersion to $InstallDirectory..."
    New-Item -ItemType Directory -Path $InstallDirectory -Force | Out-Null
    Copy-Item -LiteralPath $tempExecutable -Destination $ExecutablePath -Force
    Copy-Item -LiteralPath $tempUninstaller -Destination $UninstallerPath -Force

    $integrityManifest = [ordered]@{
        product = 'PulseNet'
        version = $releaseVersion.ToString()
        release_tag = $resolvedTag
        executable = 'PulseNet.exe'
        sha256 = $expectedExecutableHash
        generated_at = (Get-Date).ToUniversalTime().ToString('o')
    }
    Write-Utf8NoBom -Path $IntegrityPath -Content ($integrityManifest | ConvertTo-Json)

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

    Write-Step 'Running installed-command smoke tests...'
    $versionOutput = (& $ExecutablePath version 2>$null | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or $versionOutput -notmatch [regex]::Escape($releaseVersion.ToString())) {
        throw "Installed executable returned an unexpected version: '$versionOutput'."
    }

    $integrityOutput = (& $ExecutablePath integrity --json 2>$null | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($integrityOutput)) {
        throw 'Installed executable does not support the integrity command.'
    }
    try {
        $integrityResult = $integrityOutput | ConvertFrom-Json
    }
    catch {
        throw "Installed integrity command returned invalid JSON: $integrityOutput"
    }
    if ([string]$integrityResult.state -ne 'verified') {
        throw "Installed executable failed its integrity check: $($integrityResult.state)."
    }

    Write-Step 'Registering the uninstaller for the current user...'
    New-Item -Path $UninstallKey -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayName -Value 'PulseNet' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayVersion -Value $releaseVersion.ToString() -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name Publisher -Value 'PulseNet' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name InstallLocation -Value $InstallDirectory -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayIcon -Value $ExecutablePath -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name UninstallString -Value ('powershell.exe -NoProfile -ExecutionPolicy RemoteSigned -File "' + $UninstallerPath + '"') -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoModify -Value 1 -PropertyType DWord -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoRepair -Value 1 -PropertyType DWord -Force | Out-Null

    $installationManifest = [ordered]@{
        installed_at = (Get-Date).ToUniversalTime().ToString('o')
        version = $releaseVersion.ToString()
        release_tag = $resolvedTag
        install_directory = $InstallDirectory
        added_to_path = $addToPath
        quick_commands = $installQuickCommands
        start_menu_shortcut = $createShortcut
    }
    Write-Utf8NoBom -Path $InstallationPath -Content ($installationManifest | ConvertTo-Json)

    Write-Host ''
    Write-Host "PulseNet $releaseVersion was installed and verified successfully." -ForegroundColor Green
    Write-Host "Executable: $ExecutablePath"
    if ($addToPath) {
        Write-Host 'Open a new terminal and run: pulsenet version'
    }
    if ($installQuickCommands) {
        Write-Host 'Quick commands: pn, pncheck, pnlogs, pndump, pnwatch, pnrouter, pnsha'
    }
    Write-Host "Uninstall from Windows Settings or run: $UninstallerPath"
}
finally {
    Remove-Item -LiteralPath $tempDirectory -Recurse -Force -ErrorAction SilentlyContinue
}
