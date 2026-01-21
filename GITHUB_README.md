# 🐦 Budgie

<div align="center">

![Budgie Logo](https://img.shields.io/badge/Budgie-v0.1-blue)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8E1blue)
[![License](https://img.shields.io/badge/License-MIT-green)
[![Release](https://img.shields.io/github/v/budgie/budgie)
[![CI/CD](https://img.shields.io/github/actions/budgie/budgie/workflow/Build%20Budgie/badge.svg)]
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-blue)

*A simple yet powerful distributed container orchestration tool for local networks*

</div>

## 📖 Table of Contents

- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [📦 Installation](#-installation)
- [📚 Documentation](#-documentation)
- [🎯 Roadmap](#-roadmap)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

## ✨ Features

### Core Capabilities
- **🐋 Docker-like CLI**: Simple commands to run, list, stop containers
- **📡 LAN Discovery**: Find containers on local network via mDNS (`budgie chirp`)
- **🔄 Easy Replication**: Join containers as peers with one command
- **⚖️ Load Balancing**: Round-robin and least-connections algorithms
- **📁 Data Sync**: Automatic file synchronization using rsync algorithm
- **⚡ Health Checks**: Configurable health monitoring with automatic failover
- **🎯 Port-Based Routing**: Resilient routing that doesn't depend on fixed IPs

### Unique Features
- **🏠️ Nest Wizard**: Interactive setup for new users (`budgie nest`)
- **📦 .bun Format**: Declarative container definitions (YAML-based)
- **🌐 Multi-Platform**: Support for Linux, macOS (Intel & ARM), Windows
- **💾 State Persistence**: Automatic state saving and recovery
- **🛡️ Containerd Runtime**: Modern, fast container engine

## 🚀 Quick Start

### 1. Installation

```bash
# Linux/macOS
curl -L https://github.com/budgie/budgie/releases/latest/download/install.sh | sh

# Windows (PowerShell)
irm https://github.com/budgie/budgie/releases/latest/download/install.ps1 | iex
```

### 2. Run Setup Wizard

```bash
budgie nest
```

The wizard will guide you through:
- ✅ System detection (OS, architecture)
- ✅ Build budgie for your platform
- ✅ First container tutorial
- ✅ Troubleshooting

### 3. Run Your First Container

Create `example.bun`:

```yaml
version: "1.0"
name: "myapp"
image:
  docker_image: "nginx:alpine"
ports:
  - container_port: 80
    host_port: 8080
    protocol: tcp
```

Run it:

```bash
budgie run example.bun
```

### 4. Discover on Network

```bash
budgie chirp
```

Output:
```
CONTAINER ID   NAME    IP           PORT  STATUS
abc123456789   webapp  192.168.1.5  8080  Running
```

## 📦 Installation

### Automated Installation

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/budgie/budgie/v0.1/install.sh -o install.sh
sh install.sh
```

**Windows:**
```powershell
irm https://raw.githubusercontent.com/budgie/budgie/v0.1/install.ps1 -o install.ps1
.\install.ps1
```

### Manual Installation

1. Download appropriate binary:
   - `budgie-linux-amd64` or `budgie-linux-arm64` (Linux)
   - `budgie-darwin-amd64` or `budgie-darwin-arm64` (macOS)
   - `budgie-windows-amd64.exe` or `budgie-windows-arm64.exe` (Windows)

2. Make executable (Linux/macOS):
```bash
chmod +x budgie-*
```

3. Move to PATH (Linux/macOS):
```bash
sudo mv budgie-* /usr/local/bin/budgie
```

4. Add to PATH (Windows):
   - Extract to `%LOCALAPPDATA%\budgie`
   - Add to user PATH

5. Verify installation:
```bash
budgie --version
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/budgie/budgie.git
cd budgie

# Build for current platform
make build

# Or build all platforms (requires Go 1.21+)
./build-all.sh

# Or use Makefile targets
make build-all     # All platforms
make build-linux   # Linux only
make build-darwin  # macOS only
make build-windows # Windows only
```

## 📚 Documentation

### Commands

| Command | Description | Example |
|---------|-------------|---------|
| `budgie run <file.bun>` | Create and start container from .bun file | `budgie run myapp.bun` |
| `budgie run -d <file.bun>` | Run in background (detached) | `budgie run -d myapp.bun` |
| `budgie ps` | List all containers | `budgie ps` |
| `budgie ps --all` | List including stopped | `budgie ps -a` |
| `budgie stop <id>` | Stop a running container | `budgie stop abc123...` |
| `budgie stop -t 30s <id>` | Stop with timeout | `budgie stop -t 30s abc...` |
| `budgie chirp` | Discover containers on LAN | `budgie chirp` |
| `budgie chirp <id>` | Join container as peer/replica | `budgie chirp abc123...` |
| `budgie nest` | Interactive setup wizard | `budgie nest` |

### .bun File Format

```yaml
version: "1.0"                    # Bundle format version
name: "container-name"              # Friendly name

image:
  docker_image: "nginx:alpine"     # Docker image to use
  command: ["/bin/sh", "-c"]        # Override command
  workdir: "/app"                   # Working directory

ports:                               # Port mappings
  - container_port: 80              # Container port
    host_port: 8080                # Host port
    protocol: tcp                     # tcp or udp

volumes:                              # Volume mounts
  - source: "./local/path"          # Host path
    target: "/container/path"           # Container path
    mode: rw                        # rw or ro

environment:                           # Environment variables
  - KEY=value
  - ANOTHER_KEY=another_value

healthcheck:                          # Optional health check
  path: "/health"
  interval: 30s
  timeout: 5s
  retries: 3

replicas:                             # Optional replica config
  min: 2
  max: 5
```

### Configuration

Default locations:
- **Linux/macOS**: `/etc/budgie/config.yaml`, `/var/lib/budgie/`
- **Windows**: `%LOCALAPPDATA%\budgie\config.yaml`, `%LOCALAPPDATA%\budgie\data\`

Configuration template:
```yaml
data_dir: "/var/lib/budgie"
runtime:
  address: "/run/containerd/containerd.sock"

discovery:
  enabled: true
  port: 5353

proxy:
  type: "round-robin"  # or "least-connections"
  health_check_interval: 30s

logging:
  level: "info"  # debug, info, warn, error
  file: "/var/log/budgie/budgie.log"
```

### Architecture

```
budgie/
├── cmd/                 # CLI commands
│   ├── nest/         # Setup wizard
│   ├── root/         # Root command
│   ├── run/          # Run containers
│   ├── ps/           # List containers
│   ├── stop/         # Stop containers
│   └── chirp/        # Network discovery
├── internal/            # Internal packages
│   ├── api/           # Container lifecycle
│   ├── bundle/        # .bun parser
│   ├── discovery/     # mDNS service
│   ├── runtime/       # containerd wrapper
│   ├── sync/          # File synchronization
│   └── proxy/         # Load balancer
├── pkg/               # Public packages
│   └── types/        # Core data structures
├── .github/           # GitHub Actions workflows
├── build-all.sh       # Cross-platform build
├── install.sh         # Linux/macOS installer
├── install.ps1        # Windows installer
└── go.mod            # Go modules
```

## 🎯 Roadmap

### v0.2 (Next Release)
- [ ] Add `budgie rm` command for container removal
- [ ] Add `budgie logs` command for log viewing
- [ ] Add `budgie exec` for running commands in containers
- [ ] Implement full replication in `chirp join`
- [ ] Add configuration validation
- [ ] Add systemd service (Linux)
- [ ] Add launchd service (macOS)
- [ ] Add Windows service

### Future Features
- [ ] Web UI for container management
- [ ] Metrics dashboard and monitoring
- [ ] Integration with container registries
- [ ] Support for Docker Compose files
- [ ] Automatic updates
- [ ] Advanced networking (DNS)
- [ ] Raft consensus for leader election

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/budgie/budgie.git
cd budgie

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make fmt
make lint
```

### Code Style

- Follow Go best practices and idiomatic code
- Use `gofmt` for formatting
- Write tests for new features
- Update documentation for API changes

### Reporting Issues

When reporting issues, please include:
- Budgie version (`budgie --version`)
- OS and architecture
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behavior
- Logs if available

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [containerd](https://containerd.io/) - High-performance container runtime
- Uses [cobra](https://github.com/spf13/cobra) - CLI framework
- Uses [hashicorp/mdns](https://github.com/hashicorp/mdns) - mDNS service discovery
- Uses [minio/rsync-go](https://github.com/minio/rsync-go) - rsync algorithm

## 📞 Support

- **GitHub Issues**: [https://github.com/budgie/budgie/issues](https://github.com/budgie/budgie/issues)
- **Documentation**: [https://budgie.dev/docs](https://budgie.dev/docs)
- **Discord**: [Join our community](https://discord.gg/budgie)

## 🌟 Star Us!

If you find Budgie useful, please consider giving us a star on [GitHub](https://github.com/budgie/budgie)!

<div align="center">

Made with ❤️ by the Budgie community

</div>
