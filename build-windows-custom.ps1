$ErrorActionPreference = "Stop"

Write-Host "=== WBG Albion Data Client Build ===" -ForegroundColor Green

$AppVersion = "1.2.5"
$OutputName = "albiondata-client.exe"  # use standard name for update archive

Write-Host "Checking Go installation..." -ForegroundColor Cyan
$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go not found in PATH" -ForegroundColor Red
    exit 1
}
Write-Host $goVersion

Write-Host "Installing go-winres..." -ForegroundColor Cyan
go install github.com/tc-hib/go-winres@v0.3.1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to install go-winres" -ForegroundColor Red
    exit 1
}

Write-Host "Cleaning old files..." -ForegroundColor Cyan
Remove-Item -Path "rsrc_windows_*" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "$OutputName" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "*.bak" -Force -ErrorAction SilentlyContinue

Write-Host "Generating Windows resources..." -ForegroundColor Cyan
$gopath = if ($env:GOPATH) { $env:GOPATH } else { Join-Path $env:USERPROFILE 'go' }
$goBin = Join-Path $gopath 'bin'

$oldPath = $env:PATH
$env:PATH = $goBin + ';' + $env:PATH

& go-winres make

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to generate resources" -ForegroundColor Red
    $env:PATH = $oldPath
    exit 1
}

Write-Host "Building application..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"

& go build -ldflags "-s -w -X main.version=$AppVersion" -o $OutputName -v albiondata-client.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    $env:PATH = $oldPath
    exit 1
}

Write-Host "Patching executable with resources..." -ForegroundColor Cyan
& go-winres patch $OutputName

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to patch executable" -ForegroundColor Red
    $env:PATH = $oldPath
    exit 1
}

$env:PATH = $oldPath

Write-Host ""
Write-Host "=== Build Successful! ===" -ForegroundColor Green
Write-Host "Output: $OutputName" -ForegroundColor Yellow

if (Test-Path $OutputName) {
    $FileSize = (Get-Item $OutputName).Length / 1MB
    Write-Host ("Size: " + [Math]::Round($FileSize, 2) + " MB") -ForegroundColor Yellow
    $FullPath = (Get-Item $OutputName).FullName
    Write-Host ("Location: " + $FullPath) -ForegroundColor Cyan
}

Write-Host ""
