# Nest Interactive UI

The `nest` command provides an interactive terminal user interface (TUI) for Budgie.

## Launching Nest

```bash
budgie nest
```

This opens a full-screen interactive interface.

## Features

### Main Menu

The main menu offers:

1. **Quick Start** - Step-by-step setup guide
2. **Custom Build** - Cross-compile for different platforms
3. **Learn Budgie** - Interactive tutorials
4. **Monitor** - View running containers
5. **System Check** - Verify dependencies
6. **Exit** - Close the wizard

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` / `Space` | Select |
| `b` / `Esc` | Go back |
| `q` | Quit |

## Quick Start

The Quick Start section walks you through:

1. Installing dependencies
2. Building budgie
3. Running your first container
4. Discovering network containers

## Custom Build

Build for specific platforms:

- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64, arm64)
- All platforms at once

Select a platform to see the exact build command.

## Learn Budgie

Interactive tutorials covering:

1. **Running Your First Container**
   - Creating bun files
   - Running containers
   - Using detach mode

2. **Discovering Containers**
   - Using chirp command
   - Understanding discovery output
   - mDNS basics

3. **Container Replication**
   - Joining as a peer
   - Volume synchronization
   - High availability

4. **Managing Containers**
   - Listing containers
   - Stopping containers
   - Cleanup

## Monitor Dashboard

View real-time container information:

- Container ID and name
- Status (running/stopped)
- CPU and memory usage
- Network I/O
- Process count

## System Check

Verifies your system has:

- Go 1.21 or later
- containerd installed and running
- mDNS support

## Tips

- Use `nest` for initial setup on new machines
- The tutorials are great for learning
- System Check helps diagnose issues
- Monitor is useful for debugging

## Aliases

The nest command has several aliases:

```bash
budgie nest
budgie setup
budgie wizard
budgie init
```

All launch the same interactive UI.
