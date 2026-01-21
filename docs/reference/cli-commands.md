# CLI Commands Reference

Complete reference for all Budgie CLI commands.

## Global Flags

Available on all commands:

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for command |
| `--version`, `-v` | Show version information |

## Commands

### budgie run

Run a container from a bun file.

```bash
budgie run [flags] <bundle.bun>
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--detach`, `-d` | Run in background |
| `--name` | Override container name |

**Examples:**
```bash
# Run interactively
budgie run myapp.bun

# Run in background
budgie run --detach myapp.bun

# Run with custom name
budgie run --name webserver nginx.bun
```

### budgie ps

List containers.

```bash
budgie ps [flags]
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--all`, `-a` | Show all containers (including stopped) |
| `--quiet`, `-q` | Only show container IDs |

**Examples:**
```bash
# List running containers
budgie ps

# List all containers
budgie ps --all

# List only IDs
budgie ps -q
```

### budgie stop

Stop a running container.

```bash
budgie stop [flags] <container-id>
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--timeout`, `-t` | Timeout before force kill (default: 10s) |

**Examples:**
```bash
# Stop with default timeout
budgie stop abc123

# Stop with 30 second timeout
budgie stop --timeout 30s abc123
```

### budgie chirp

Discover containers on network or join as peer.

```bash
budgie chirp [container-id]
```

**Without arguments:** Lists all discoverable containers on the local network.

**With container ID:** Joins that container as a peer/replica.

**Examples:**
```bash
# Discover containers
budgie chirp

# Join container as peer
budgie chirp abc123456789
```

**Aliases:** `discover`

### budgie nest

Interactive setup and build wizard.

```bash
budgie nest
```

Launches the interactive TUI for:
- System detection
- Platform selection
- Tutorials
- Container monitoring
- Dependency checking

**Aliases:** `setup`, `wizard`, `init`

## Container ID

Most commands accept either:
- Full 64-character container ID
- Short 12-character prefix

```bash
# Both work
budgie stop abc123456789012345678901234567890123456789012345678901234567890123
budgie stop abc123456789
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Command not found |
| 3 | Container not found |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CONTAINERD_ADDRESS` | containerd socket path (default: `/run/containerd/containerd.sock`) |
| `BUDGIE_LOG_LEVEL` | Log level: debug, info, warn, error |

## Configuration

Budgie reads configuration from:
1. Command line flags (highest priority)
2. Environment variables
3. Config file (if present)

Default config location: `~/.config/budgie/config.yaml`
