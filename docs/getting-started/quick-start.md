# Quick Start

Get up and running with Budgie in minutes.

## Your First Container

### 1. Create a Bundle File

Create a file named `myapp.bun`:

```yaml
version: "1.0"
name: "myapp"

image:
  docker_image: "nginx:alpine"
  workdir: "/usr/share/nginx/html"

ports:
  - container_port: 80
    host_port: 8080
    protocol: tcp

environment:
  - "NGINX_HOST=localhost"
```

### 2. Run the Container

```bash
budgie run myapp.bun
```

### 3. Access Your Application

Open your browser to `http://localhost:8080`

## Running in Background

Use the `--detach` flag to run containers in the background:

```bash
budgie run --detach myapp.bun
```

## Listing Containers

```bash
# List running containers
budgie ps

# List all containers (including stopped)
budgie ps --all
```

## Stopping Containers

```bash
# Stop gracefully (waits for SIGTERM)
budgie stop <container-id>

# Stop with custom timeout
budgie stop --timeout 30s <container-id>
```

## Discovering Containers on Network

Use chirp to find Budgie containers on your local network:

```bash
budgie chirp
```

Output:
```
CONTAINER ID   NAME     IP            PORT   IMAGE                  NODE
abc123456789   myapp    192.168.1.5   8080   nginx:alpine          laptop
def987654321   webapp   192.168.1.6   3000   node:18-alpine        desktop
```

## Joining a Container as Peer

To replicate a container on your machine:

```bash
budgie chirp abc123
```

This downloads the image and synchronizes volume data.

## Interactive Setup

For an interactive experience, use the nest command:

```bash
budgie nest
```

This launches a TUI wizard that guides you through:
- System detection
- Platform selection
- Tutorials and documentation

## Next Steps

- Learn about [resource limits](../guides/resource-limits.md)
- Understand [discovery and replication](../guides/discovery-replication.md)
- Explore the [bun file format](../guides/bun-file-format.md)
