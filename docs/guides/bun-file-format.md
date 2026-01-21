# Bun File Format

The `.bun` file is Budgie's container definition format. It's a YAML file that defines how to run your container.

## Basic Structure

```yaml
version: "1.0"
name: "myapp"

image:
  docker_image: "nginx:alpine"
  command: ["/bin/sh", "-c", "nginx -g 'daemon off;'"]
  workdir: "/app"

ports:
  - container_port: 80
    host_port: 8080
    protocol: tcp

volumes:
  - source: "./data"
    target: "/app/data"
    mode: rw

environment:
  - "NODE_ENV=production"
  - "DEBUG=false"

resources:
  cpu_shares: 512
  memory_limit: 134217728
  pids_limit: 100

healthcheck:
  path: "/health"
  interval: 30s
  timeout: 5s
  retries: 3

replicas:
  min: 1
  max: 5
```

## Fields Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Bundle format version (currently "1.0") |
| `ports` | array | At least one port mapping required |

### Image Configuration

```yaml
image:
  docker_image: "nginx:alpine"    # Required: Docker image reference
  command: ["nginx", "-g", "..."] # Optional: Override CMD
  workdir: "/app"                 # Optional: Working directory
```

### Port Mappings

```yaml
ports:
  - container_port: 80           # Port inside container
    host_port: 8080              # Port on host machine
    protocol: tcp                # tcp or udp
```

### Volume Mounts

```yaml
volumes:
  - source: "./data"             # Host path (relative or absolute)
    target: "/app/data"          # Container path
    mode: rw                     # rw (read-write) or ro (read-only)
```

### Environment Variables

```yaml
environment:
  - "KEY=value"
  - "ANOTHER_KEY=another_value"
```

### Resource Limits

```yaml
resources:
  cpu_shares: 512                # CPU shares (relative weight)
  cpu_quota: 50000               # CPU quota in microseconds
  memory_limit: 134217728        # Memory limit in bytes (128MB)
  memory_swap: 268435456         # Memory + swap limit
  blkio_weight: 500              # Block I/O weight (10-1000)
  pids_limit: 100                # Max number of processes
```

### Health Checks

```yaml
healthcheck:
  path: "/health"                # HTTP path to check
  interval: 30s                  # Check interval
  timeout: 5s                    # Request timeout
  retries: 3                     # Failures before unhealthy
```

### Replicas

```yaml
replicas:
  min: 1                         # Minimum replicas
  max: 5                         # Maximum replicas
```

## Examples

### Minimal Bundle

```yaml
version: "1.0"
name: "simple"

image:
  docker_image: "alpine:latest"

ports:
  - container_port: 8080
    host_port: 8080
    protocol: tcp
```

### Web Server

```yaml
version: "1.0"
name: "webserver"

image:
  docker_image: "nginx:alpine"

ports:
  - container_port: 80
    host_port: 80
    protocol: tcp
  - container_port: 443
    host_port: 443
    protocol: tcp

volumes:
  - source: "./html"
    target: "/usr/share/nginx/html"
    mode: ro
  - source: "./config/nginx.conf"
    target: "/etc/nginx/nginx.conf"
    mode: ro

resources:
  memory_limit: 268435456  # 256MB
  pids_limit: 50

healthcheck:
  path: "/"
  interval: 10s
  timeout: 3s
  retries: 3
```

### Node.js Application

```yaml
version: "1.0"
name: "node-api"

image:
  docker_image: "node:20-alpine"
  command: ["npm", "start"]
  workdir: "/app"

ports:
  - container_port: 3000
    host_port: 3000
    protocol: tcp

volumes:
  - source: "./src"
    target: "/app/src"
    mode: ro
  - source: "./data"
    target: "/app/data"
    mode: rw

environment:
  - "NODE_ENV=production"
  - "PORT=3000"

resources:
  cpu_shares: 1024
  memory_limit: 536870912  # 512MB
  pids_limit: 200

replicas:
  min: 2
  max: 10
```

## Validation

Budgie validates bun files on load. Common errors:

- Missing `version` field
- No port mappings defined
- Invalid resource values
- Malformed YAML syntax
