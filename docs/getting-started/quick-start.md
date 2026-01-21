# Quick Start

This guide will walk you through the basics of using Budgie to run and manage containers on your local network.

## 1. Installation

First, you need to install the Budgie CLI.

**Linux / macOS:**
```bash
curl -fsSL https://zarigata.github.io/budgie/install.sh | sudo bash
```

**Windows (PowerShell as Admin):**
```powershell
irm https://zarigata.github.io/budgie/install.ps1 | iex
```

## 2. Create a `.bun` file

Budgie uses `.bun` files to define containers. Create a file named `webapp.bun` with the following content:

```yaml
version: "1.0"
name: "webapp"

image:
  docker_image: "nginx:alpine"

ports:
  - container_port: 80
    host_port: 8080

volumes:
  - source: "./html"
    target: "/usr/share/nginx/html"
    mode: rw

healthcheck:
  path: "/"
  interval: 30s
  timeout: 5s
  retries: 3

replicas:
  min: 2
  max: 5
```

This file defines a simple `nginx` container with a port mapping, a volume, a health check, and a replication policy.

## 3. Run the container

Now, you can run the container using the `budgie run` command:

```bash
budgie run webapp.bun
```

Budgie will pull the `nginx:alpine` image if it's not already present on your system and start the container in the background.

## 4. Discover the container

You can use the `budgie chirp` command to discover all running Budgie containers on your local network:

```bash
budgie chirp
```

This will output a list of all discovered containers, including their ID, name, image, and IP address.

## 5. Join as a replica

If you have another machine on the same network, you can join it to the `webapp` service as a replica. Simply run the following command on the second machine, replacing `<id>` with the ID of the `webapp` container you discovered in the previous step:

```bash
budgie chirp <id>
```

Budgie will automatically pull the `nginx:alpine` image and start a new container on the second machine. The two containers will be part of the same service and will synchronize their volumes.

## 6. Stop the container

To stop a container, use the `budgie stop` command with the container ID:

```bash
budgie stop <id>
```

This will gracefully stop the container.

That's it! You have now successfully run, discovered, and replicated a container using Budgie. For more detailed information, please refer to the other documentation guides.