# build-prod.ps1 -- Full production build script for Photobox
# Usage: .\build-prod.ps1

$root      = Split-Path -Parent $MyInvocation.MyCommand.Path
$binDir    = Join-Path $root "build\bin"
$bridgeSrc = Join-Path $root "DSLRBridge\bin\x86\Debug"

Write-Host ""
Write-Host "[1/2] Building Photobox (wails build --webview2 embed)..." -ForegroundColor Cyan
wails build --webview2 embed
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] wails build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "[2/2] Copying DSLRBridge assets to build\bin ..." -ForegroundColor Cyan

$requiredFiles = @(
    "DSLRBridge.exe",
    "Accord.dll",
    "Accord.Video.DirectShow.dll",
    "Accord.Video.dll",
    "CameraControl.Devices.dll",
    "Canon.Eos.Framework.dll",
    "EDSDK.dll",
    "Interop.PortableDeviceApiLib.dll",
    "Interop.PortableDeviceTypesLib.dll",
    "log4net.dll",
    "Newtonsoft.Json.dll",
    "PortableDeviceLib.dll",
    "Rssdp.Native.dll",
    "websocket-sharp.dll",
    "wiaaut.dll"
)

foreach ($file in $requiredFiles) {
    $src = Join-Path $bridgeSrc $file
    $dst = Join-Path $binDir $file
    if (Test-Path $src) {
        Copy-Item $src $dst -Force
        Write-Host "  [OK] $file" -ForegroundColor Green
    } else {
        Write-Host "  [SKIP] $file not found at $src" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "[DONE] Build complete! Output: $binDir\photobox-init.exe" -ForegroundColor Green
Write-Host ""
