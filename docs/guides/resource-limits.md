# Resource Limits

Control CPU, memory, and other resources for your containers.

## Overview

Budgie uses Linux cgroups (via containerd) to enforce resource limits. This ensures containers don't consume excessive system resources.

## CPU Limits

### CPU Shares

Relative CPU weight (default: 1024). Higher values get more CPU time relative to other containers.

```yaml
resources:
  cpu_shares: 512      # Half the default weight
  cpu_shares: 2048     # Double the default weight
```

### CPU Quota

Hard limit on CPU time in microseconds per 100ms period.

```yaml
resources:
  cpu_quota: 50000     # 50% of one CPU core
  cpu_quota: 100000    # 100% of one CPU core
  cpu_quota: 200000    # 200% (2 cores)
```

## Memory Limits

### Memory Limit

Maximum memory the container can use (in bytes).

```yaml
resources:
  memory_limit: 134217728   # 128 MB
  memory_limit: 268435456   # 256 MB
  memory_limit: 536870912   # 512 MB
  memory_limit: 1073741824  # 1 GB
```

Helpful values:
- 64 MB: `67108864`
- 128 MB: `134217728`
- 256 MB: `268435456`
- 512 MB: `536870912`
- 1 GB: `1073741824`
- 2 GB: `2147483648`

### Memory + Swap Limit

Total memory + swap the container can use.

```yaml
resources:
  memory_limit: 268435456   # 256 MB RAM
  memory_swap: 536870912    # 512 MB total (RAM + swap)
```

## Process Limits

### PIDs Limit

Maximum number of processes/threads the container can create.

```yaml
resources:
  pids_limit: 100          # Max 100 processes
  pids_limit: 500          # Max 500 processes
```

This prevents fork bombs and runaway process creation.

## Block I/O Limits

### Block I/O Weight

Relative weight for disk I/O (10-1000, default: 500).

```yaml
resources:
  blkio_weight: 100        # Low I/O priority
  blkio_weight: 500        # Normal priority
  blkio_weight: 1000       # High I/O priority
```

## Common Configurations

### Minimal Container

For simple, lightweight services:

```yaml
resources:
  cpu_shares: 256
  memory_limit: 67108864   # 64 MB
  pids_limit: 50
```

### Standard Web Application

For typical web services:

```yaml
resources:
  cpu_shares: 512
  memory_limit: 268435456  # 256 MB
  pids_limit: 100
```

### Database

For memory-intensive applications:

```yaml
resources:
  cpu_shares: 1024
  memory_limit: 1073741824  # 1 GB
  memory_swap: 2147483648   # 2 GB total
  pids_limit: 200
  blkio_weight: 800
```

### Worker Process

For CPU-intensive tasks:

```yaml
resources:
  cpu_shares: 2048
  cpu_quota: 200000         # 2 CPU cores
  memory_limit: 536870912   # 512 MB
  pids_limit: 50
```

## Best Practices

1. **Always set memory limits** - Prevents containers from consuming all system memory

2. **Set PID limits** - Protects against fork bombs

3. **Use CPU shares for relative prioritization** - Let the kernel schedule fairly

4. **Use CPU quota for hard limits** - When you need guaranteed isolation

5. **Monitor actual usage** - Adjust limits based on real workload patterns

## Troubleshooting

### Container Killed (OOM)

If your container is being killed:
- Check logs for "OOM killed" messages
- Increase `memory_limit`
- Investigate memory leaks in your application

### Slow Performance

If container is slow:
- Check CPU quota isn't too restrictive
- Increase `cpu_shares`
- Check `blkio_weight` for I/O-bound workloads

### Process Creation Failed

If container can't create processes:
- Increase `pids_limit`
- Check for process leaks (zombies)
