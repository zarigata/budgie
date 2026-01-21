# Installation

This guide covers how to install Budgie on your system.

## Prerequisites

- Go 1.21 or later
- containerd 1.7 or later
- Linux, macOS, or Windows

## Installing from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/budgie/budgie.git
   cd budgie
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Build the binary:
   ```bash
   make build
   ```

4. Install (optional):
   ```bash
   sudo make install
   ```

## Cross-Compilation

Build for specific platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 make build

# Linux ARM64
GOOS=linux GOARCH=arm64 make build

# macOS AMD64
GOOS=darwin GOARCH=amd64 make build

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 make build

# Windows AMD64
GOOS=windows GOARCH=amd64 make build

# Build for all platforms
make build-all
```

## Verifying Installation

After installation, verify budgie works:

```bash
budgie --version
budgie nest  # Run interactive wizard
```

## containerd Setup

Budgie requires containerd to be running. On most Linux systems:

```bash
# Install containerd
sudo apt install containerd

# Start containerd
sudo systemctl start containerd
sudo systemctl enable containerd
```

## Next Steps

- Read the [Quick Start](quick-start.md) guide
- Learn about [.bun file format](../guides/bun-file-format.md)
- Explore [CLI commands](../reference/cli-commands.md)
