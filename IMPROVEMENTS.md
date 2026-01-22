# Budgie Improvements Summary

## Completed Features

### 1. Fixed `budgie ps` command (Critical)
- **File**: `cmd/ps/ps.go`
- Now reads actual container state from `/var/lib/budgie/state.json`
- Shows real container information instead of hardcoded examples
- Added `--quiet` flag for script-friendly output
- Improved formatting with status, ports, and timestamps

### 2. Implemented `budgie chirp <id>` replication (Critical)
- **File**: `cmd/chirp/chirp.go`
- Full replication workflow:
  1. Discovers target container on network via mDNS
  2. Initializes local runtime and manager
  3. Creates replica container configuration
  4. Pulls the same image from registry
  5. Syncs volumes from primary node (optional)
  6. Starts and announces replica on network
- Added `--sync` flag for volume synchronization
- Added `--dry-run` flag for preview mode

### 3. Added `budgie rm` command (Critical)
- **File**: `cmd/rm/rm.go`
- Remove one or more containers
- `--force` flag to remove running containers
- `--volumes` flag to also remove associated volumes
- Supports ID prefix matching and name lookup
- Aliases: `remove`, `delete`

### 4. Added `budgie logs` command (High Priority)
- **File**: `cmd/logs/logs.go`
- View container stdout/stderr logs
- `--follow` flag for real-time streaming
- `--tail` flag to show last N lines
- `--timestamps` flag for timestamp display

### 5. Added `budgie exec` command (High Priority)
- **File**: `cmd/exec/exec.go`
- Execute commands inside running containers
- `-i` flag for interactive mode
- `-t` flag for TTY allocation
- `-u` flag for user specification
- `-w` flag for working directory

### 6. Added `budgie inspect` command (High Priority)
- **File**: `cmd/inspect/inspect.go`
- Display detailed container information as JSON
- Compatible with Docker inspect format
- Shows state, config, network, mounts, and Budgie-specific info

### 7. Added Configuration File Support (High Priority)
- **Files**: `internal/config/config.go`, `cmd/config/config.go`
- Configuration file locations:
  - `./budgie.yaml`
  - `~/.budgie/budgie.yaml`
  - `~/.config/budgie/budgie.yaml`
  - `/etc/budgie/budgie.yaml`
- Environment variable overrides
- Subcommands:
  - `budgie config` - Show current configuration
  - `budgie config init` - Create default config file
  - `budgie config path` - Show config file location
  - `budgie config get <key>` - Get specific value

### 8. Added Container Restart Policies (High Priority)
- **Files**: `internal/api/restart.go`, `internal/bundle/bundle.go`, `pkg/types/container.go`
- Supported policies:
  - `no` - Never restart
  - `always` - Always restart
  - `on-failure` - Restart on non-zero exit
  - `unless-stopped` - Restart unless explicitly stopped
- Configurable maximum retry count
- Exponential backoff for restarts
- Restart count tracking

### 9. Added TLS Encryption for Sync Protocol (High Priority)
- **File**: `internal/sync/tls.go`
- TLS 1.2+ encryption for volume sync
- Self-signed certificate generation
- Certificate/key file configuration
- CA verification support
- `TLSServer` and `TLSClient` wrappers

### 10. Added `budgie pull <image>` command (Medium Priority) - COMPLETED
- **File**: `cmd/pull/pull.go`
- Pre-pull images from registries before running containers
- Normalizes image names (adds docker.io/library/ if needed)
- Supports `--quiet` flag for script-friendly output
- Supports `--platform` flag for multi-platform images

### 11. Added `budgie images` command (Medium Priority) - COMPLETED
- **File**: `cmd/images/images.go`
- List locally stored images
- Supports `--digests`, `--no-trunc`, `--quiet`, `--format` flags
- Filter by repository name
- Shows repository, tag, image ID, creation time, and size

### 12. Added Health Check Integration (Medium Priority) - COMPLETED
- **File**: `internal/api/healthcheck.go`
- Monitors container health via HTTP health checks
- Auto-restarts unhealthy containers (integrates with restart policy)
- Tracks health status: healthy, unhealthy, starting
- Configurable retry count and timeout

### 13. Added Container Dependencies (Medium Priority) - COMPLETED
- **File**: `internal/api/dependency.go`
- `depends_on` field in container configuration
- Topological sort for startup ordering
- Cycle detection for circular dependencies
- Wait for dependencies before starting container

