# Budgie Improvement Roadmap

This document tracks all planned improvements and their status.

## Legend
- [ ] Not started
- [x] Completed
- [~] In progress
- [-] Skipped/Deferred

---

## Critical - Code Quality & Bug Fixes

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 1 | Extract duplicate `findContainer()` function - Same 106-line function exists in 4 files (logs, exec, inspect, rm). Create shared utility. | [ ] | Critical |
| 2 | Fix silent state save failures - `manager.go` logs errors but doesn't propagate them, risking data loss. | [ ] | Critical |
| 3 | Remove panic in production code - `pkg/types/container.go:107` panics on ID generation error. | [ ] | Critical |
| 4 | Add input validation - Port ranges (0-65535), container names, volume paths (prevent directory traversal). | [ ] | Critical |

---

## Complete Incomplete Features

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 5 | Implement `--name` flag in `run` command - Flag exists but is ignored. | [ ] | High |
| 6 | Implement `--timestamps` and `--since` for `logs` - Flags declared but not functional. | [ ] | High |
| 7 | Implement `--user`, `--workdir`, `--env` for `exec` - Flags declared but not functional. | [ ] | High |
| 8 | Implement `--format` template for `inspect` - Go template support like Docker. | [ ] | High |
| 9 | Fix `nest` containerd check - Currently hardcoded to return `true`. | [ ] | High |

---

## New Features - Developer Experience

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 10 | Add `budgie version` command - Show version, Go version, containerd version, build info. | [ ] | Medium |
| 11 | Add `budgie pull <image>` command - Pre-pull images without running containers. | [ ] | Medium |
| 12 | Add `budgie images` command - List locally available images. | [ ] | Medium |
| 13 | Add `budgie stats <id>` command - Real-time CPU/memory/network stats. | [ ] | Medium |
| 14 | Add `budgie top <id>` command - Show processes running inside container. | [ ] | Medium |
| 15 | Add `budgie rename <id> <name>` - Rename containers. | [ ] | Medium |
| 16 | Add `budgie pause/unpause <id>` - Pause and resume containers. | [ ] | Medium |
| 17 | Add `budgie cp <src> <dst>` - Copy files between host and container. | [ ] | Medium |
| 18 | Add shell completion scripts - Bash, Zsh, Fish autocompletion. | [ ] | Medium |

---

## New Features - Operations

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 19 | Health check monitoring - Actually monitor health checks and restart unhealthy containers. | [ ] | Medium |
| 20 | Container dependency enforcement - Enforce `depends_on` startup ordering. | [ ] | Medium |
| 21 | Graceful shutdown handling - Proper SIGTERM/SIGINT handling for the daemon. | [ ] | Medium |
| 22 | Image garbage collection - Clean up unused images automatically. | [ ] | Low |
| 23 | Container events/streaming - `budgie events` to stream container lifecycle events. | [ ] | Low |
| 24 | Secrets management - Encrypted secrets injection into containers. | [ ] | Low |
| 25 | Custom networks - Create isolated container networks. | [ ] | Low |
| 26 | Privileged mode support - Allow privileged containers when needed. | [ ] | Low |

---

## Observability & Monitoring

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 27 | Prometheus metrics endpoint - Export container metrics for monitoring. | [ ] | Low |
| 28 | Structured logging with correlation IDs - Track operations across distributed nodes. | [ ] | Low |
| 29 | Web dashboard - Browser-based UI for container management. | [ ] | Low |
| 30 | Resource usage history - Store and display historical CPU/memory usage. | [ ] | Low |

---

## Testing & Documentation

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 31 | Unit test suite - Cover internal packages (bundle, config, api, discovery, sync). | [ ] | Medium |
| 32 | Integration test suite - Test full workflows. | [ ] | Medium |
| 33 | Godoc comments - Document all exported functions. | [ ] | Low |
| 34 | Configuration schema documentation - Full YAML reference. | [ ] | Low |
| 35 | Architecture diagram - Visual component relationships. | [ ] | Low |
| 36 | Troubleshooting guide - Common issues and solutions. | [ ] | Low |

---

## Security Hardening

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 37 | Certificate pinning for TLS - Prevent MITM attacks. | [ ] | Low |
| 38 | Config file permission checks - Warn if config is world-readable. | [ ] | Low |
| 39 | Data directory permissions - Change from 755 to 700. | [ ] | Low |
| 40 | Container name sanitization - Prevent injection attacks. | [ ] | Low |

---

## Build & CI Improvements

| # | Feature | Status | Priority |
|---|---------|--------|----------|
| 41 | Pre-commit hooks - Auto-format, lint, test before commits. | [ ] | Low |
| 42 | Test coverage reporting - Track and enforce minimum coverage. | [ ] | Low |
| 43 | Makefile improvements - Add lint, fmt-check, coverage targets. | [ ] | Low |
| 44 | Release automation - Automated versioned releases. | [ ] | Low |

---

## Implementation Notes

### File Locations for Key Changes

**Critical Fixes:**
- `findContainer()` duplication: `cmd/logs/logs.go`, `cmd/exec/exec.go`, `cmd/inspect/inspect.go`, `cmd/rm/rm.go`
- State save failures: `internal/api/manager.go:59, 86, 113, 139`
- Panic removal: `pkg/types/container.go:107`
- Input validation: `internal/bundle/bundle.go`

**Incomplete Features:**
- `--name` flag: `cmd/run/run.go:18, 68, 105`
- `--timestamps/--since`: `cmd/logs/logs.go:18-22, 115-116`
- `--user/--workdir/--env`: `cmd/exec/exec.go:19-22, 114-117`
- `--format` template: `cmd/inspect/inspect.go:21, 212`
- Containerd check: `internal/ui/nest.go:438`

---

## Progress Tracking

**Last Updated:** (will be updated as work progresses)

**Session Progress:**
- Started: Items 1-9 (Critical + Incomplete)
- Completed: (to be filled)
- Remaining: (to be filled)

---

## Quick Start for Future Sessions

To continue work on this roadmap:

```bash
# Check current status
cat ROADMAP.md | grep -E "^\| [0-9]+"

# Run tests after changes
make test

# Build and verify
make build
./budgie --help
```
