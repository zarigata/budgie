# Budgie - Distributed Container Orchestration

Budgie is a distributed container orchestration tool that simplifies running and replicating containers across machines in a local network.

## Version

**v0.1** - Initial release

## Features

- **Simple CLI**: Run containers with `budgie run example.bun`
- **LAN Discovery**: Find containers on local network with `budgie chirp`
- **Easy Replication**: Join a container as a peer with `budgie chirp <container-id>`
- **Load Balancing**: Port-based routing with automatic failover
- **Data Sync**: Automatic file synchronization across replicas
- **Interactive Setup**: `budgie nest` - wizard for new users and builders

## Installation

### Quick Install (Pre-built Binaries)

If you have pre-built binaries, run the installer:

**Linux/macOS:**
```bash
sudo ./install.sh
```

**Windows:**
```powershell
.\install.ps1
```

### Build from Source

```bash
# Build for current platform
make build

# Build for all platforms (requires Go 1.21+)
./build-all.sh

# Or manually:
make build-all
```

### Using Nest Wizard

For first-time setup or building for different platforms:
```bash
budgie nest
```

The wizard will:
- Detect your system (OS and architecture)
- Guide you through installation
- Help you choose build targets
- Teach you how to use budgie
- Check system compatibility

## Quick Start

### 1. Create a .bun file

```yaml
# example.bun
version: "1.0"
name: "webapp"

image:
  docker_image: "nginx:alpine"

ports:
  - container_port: 80
    host_port: 8080
    protocol: tcp

volumes:
  - source: "./data"
    target: "/app/data"
    mode: rw

environment:
  - APP_ENV=production

healthcheck:
  path: "/health"
  interval: 30s
  timeout: 5s
  retries: 3

replicas:
  min: 2
  max: 5
```

### 2. Run a container

```bash
budgie run example.bun
```

### 3. Discover containers on LAN

```bash
budgie chirp
```

### 4. Join as replica

```bash
budgie chirp abc123456789def
```

### 5. List local containers

```bash
budgie ps
```

## Available Binaries

Pre-built binaries are available for:
- **Linux**: `budgie-linux-amd64` or `budgie-linux-arm64`
- **macOS**: `budgie-darwin-amd64` (Intel) or `budgie-darwin-arm64` (Apple Silicon)
- **Windows**: `budgie-windows-amd64.exe` or `budgie-windows-arm64.exe`

Download the binary matching your system, make it executable (Linux/macOS), and run it.

## Installation Scripts

### `install.sh` (Linux/macOS)
- Detects platform automatically
- Copies binary to `/usr/local/bin/`
- Creates configuration in `/etc/budgie/config.yaml`
- Sets up data directory `/var/lib/budgie/`
- Creates log directory `/var/log/budgie/`

### `install.ps1` (Windows)
- Installs to `%LOCALAPPDATA%\budgie`
- Adds to user PATH
- Creates desktop shortcut
- Generates configuration file

## Build Scripts

### `build-all.sh`
Cross-platform build script that:
- Builds for Linux (amd64, arm64)
- Builds for macOS (amd64, arm64)
- Builds for Windows (amd64, arm64)
- Creates tar.gz packages for Linux/macOS
- Creates zip packages for Windows
- Optimized binaries with `-ldflags="-s -w"`

## Commands

| Command | Description |
|---------|-------------|
| `budgie run <file.bun>` | Run a container from .bun file |
| `budgie ps` | List all containers |
| `budgie stop <id>` | Stop a running container |
| `budgie chirp` | Discover containers on local network |
| `budgie chirp <id>` | Join a container as replica |
| `budgie nest` | Interactive setup and build wizard |

## Nest Wizard

The `budgie nest` command provides an interactive wizard for:

### 1. Quick Start 🚀
- Downloads Go modules
- Builds budgie for your system
- Shows next steps

### 2. Custom Build 🔨
- Choose target platform (Linux/macOS/Windows)
- Choose architecture (amd64/arm64)
- Build for single platform or all platforms

### 3. Learn Budgie 📚
- Running your first container
- Discovering containers on LAN
- Container replication
- Managing containers

### 4. System Check 🏥️
- Detects OS and architecture
- Checks Go version
- Verifies dependencies
- Shows compatibility status

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      budgie CLI                           │
│  ┌────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   run      │  │    chirp      │  │   ps/list    │   │
│  └─────┬──────┘  └──────┬───────┘  └──────┬───────┘   │
└────────┼────────────────┼──────────────────┼───────────────┘
         │                │                  │
         ▼                ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Budgie Daemon                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Container Manager  │  Node Discovery  │  Sync   │  │
│  └──────────────────┴─────────────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Technology Stack

- **Language**: Go 1.21+
- **CLI**: github.com/spf13/cobra
- **Container Runtime**: containerd
- **Service Discovery**: mDNS (hashicorp/mdns)
- **File Sync**: rsync algorithm (minio/rsync-go)
- **Load Balancing**: HTTP reverse proxy with round-robin

## Project Structure

```
budgie/
├── cmd/                # CLI commands
│   ├── root/          # Root command
│   ├── run/           # run command
│   ├── ps/            # ps command
│   └── chirp/         # chirp command
├── internal/           # Private application code
│   ├── api/           # HTTP/gRPC API
│   ├── bundle/        # .bun file parser
│   ├── cluster/       # Cluster management
│   ├── discovery/     # mDNS service discovery
│   ├── runtime/       # Container runtime wrapper
│   ├── sync/          # File synchronization
│   └── proxy/         # Load balancer proxy
├── pkg/               # Public libraries
│   └── types/        # Core data structures
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Development

```bash
# Install dependencies
make mod-tidy

# Run tests
make test

# Format code
make fmt

# Build for current platform
make build

# Build for all platforms
make build-all
```

## Roadmap

- [x] Project structure
- [x] Basic CLI framework
- [ ] .bun file parser
- [ ] Container runtime integration
- [ ] Container lifecycle management
- [ ] mDNS service discovery
- [ ] File synchronization
- [ ] Load balancing
- [ ] Advanced networking (DNS)

## License

MIT License - See LICENSE file for details
