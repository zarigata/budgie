# CLI Commands

This page provides a reference for all the available Budgie CLI commands.

## `budgie run`

Starts a container from a `.bun` file.

**Usage:**
```bash
budgie run <file.bun>
```

**Arguments:**
- `<file.bun>`: The path to the `.bun` file.

## `budgie ps`

Lists all running containers.

**Usage:**
```bash
budgie ps
```

**Flags:**
- `--all`, `-a`: Show all containers, including stopped ones.

## `budgie stop`

Gracefully stops a container.

**Usage:**
```bash
budgie stop <id>
```

**Arguments:**
- `<id>`: The ID of the container to stop.

## `budgie chirp`

Discovers containers on the local network or joins a container as a replica.

**Usage:**
```bash
# Discover containers
budgie chirp

# Join as a replica
budgie chirp <id>
```

**Arguments:**
- `<id>` (optional): The ID of the container to join as a replica.

## `budgie nest`

Starts an interactive setup wizard.

**Usage:**
```bash
budgie nest
```

This command will guide you through the process of creating a `.bun` file and running a container.