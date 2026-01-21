# Budgie Installer for Windows
# Usage: irm https://zarigata.github.io/budgie/install.ps1 | iex
#
# This script downloads and installs the latest version of Budgie

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$VERSION = "0.1.0"
$REPO = "zarigata/budgie"
$BINARY_NAME = "budgie"
$INSTALL_DIR = "$env:LOCALAPPDATA\budgie"

function Write-Banner {
    Write-Host @"

    ____            __     _
   / __ )__  ______/ /__  (_)__
  / __  / / / / __  / _ \/ / _ \
 / /_/ / /_/ / /_/ /  __/ /  __/
/_____/\__,_/\__,_/\___/_/\___/

"@ -ForegroundColor Cyan
    Write-Host "Budgie Installer v$VERSION" -ForegroundColor Green
    Write-Host "==================================" -ForegroundColor Green
    Write-Host ""
}

function Write-Info {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "Warning: $Message" -ForegroundColor Yellow
}

function Write-Error-Exit {
    param([string]$Message)
    Write-Host "Error: $Message" -ForegroundColor Red
    exit 1
}

function Get-Platform {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" }
            elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" }
            else { Write-Error-Exit "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }

    Write-Info "Detected platform: windows-$arch"
    return $arch
}

function Install-Budgie {
    param([string]$Arch)

    $binary = "$BINARY_NAME-windows-$Arch.exe"
    $downloadUrl = "https://github.com/$REPO/releases/download/v$VERSION/$BINARY_NAME-windows-$Arch.zip"
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

    Write-Info "Downloading Budgie v$VERSION..."

    try {
        # Try to download pre-built binary
        $zipPath = Join-Path $tempDir "budgie.zip"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        $binaryPath = Get-ChildItem -Path $tempDir -Filter "*.exe" -Recurse | Select-Object -First 1
        if (-not $binaryPath) {
            throw "Binary not found in archive"
        }
    }
    catch {
        Write-Warn "Pre-built binary not available. Attempting to build from source..."

        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Write-Error-Exit "Go is required to build from source. Install Go 1.21+ and try again."
        }

        Write-Info "Cloning repository..."
        $srcDir = Join-Path $tempDir "budgie-src"
        git clone --depth 1 "https://github.com/$REPO.git" $srcDir

        Write-Info "Building Budgie..."
        Push-Location $srcDir
        go build -ldflags="-s -w" -o (Join-Path $tempDir "budgie.exe") ./cmd/budgie
        Pop-Location

        $binaryPath = Get-Item (Join-Path $tempDir "budgie.exe")
    }

    # Create installation directory
    Write-Info "Installing to $INSTALL_DIR..."
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    }

    Copy-Item $binaryPath.FullName -Destination (Join-Path $INSTALL_DIR "$BINARY_NAME.exe") -Force

    # Cleanup
    Remove-Item $tempDir -Recurse -Force

    Write-Success "Binary installed to: $INSTALL_DIR\$BINARY_NAME.exe"
}

function Add-ToPath {
    Write-Info "Configuring PATH..."

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$INSTALL_DIR*") {
        $newPath = "$currentPath;$INSTALL_DIR"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Success "Added to user PATH"
        Write-Warn "Please restart your terminal for PATH changes to take effect"
    }
    else {
        Write-Success "Already in PATH"
    }
}

function New-Config {
    $configPath = Join-Path $INSTALL_DIR "config.yaml"

    if (-not (Test-Path $configPath)) {
        Write-Info "Creating default configuration..."

        $configContent = @"
# Budgie Configuration
# https://github.com/zarigata/budgie

data_dir: "$INSTALL_DIR\data"

runtime:
  address: "\\.\pipe\containerd-containerd"

discovery:
  enabled: true
  port: 5353
  service_name: "_budgie._tcp"

proxy:
  type: "round-robin"
  health_check_interval: 30s
  health_check_timeout: 5s

sync:
  enabled: true
  port: 9876

logging:
  level: "info"
  file: "$INSTALL_DIR\budgie.log"
"@

        Set-Content -Path $configPath -Value $configContent -Encoding UTF8
        Write-Success "Configuration created: $configPath"
    }
    else {
        Write-Warn "Configuration already exists, skipping..."
    }
}

function New-DesktopShortcut {
    Write-Info "Creating desktop shortcut..."

    $desktopPath = [Environment]::GetFolderPath("Desktop")
    $shortcutPath = Join-Path $desktopPath "Budgie.lnk"

    $WScriptShell = New-Object -ComObject WScript.Shell
    $shortcut = $WScriptShell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
    $shortcut.Arguments = "nest"
    $shortcut.WorkingDirectory = $INSTALL_DIR
    $shortcut.Description = "Budgie - Distributed Container Orchestration"
    $shortcut.Save()

    Write-Success "Desktop shortcut created"
}

function Test-Installation {
    $budgiePath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"

    if (Test-Path $budgiePath) {
        Write-Success "Budgie installed successfully!"
    }
    else {
        Write-Error-Exit "Installation verification failed"
    }
}

function Write-Instructions {
    Write-Host ""
    Write-Host "Quick Start:" -ForegroundColor Cyan
    Write-Host "  1. Open a new terminal (restart needed for PATH)"
    Write-Host "  2. Run the setup wizard:  budgie nest"
    Write-Host "  3. Or run a container:    budgie run example.bun"
    Write-Host "  4. Discover on LAN:       budgie chirp"
    Write-Host "  5. Get help:              budgie --help"
    Write-Host ""
    Write-Host "Documentation:" -ForegroundColor Cyan
    Write-Host "  https://zarigata.github.io/budgie/"
    Write-Host ""
    Write-Host "Requirements:" -ForegroundColor Cyan
    Write-Host "  - containerd for Windows (container runtime)"
    Write-Host "  - Windows 10/11 or Windows Server 2019+"
    Write-Host ""
}

# Main installation flow
function Main {
    Write-Banner

    $arch = Get-Platform
    Install-Budgie -Arch $arch
    Add-ToPath
    New-Config
    New-DesktopShortcut
    Test-Installation
    Write-Instructions
}

Main
