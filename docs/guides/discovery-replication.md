# Discovery and Replication

Learn how Budgie discovers containers on your network and enables replication.

## How Discovery Works

Budgie uses mDNS (multicast DNS) to announce and discover containers on your local network. When a container starts, it broadcasts its presence using the `_budgie._tcp` service type.

### Automatic Announcement

When you run a container with `budgie run`, it's automatically announced:

```bash
budgie run myapp.bun
# Container announces itself on the network
```

### Discovering Containers

Use the `chirp` command to find containers:

```bash
budgie chirp
```

Output:
```
CONTAINER ID   NAME      IP            PORT   IMAGE           NODE
abc123456789   webapp    192.168.1.5   8080   nginx:alpine    laptop
def987654321   api       192.168.1.6   3000   node:18         desktop
ghi456789012   db        192.168.1.7   5432   postgres:15     server
```

## What's Announced

Each container broadcasts:
- Container ID
- Container name
- Node ID (hostname)
- Docker image
- Exposed ports

## Replication

Budgie supports container replication for high availability.

### Joining as a Peer

To replicate a container on your machine:

```bash
budgie chirp abc123
```

This will:
1. Connect to the primary node
2. Download the container image
3. Synchronize volume data
4. Start a local replica

### How Sync Works

Volume synchronization uses an efficient delta-sync algorithm:

1. **Signature Exchange**: Nodes exchange file signatures
2. **Delta Calculation**: Only changed blocks are identified
3. **Transfer**: Only changed data is sent
4. **Apply**: Changes are applied to the replica

### Sync Protocol

Budgie uses TCP port 18733 for volume synchronization:

```
Primary Node                 Replica Node
     |                            |
     |  <-- Connect               |
     |                            |
     |  Signatures -->            |
     |                            |
     |  <-- Needed Files          |
     |                            |
     |  File Data -->             |
     |                            |
     |  <-- ACK                   |
```

## Replica Configuration

Configure replica limits in your bun file:

```yaml
replicas:
  min: 1    # Minimum replicas to maintain
  max: 5    # Maximum replicas allowed
```

## Network Requirements

For discovery to work:

1. **Same Network Segment**: Nodes must be on the same LAN
2. **mDNS Traffic Allowed**: UDP port 5353 must not be blocked
3. **Sync Port Open**: TCP port 18733 for volume sync

### Firewall Rules

Allow Budgie traffic:

```bash
# mDNS for discovery
sudo ufw allow 5353/udp

# Volume sync
sudo ufw allow 18733/tcp
```

## Use Cases

### High Availability

Run replicas across multiple machines:

```bash
# On machine 1
budgie run webapp.bun

# On machine 2
budgie chirp abc123

# Now webapp runs on both machines
```

### Local Development

Sync your development environment:

```bash
# On your laptop
budgie chirp xyz789
# Now you have a local copy of the production config
```

### Disaster Recovery

Keep replicas on separate hardware:

```yaml
replicas:
  min: 2    # Ensure at least 2 copies exist
  max: 3
```

## Monitoring Replicas

Check replica status:

```bash
budgie ps --all
```

Shows which nodes have replicas of your containers.

## Troubleshooting

### Container Not Discovered

- Check firewall allows UDP 5353
- Verify nodes are on same network
- Check mDNS is not blocked by router

### Sync Fails

- Verify TCP 18733 is open
- Check source node is running
- Verify volume paths exist

### Slow Sync

- Large volumes take time on first sync
- Subsequent syncs only transfer changes
- Check network bandwidth between nodes
