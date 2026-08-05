[CmdletBinding()]
param(
    [string]$InstallDirectory = (Join-Path $env:LOCALAPPDATA 'Programs\PulseNet'),
    [string]$ReleaseTag,
    [switch]$QuickCommands,
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
    # PowerShell 7 handles TLS negotiation without this compatibility setting.
}

$Repository = 'appoloncel283-debug/pulsenet'
$ProductVersion = '2.4.0'
$ApiHeaders = @{
    Accept = 'application/vnd.github+json'
    'User-Agent' = "PulseNet-Installer/$ProductVersion"
}
$ExecutablePath = Join-Path $InstallDirectory 'PulseNet.exe'
$UninstallerPath = Join-Path $InstallDirectory 'Uninstall-PulseNet.ps1'
$IntegrityPath = Join-Path $InstallDirectory 'integrity.json'
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

function Resolve-ReleaseTag {
    param([string]$RequestedTag)

    if (-not [string]::IsNullOrWhiteSpace($RequestedTag)) {
        return $RequestedTag.Trim()
    }

    Write-Step 'Resolving the latest stable release...'
    try {
        $request = @{
            Uri = "https://api.github.com/repos/$Repository/releases/latest"
            Headers = $ApiHeaders
            UseBasicParsing = $true
        }
        $release = Invoke-RestMethod @request
    }
    catch {
        throw "Could not resolve the latest PulseNet release: $($_.Exception.Message)"
    }

    $tag = [string]$release.tag_name
    if ([string]::IsNullOrWhiteSpace($tag)) {
        throw 'GitHub returned a release without a tag name.'
    }
    return $tag.Trim()
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

        $match = [regex]::Match(
            $trimmed,
            '^(?<hash>[A-Fa-f0-9]{64})\s+\*?(?<name>.+?)\s*$'
        )
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

$addToPath = -not $NoPath
$installQuickCommands = $QuickCommands.IsPresent
$createShortcut = -not $NoShortcut

if (-not $Quiet) {
    Write-Host ''
    Write-Host 'PulseNet installer' -ForegroundColor Cyan
    Write-Host 'The installer pins one official release and verifies every downloaded file with SHA-256.'
    Write-Host ''
    $addToPath = Ask-YesNo 'Add PulseNet to your user PATH so the pulsenet command works everywhere?' $true
    $installQuickCommands = Ask-YesNo 'Install short commands (pn, pncheck, pnlogs, pndump, pnwatch, pnrouter, pnsha)?' $false
    $createShortcut = Ask-YesNo 'Create a Start Menu shortcut?' $true
}

if ($installQuickCommands) {
    $addToPath = $true
}

$resolvedTag = Resolve-ReleaseTag -RequestedTag $ReleaseTag
$encodedTag = [uri]::EscapeDataString($resolvedTag)
$releaseBase = "https://github.com/$Repository/releases/download/$encodedTag"
$executableUrl = "$releaseBase/PulseNet.exe"
$uninstallerUrl = "$releaseBase/Uninstall-PulseNet.ps1"
$checksumsUrl = "$releaseBase/SHA256SUMS.txt"

$tempDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("PulseNet-Install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempDirectory -Force | Out-Null
$tempExecutable = Join-Path $tempDirectory 'PulseNet.exe'
$tempUninstaller = Join-Path $tempDirectory 'Uninstall-PulseNet.ps1'
$tempChecksums = Join-Path $tempDirectory 'SHA256SUMS.txt'

try {
    Write-Step "Using release $resolvedTag."
    Write-Step 'Downloading release checksums...'
    Download-File -Uri $checksumsUrl -Destination $tempChecksums

    Write-Step 'Downloading PulseNet.exe...'
    Download-File -Uri $executableUrl -Destination $tempExecutable

    Write-Step 'Downloading the uninstaller...'
    Download-File -Uri $uninstallerUrl -Destination $tempUninstaller

    $checksumText = Get-Content -LiteralPath $tempChecksums -Raw
    $expectedExecutableHash = Get-ExpectedHash -ChecksumText $checksumText -FileName 'PulseNet.exe'
    $expectedUninstallerHash = Get-ExpectedHash -ChecksumText $checksumText -FileName 'Uninstall-PulseNet.ps1'

    Write-Step 'Verifying SHA-256 checksums...'
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

    Write-Step "Installing to $InstallDirectory..."
    New-Item -ItemType Directory -Path $InstallDirectory -Force | Out-Null
    Copy-Item -LiteralPath $tempExecutable -Destination $ExecutablePath -Force
    Copy-Item -LiteralPath $tempUninstaller -Destination $UninstallerPath -Force

    Write-Step 'Writing the startup integrity manifest...'
    $integrityManifest = [ordered]@{
        product = 'PulseNet'
        version = $ProductVersion
        release_tag = $resolvedTag
        executable = 'PulseNet.exe'
        sha256 = $expectedExecutableHash
        generated_at = (Get-Date).ToUniversalTime().ToString('o')
    }
    $integrityManifest | ConvertTo-Json | Set-Content -LiteralPath $IntegrityPath -Encoding ASCII

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
    New-ItemProperty -Path $UninstallKey -Name DisplayVersion -Value $ProductVersion -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name Publisher -Value 'PulseNet' -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name InstallLocation -Value $InstallDirectory -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name DisplayIcon -Value $ExecutablePath -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name UninstallString -Value ('powershell.exe -NoProfile -ExecutionPolicy RemoteSigned -File "' + $UninstallerPath + '"') -PropertyType String -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoModify -Value 1 -PropertyType DWord -Force | Out-Null
    New-ItemProperty -Path $UninstallKey -Name NoRepair -Value 1 -PropertyType DWord -Force | Out-Null

    $manifest = [ordered]@{
        installed_at = (Get-Date).ToString('o')
        release_tag = $resolvedTag
        version = $ProductVersion
        install_directory = $InstallDirectory
        added_to_path = $addToPath
        quick_commands = $installQuickCommands
        start_menu_shortcut = $createShortcut
    }
    $manifest | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $InstallDirectory 'installation.json') -Encoding ASCII

    Write-Host ''
    Write-Host 'PulseNet was installed successfully.' -ForegroundColor Green
    Write-Host "Release: $resolvedTag"
    Write-Host "Executable: $ExecutablePath"
    Write-Host "Verified SHA-256: $expectedExecutableHash"
    if ($addToPath) {
        Write-Host 'Open a new terminal and run: pulsenet'
    }
    if ($installQuickCommands) {
        Write-Host 'Quick commands: pn, pncheck, pnlogs, pndump, pnwatch, pnrouter, pnsha'
    }
    Write-Host "Uninstall from Windows Settings or run: $UninstallerPath"
}
finally {
    Remove-Item -LiteralPath $tempDirectory -Recurse -Force -ErrorAction SilentlyContinue
}