### 14. Added Network Isolation (Medium Priority) - COMPLETED
- **Files**: `internal/network/network.go`, `cmd/network/network.go`
- Container network groups with IP allocation
- Default `budgie0` network (172.20.0.0/16)
- Commands: `budgie network ls`, `create`, `rm`, `inspect`
- Connect/disconnect containers from networks

### 15. Added Secrets Management (Medium Priority) - COMPLETED
- **Files**: `internal/secrets/secrets.go`, `cmd/secret/secret.go`
- AES-GCM encrypted secret storage
- PBKDF2 key derivation with salt
- Commands: `budgie secret create`, `ls`, `rm`, `inspect`
- Secure file permissions (0600/0700)

## Additional Improvements

### Environment File Support
- **File**: `internal/bundle/bundle.go`
- `env_file` field in .bun files
- Loads KEY=VALUE pairs from file
- Supports comments with `#`
- Bundle environment overrides file environment

### Extended Runtime Interface
- **File**: `internal/runtime/containerd.go`
- Added `Logs()` method for container log access
- Added `Exec()` method for command execution
- `LogReader` interface for streaming logs

### Updated Container Types
- **File**: `pkg/types/container.go`
- Added `RestartPolicy` struct
- Added `RestartCount` field
- Added restart policy to Container struct

### Updated Example Bundle
- **File**: `example.bun`
- Shows restart_policy configuration
- Shows stop_timeout setting
- Shows resource limits
- Documents env_file option

## New CLI Commands Summary

| Command | Description |
|---------|-------------|
| `budgie ps` | List containers (fixed) |
| `budgie rm` | Remove containers |
| `budgie logs` | View container logs |
| `budgie exec` | Execute commands in containers |
| `budgie inspect` | Show detailed container info |
| `budgie config` | Manage configuration |
| `budgie chirp <id>` | Join as replica (fixed) |
| `budgie pull <image>` | Pre-pull images from registry |
| `budgie images` | List locally stored images |
| `budgie network` | Manage container networks |
| `budgie secret` | Manage encrypted secrets |

## Files Modified

- `cmd/ps/ps.go` - Complete rewrite
- `cmd/chirp/chirp.go` - Full implementation
- `cmd/stop/stop.go` - Added missing function
- `cmd/root/main.go` - Added new commands
- `internal/runtime/containerd.go` - Extended interface
- `internal/bundle/bundle.go` - Added restart_policy, env_file
- `internal/sync/volume.go` - Fixed imports
- `pkg/types/container.go` - Added RestartPolicy
- `example.bun` - Updated with new features

## Files Created

- `cmd/rm/rm.go` - Remove command
- `cmd/logs/logs.go` - Logs command
- `cmd/exec/exec.go` - Exec command
- `cmd/inspect/inspect.go` - Inspect command
- `cmd/config/config.go` - Config command
- `cmd/pull/pull.go` - Pull command
- `cmd/images/images.go` - Images command
- `cmd/network/network.go` - Network command
- `cmd/secret/secret.go` - Secret command
- `internal/config/config.go` - Configuration package
- `internal/api/restart.go` - Restart monitor
- `internal/api/healthcheck.go` - Health check monitor
- `internal/api/dependency.go` - Dependency resolver
- `internal/network/network.go` - Network manager
- `internal/secrets/secrets.go` - Secrets manager
- `internal/sync/tls.go` - TLS support

## Remaining Features (Not Yet Implemented)

16. Web dashboard - Browser UI for container management

## Building

```bash
# Build all targets
make all

# Build for current platform
make build

# Run tests
make test
```

## Testing the New Features

```bash
# List containers
budgie ps --all

# Remove a container
budgie rm <container-id>

# View logs
budgie logs -f <container-id>

# Execute command
budgie exec <container-id> /bin/sh

# Inspect container
budgie inspect <container-id>

# View configuration
budgie config

# Initialize config
budgie config init

# Join as replica
budgie chirp --sync <container-id>

# Pull an image
budgie pull nginx:latest

# List images
budgie images

# Network management
budgie network ls
budgie network create mynetwork --subnet 172.22.0.0/16
budgie network inspect mynetwork
budgie network rm mynetwork

# Secrets management
echo "mysecret" | budgie secret create my-secret
budgie secret ls
budgie secret inspect my-secret
budgie secret rm my-secret
```
