# Budgie Installation Script for Windows
# Usage: iex (irm https://raw.githubusercontent.com/zarigata/budgie/main/install.ps1)
# Requires: PowerShell 5.1+ and Administrator privileges

# --- Configuration ---
$GITHUB_REPO = "zarigata/budgie"
$INSTALL_DIR = "$env:LOCALAPPDATA\budgie"
$CONFIG_DIR = $INSTALL_DIR
$DATA_DIR = "$INSTALL_DIR\data"
$LOG_DIR = "$INSTALL_DIR\logs"

$BINARY_NAME = "budgie.exe"
$LATEST_RELEASE = (Invoke-RestMethod -Uri "https://api.github.com/repos/$GITHUB_REPO/releases/latest").tag_name

# --- Helper Functions ---
function Info {
    param([string]$message)
    Write-Host "INFO: $message" -ForegroundColor Blue
}

function Success {
    param([string]$message)
    Write-Host "✅ SUCCESS: $message" -ForegroundColor Green
}

function Error {
    param([string]$message)
    Write-Host "❌ ERROR: $message" -ForegroundColor Red
    exit 1
}

# --- Main Functions ---

function Check-Admin {
    if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Error "This script must be run as Administrator."
    }
}

function Detect-Arch {
    $env:PROCESSOR_ARCHITECTURE -match "64" | Out-Null
    if ($?) {
        $script:ARCH = "amd64"
    } else {
        $script:ARCH = "arm64"
    }
}

function Download-Binary {
    $binary_url = "https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/budgie-windows-$ARCH.exe"
    Info "Downloading Budgie from $binary_url..."
    Invoke-WebRequest -Uri $binary_url -OutFile "$INSTALL_DIR\$BINARY_NAME"
}

function Create-Directories {
    Info "Creating directories..."
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $DATA_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $LOG_DIR -Force | Out-Null
}

function Create-Config-File {
    Info "Creating configuration file..."
    $configContent = @"
# Budgie Configuration
data_dir: "$DATA_DIR"

runtime:
  address: "\.\pipe\containerd-containerd"

discovery:
  enabled: true
  port: 5353

proxy:
  type: "round-robin"
  health_check_interval: 30s

logging:
  level: "info"
  file: "$LOG_DIR\budgie.log"
"@
    Set-Content -Path "$CONFIG_DIR\config.yaml" -Value $configContent
}

function Add-To-Path {
    Info "Adding Budgie to your PATH..."
    $userPath = [System.Environment]::GetEnvironmentVariable('PATH', 'User')
    if ($userPath -notlike "*$INSTALL_DIR*") {
        $newPath = "$userPath;$INSTALL_DIR"
        [System.Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')
        Info "Budgie has been added to your PATH. Please restart your terminal."
    } else {
        Info "Budgie is already in your PATH."
    }
}

function Check-Containerd {
    Info "Checking for containerd..."
    $containerdService = Get-Service -Name containerd -ErrorAction SilentlyContinue
    if ($null -eq $containerdService) {
        Error "containerd is not installed. Please install it before running Budgie."
    }
    if ($containerdService.Status -ne "Running") {
        Error "containerd is not running. Please start it before running Budgie."
    }
}

function Install {
    Check-Admin
    Detect-Arch
    Check-Containerd

    Info "Installing Budgie $LATEST_RELEASE for Windows-$ARCH..."
    
    Create-Directories
    Download-Binary
    Create-Config-File
    Add-To-Path
    
    Success "Budgie has been installed successfully!"
    Info "Run 'budgie --help' in a new terminal to get started."
}

function Uninstall {
    Check-Admin
    Info "Uninstalling Budgie..."
    
    # Remove from PATH
    $userPath = [System.Environment]::GetEnvironmentVariable('PATH', 'User')
    $newPath = ($userPath.Split(';') | Where-Object { $_ -ne $INSTALL_DIR }) -join ';'
    [System.Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')

    # Remove files
    Remove-Item -Recurse -Force $INSTALL_DIR
    
    Success "Budgie has been uninstalled successfully."
}

# --- Script Entrypoint ---

param([string]$Action)

switch ($Action) {
    "uninstall" {
        Uninstall
    }
    default {
        Install
    }
}