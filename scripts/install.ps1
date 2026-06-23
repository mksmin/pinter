$ErrorActionPreference = "Stop"

function Write-Step {
    param(
        [string]$Message
    )

    Write-Host "==> $Message"
}

$Repo = "mksmin/pinter"
$Version = $env:PINTER_VERSION
if (-not $Version) {
    $Version = "latest"
}

$InstallDir = $env:PINTER_INSTALL_DIR
if (-not $InstallDir) {
    $InstallDir = Join-Path $HOME "bin"
}
$InstallDir = [System.IO.Path]::GetFullPath($InstallDir)

function Get-NormalizedPathEntry {
    param(
        [string]$PathEntry
    )

    if (-not $PathEntry) {
        return ""
    }

    try {
        return [System.IO.Path]::GetFullPath($PathEntry).TrimEnd('\', '/').ToLowerInvariant()
    } catch {
        return $PathEntry.TrimEnd('\', '/').ToLowerInvariant()
    }
}

function Test-PathContains {
    param(
        [string]$PathValue,
        [string]$Directory
    )

    $Needle = Get-NormalizedPathEntry $Directory
    foreach ($Entry in ($PathValue -split [System.IO.Path]::PathSeparator)) {
        if ((Get-NormalizedPathEntry $Entry) -eq $Needle) {
            return $true
        }
    }

    return $false
}

function Add-PathEntry {
    param(
        [string]$PathValue,
        [string]$Directory
    )

    if (-not $PathValue) {
        return $Directory
    }

    return "$PathValue$([System.IO.Path]::PathSeparator)$Directory"
}

$Asset = "pinter-windows-amd64.exe"

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

Write-Step "Preparing pinter installer"
Write-Host "Repository: $Repo"
Write-Host "Version: $Version"
Write-Host "Asset: $Asset"
Write-Host "Install directory: $InstallDir"
Write-Host "Target: $(Join-Path $InstallDir "pinter.exe")"
Write-Host "Download URL: $Url"

Write-Step "Creating install directory"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Target = Join-Path $InstallDir "pinter.exe"

Write-Step "Downloading release asset"
Invoke-WebRequest -Uri $Url -OutFile $Target
Write-Host "Downloaded: $Target"

Write-Step "Updating PATH"
$UserPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
if (-not (Test-PathContains $UserPath $InstallDir)) {
    [System.Environment]::SetEnvironmentVariable(
        "Path",
        (Add-PathEntry $UserPath $InstallDir),
        [System.EnvironmentVariableTarget]::User
    )
    Write-Host "Added to user PATH: $InstallDir"
} else {
    Write-Host "User PATH already contains: $InstallDir"
}

if (-not (Test-PathContains $env:Path $InstallDir)) {
    $env:Path = Add-PathEntry $env:Path $InstallDir
    Write-Host "Added to current session PATH: $InstallDir"
} else {
    Write-Host "Current session PATH already contains: $InstallDir"
}

Write-Step "Install complete"
Write-Host "Installed binary: $Target"
Write-Host "Run: pinter help"
