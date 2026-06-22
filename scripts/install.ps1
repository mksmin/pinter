$ErrorActionPreference = "Stop"

$Repo = "mksmin/pinter"
$Version = $env:PINTER_VERSION
if (-not $Version) {
    $Version = "latest"
}

$InstallDir = $env:PINTER_INSTALL_DIR
if (-not $InstallDir) {
    $InstallDir = Join-Path $HOME "bin"
}

$Asset = "pinter-windows-amd64.exe"

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Target = Join-Path $InstallDir "pinter.exe"

Write-Host "Downloading $Url"
Invoke-WebRequest -Uri $Url -OutFile $Target

Write-Host "Installed: $Target"
Write-Host "Run: pinter help"
Write-Host "If pinter is not found, add this to PATH: $InstallDir"