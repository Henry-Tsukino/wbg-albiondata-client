$ErrorActionPreference = "Stop"

Write-Host "=== WBG Albion Data Client Update Build ===" -ForegroundColor Green

# Import the build script
& ".\build-windows-custom.ps1"

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Creating compressed update file..." -ForegroundColor Cyan

$ExeName = "WBG-albion-data-client.exe"
$OutputGzName = "update-windows-amd64.exe.gz"

# Check if 7-Zip is available
$SevenZipPath = "C:\Program Files\7-Zip\7z.exe"
if (-Not (Test-Path $SevenZipPath)) {
    $SevenZipPath = "C:\Program Files (x86)\7-Zip\7z.exe"
}

if (-Not (Test-Path $SevenZipPath)) {
    Write-Host "ERROR: 7-Zip not found. Installing via PowerShell compression instead..." -ForegroundColor Yellow
    
    # Use PowerShell native compression as fallback
    if ($PSVersionTable.PSVersion.Major -ge 5) {
        Write-Host "Using PowerShell compression..." -ForegroundColor Cyan
        # PowerShell doesn't have native gzip support, use 7z if available or inform user
        Write-Host "ERROR: 7-Zip is required for .gz compression" -ForegroundColor Red
        Write-Host "Please install 7-Zip or specify path to 7z.exe" -ForegroundColor Yellow
        exit 1
    }
} else {
    Write-Host "Using 7-Zip at: $SevenZipPath" -ForegroundColor Cyan
}

# Remove old gz file if exists
if (Test-Path $OutputGzName) {
    Remove-Item -Path $OutputGzName -Force
    Write-Host "Removed old file: $OutputGzName" -ForegroundColor Yellow
}

# Create .gz archive with 7-Zip
Write-Host "Compressing $ExeName to $OutputGzName..." -ForegroundColor Cyan
& $SevenZipPath a -tgzip "$OutputGzName" "$ExeName"

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Compression failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Update Build Successful! ===" -ForegroundColor Green

if (Test-Path $OutputGzName) {
    $FileSize = (Get-Item $OutputGzName).Length / 1MB
    Write-Host ("Compressed Size: " + [Math]::Round($FileSize, 2) + " MB") -ForegroundColor Yellow
    $FullPath = (Get-Item $OutputGzName).FullName
    Write-Host ("Location: " + $FullPath) -ForegroundColor Cyan
    
    # Show compression ratio
    if (Test-Path $ExeName) {
        $OriginalSize = (Get-Item $ExeName).Length / 1MB
        $CompressionRatio = ([Math]::Round($FileSize / $OriginalSize, 2) * 100)
        Write-Host ("Original Size: " + [Math]::Round($OriginalSize, 2) + " MB") -ForegroundColor Yellow
        Write-Host ("Compression Ratio: " + $CompressionRatio + "%") -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Ready for update deployment!" -ForegroundColor Green
