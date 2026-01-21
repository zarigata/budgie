# Architecture Overview

Budgie is a lightweight container management system with built-in discovery and replication.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Budgie CLI                           │
├─────────────────────────────────────────────────────────────┤
│  run  │  ps  │  stop  │  chirp  │  nest                    │
└───┬───┴──────┴────────┴────┬────┴──────────────────────────┘
    │                        │
    ▼                        ▼
┌─────────────┐      ┌─────────────────┐
│   Bundle    │      │    Discovery    │
│   Parser    │      │    Service      │
└──────┬──────┘      └────────┬────────┘
       │                      │
       ▼                      ▼
┌─────────────┐      ┌─────────────────┐
│  Container  │      │     mDNS        │
│  Runtime    │      │   (hashicorp)   │
└──────┬──────┘      └─────────────────┘
       │
       ▼
┌─────────────┐      ┌─────────────────┐
│ containerd  │◄────►│   Sync Server   │
│             │      │   (TCP 18733)   │
└─────────────┘      └─────────────────┘
```

## Components

### CLI Layer

The command-line interface provides user interaction:

- **cmd/budgie/main.go**: Entry point
- **cmd/run/run.go**: Container execution
- **cmd/ps/ps.go**: Container listing
- **cmd/stop/stop.go**: Container stopping
- **cmd/chirp/chirp.go**: Discovery and replication
- **cmd/nest/nest.go**: Interactive TUI

### Bundle Parser

Parses `.bun` files into container configurations:

- **internal/bundle/bundle.go**: YAML parsing
- Validates required fields
- Converts to internal container types

### Container Runtime

Manages container lifecycle via containerd:

- **internal/runtime/containerd.go**: containerd integration
- Image pulling
- Container creation with resource limits
- Volume mounting
- Graceful shutdown (SIGTERM → SIGKILL)

### Discovery Service

mDNS-based container discovery:

- **internal/discovery/mdns.go**: mDNS integration
- Announces containers on start
- Discovers containers on LAN
- Uses `_budgie._tcp` service type

### Sync System

Volume synchronization for replication:

- **internal/sync/volume.go**: File sync logic
- **internal/sync/server.go**: TCP sync server
- **internal/sync/protocol.go**: Wire protocol
- Delta-sync for efficiency

### Load Balancer

Request distribution for replicas:

- **internal/proxy/loadbalancer.go**: Reverse proxy
- Round-robin and least-connections algorithms
- Health checking

### UI Package

Terminal user interface components:

- **internal/ui/styles.go**: Lipgloss styling
- **internal/ui/menu.go**: Interactive menus
- **internal/ui/table.go**: Data tables
- **internal/ui/monitor.go**: Real-time dashboard
- **internal/ui/progress.go**: Progress bars

## Data Flow

### Running a Container

```
1. User runs: budgie run myapp.bun
2. Bundle parser reads and validates YAML
3. Container config created with resource limits
4. Runtime pulls image via containerd
5. Container created with volumes and env vars
6. Task started with proper cgroup limits
7. Discovery service announces container
8. Sync server starts for volume replication
```

### Discovering Containers

```
1. User runs: budgie chirp
2. Discovery service sends mDNS query
3. All Budgie nodes respond with container info
4. Results collected and deduplicated
5. Displayed in tabular format
```

### Replicating a Container

```
1. User runs: budgie chirp <container-id>
2. Target container discovered via mDNS
3. Connection established to primary node
4. Image information retrieved
5. Image pulled locally
6. Sync client connects to sync server
7. File signatures exchanged
8. Delta sync transfers only changed files
9. Local container started as replica
```

## Type System

### Core Types (pkg/types/container.go)

```go
type Container struct {
    ID        string
    Name      string
    State     ContainerState
    Image     ImageConfig
    Ports     []PortMapping
    Volumes   []VolumeMapping
    Env       []string
    Resources *ResourceLimits
    Health    *HealthCheck
    Replicas  *ReplicasConfig
}
```

### Resource Limits

```go
type ResourceLimits struct {
    CPUShares   int64
    CPUQuota    int64
    MemoryLimit int64
    MemorySwap  int64
    BlkioWeight uint16
    PidsLimit   int64
}
```

## Networking

### Ports Used

| Port | Protocol | Purpose |
|------|----------|---------|
| 5353 | UDP | mDNS discovery |
| 18733 | TCP | Volume sync |
| User-defined | TCP/UDP | Container ports |

### Discovery Protocol

Uses standard mDNS with TXT records:
- `container_id`: Full container ID
- `node_id`: Hostname of running node
- `container_name`: Human-readable name
- `image`: Docker image reference

## Security Considerations

1. **Container Isolation**: Uses containerd's default isolation
2. **Resource Limits**: Enforced via cgroups
3. **Network**: mDNS limited to local network
4. **Sync**: Currently unencrypted (planned: TLS)

## Dependencies

- **containerd**: Container runtime
- **hashicorp/mdns**: mDNS library
- **spf13/cobra**: CLI framework
- **charmbracelet/bubbletea**: TUI framework
- **fsnotify**: File watching for sync
