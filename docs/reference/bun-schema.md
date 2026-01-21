# Bun File Schema Reference

Complete schema reference for `.bun` files.

## Schema

```yaml
# Required: Bundle format version
version: string

# Optional: Container name (defaults to filename)
name: string

# Required: Image configuration
image:
  docker_image: string    # Required: Docker image reference
  command: [string]       # Optional: Override entrypoint
  workdir: string         # Optional: Working directory

# Required: At least one port mapping
ports:
  - container_port: integer  # Required: Port inside container
    host_port: integer       # Required: Port on host
    protocol: string         # Optional: "tcp" or "udp" (default: tcp)

# Optional: Volume mounts
volumes:
  - source: string        # Host path
    target: string        # Container path
    mode: string          # "rw" or "ro" (default: rw)

# Optional: Environment variables
environment:
  - string               # Format: "KEY=value"

# Optional: Resource limits
resources:
  cpu_shares: integer    # CPU shares (relative weight)
  cpu_quota: integer     # CPU quota in microseconds
  memory_limit: integer  # Memory limit in bytes
  memory_swap: integer   # Memory + swap limit in bytes
  blkio_weight: integer  # Block I/O weight (10-1000)
  pids_limit: integer    # Max number of PIDs

# Optional: Health check configuration
healthcheck:
  path: string           # HTTP path to check
  interval: duration     # Check interval (e.g., "30s")
  timeout: duration      # Request timeout (e.g., "5s")
  retries: integer       # Failures before unhealthy

# Optional: Replica configuration
replicas:
  min: integer           # Minimum replicas
  max: integer           # Maximum replicas
```

## Types

### string

A text value, quoted if containing special characters.

```yaml
name: "my-app"
name: simple_name
```

### integer

A whole number.

```yaml
ports:
  - container_port: 8080
    host_port: 8080
```

### duration

A time duration with unit suffix.

Valid units: `s` (seconds), `m` (minutes), `h` (hours)

```yaml
healthcheck:
  interval: 30s
  timeout: 5s
```

### [string]

An array of strings.

```yaml
image:
  command: ["npm", "start"]

environment:
  - "NODE_ENV=production"
  - "PORT=3000"
```

## Validation Rules

### Version

- Must be present
- Currently only "1.0" is supported

### Ports

- At least one port mapping required
- `container_port` and `host_port` must be positive integers
- `protocol` defaults to "tcp"

### Volumes

- `source` and `target` are required for each volume
- `mode` defaults to "rw"

### Resources

All resource fields are optional:
- `cpu_shares`: positive integer
- `cpu_quota`: positive integer (microseconds)
- `memory_limit`: positive integer (bytes)
- `memory_swap`: positive integer (bytes), >= memory_limit
- `blkio_weight`: integer 10-1000
- `pids_limit`: positive integer

### Health Check

All health check fields are optional:
- `path`: valid HTTP path
- `interval`: valid duration
- `timeout`: valid duration
- `retries`: positive integer

### Replicas

- `min`: non-negative integer
- `max`: positive integer, >= min

## Example: Complete Bundle

```yaml
version: "1.0"
name: "complete-example"

image:
  docker_image: "myregistry.io/myapp:v1.2.3"
  command: ["/app/server", "--config", "/etc/app/config.yaml"]
  workdir: "/app"

ports:
  - container_port: 8080
    host_port: 8080
    protocol: tcp
  - container_port: 8443
    host_port: 8443
    protocol: tcp
  - container_port: 9090
    host_port: 9090
    protocol: udp

volumes:
  - source: "/data/app"
    target: "/app/data"
    mode: rw
  - source: "/etc/app"
    target: "/etc/app"
    mode: ro
  - source: "./logs"
    target: "/var/log/app"
    mode: rw

environment:
  - "APP_ENV=production"
  - "LOG_LEVEL=info"
  - "DATABASE_URL=postgres://db:5432/myapp"
  - "REDIS_URL=redis://cache:6379"

resources:
  cpu_shares: 1024
  cpu_quota: 100000
  memory_limit: 536870912
  memory_swap: 1073741824
  blkio_weight: 500
  pids_limit: 200

healthcheck:
  path: "/health"
  interval: 30s
  timeout: 10s
  retries: 3

replicas:
  min: 2
  max: 10
```
