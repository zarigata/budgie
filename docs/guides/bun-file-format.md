# .bun File Format

The `.bun` file is a YAML file that defines a Budgie container. It contains all the information needed to run, discover, and replicate a container.

## Top-level fields

| Field | Type | Description | Required |
|---|---|---|---|
| `version` | string | The version of the `.bun` file format. Currently, it should be `"1.0"`. | Yes |
| `name` | string | The name of the container. | Yes |
| `image` | object | The container image configuration. | Yes |
| `ports` | array | The port mappings for the container. | No |
| `volumes` | array | The volume mappings for the container. | No |
| `environment` | array | The environment variables for the container. | No |
| `healthcheck` | object | The health check configuration for the container. | No |
| `replicas` | object | The replication policy for the container. | No |

## `image`

The `image` object defines the container image to be used.

| Field | Type | Description | Required |
|---|---|---|---|
| `docker_image` | string | The name of the Docker image to use. | Yes |
| `command` | array | The command to run in the container. | No |
| `workdir` | string | The working directory inside the container. | No |

**Example:**
```yaml
image:
  docker_image: "nginx:alpine"
  command: ["/bin/sh", "-c", "echo 'Hello from budgie!' && nginx -g 'daemon off;'"]
  workdir: "/app"
```

## `ports`

The `ports` array defines the port mappings for the container. Each item in the array is an object with the following fields:

| Field | Type | Description | Required |
|---|---|---|---|
| `container_port` | integer | The port inside the container. | Yes |
| `host_port` | integer | The port on the host machine. | Yes |
| `protocol` | string | The protocol to use. Can be `tcp` or `udp`. Defaults to `tcp`. | No |

**Example:**
```yaml
ports:
  - container_port: 80
    host_port: 8080
    protocol: tcp
```

## `volumes`

The `volumes` array defines the volume mappings for the container. Each item in the array is an object with the following fields:

| Field | Type | Description | Required |
|---|---|---|---|
| `source` | string | The path to the directory on the host machine. | Yes |
| `target` | string | The path to the directory inside the container. | Yes |
| `mode` | string | The access mode for the volume. Can be `rw` (read-write) or `ro` (read-only). Defaults to `rw`. | No |

**Example:**
```yaml
volumes:
  - source: "./data"
    target: "/usr/share/nginx/html"
    mode: rw
```

## `environment`

The `environment` array defines the environment variables for the container. Each item in the array is a string in the format `KEY=VALUE`.

**Example:**
```yaml
environment:
  - APP_ENV=production
  - DEBUG=false
```

## `healthcheck`

The `healthcheck` object defines the health check for the container.

| Field | Type | Description | Required |
|---|---|---|---|
| `path` | string | The path to request for the health check. | Yes |
| `interval` | string | The interval between health checks (e.g., `30s`, `1m`). | No |
| `timeout` | string | The timeout for each health check (e.g., `5s`). | No |
| `retries` | integer | The number of retries before marking the container as unhealthy. | No |

**Example:**
```yaml
healthcheck:
  path: "/"
  interval: 30s
  timeout: 5s
  retries: 3
```

## `replicas`

The `replicas` object defines the replication policy for the container.

| Field | Type | Description | Required |
|---|---|---|---|
| `min` | integer | The minimum number of replicas to maintain. | No |
| `max` | integer | The maximum number of replicas to maintain. | No |

**Example:**
```yaml
replicas:
  min: 2
  max: 5
```