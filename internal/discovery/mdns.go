package discovery

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/budgie/budgie/pkg/types"
	"github.com/hashicorp/mdns"
	"github.com/sirupsen/logrus"
)

type DiscoveryService struct {
	servers []*mdns.Server
	mu      sync.RWMutex
}

func NewDiscoveryService() *DiscoveryService {
	return &DiscoveryService{
		servers: make([]*mdns.Server, 0),
	}
}

func (d *DiscoveryService) AnnounceContainer(ctr *types.Container) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, port := range ctr.Ports {
		txt := []string{
			fmt.Sprintf("container_id=%s", ctr.ID),
			fmt.Sprintf("node_id=%s", ctr.NodeID),
			fmt.Sprintf("container_name=%s", ctr.Name),
			fmt.Sprintf("image=%s", ctr.Image.DockerImage),
		}

		serviceName := fmt.Sprintf("budgie-%s", ctr.ShortID())
		service, err := mdns.NewMDNSService(
			serviceName,
			"_budgie._tcp",
			"local.",
			ctr.ID,
			port.HostPort,
			txt,
		)
		if err != nil {
			return fmt.Errorf("failed to create mDNS service: %w", err)
		}

		server, err := mdns.NewServer(&mdns.Config{Zone: service})
		if err != nil {
			return fmt.Errorf("failed to create mDNS server: %w", err)
		}

		d.servers = append(d.servers, server)

		ips := getLocalIPs()
		logrus.Infof("Announcing container %s on %s:%d", ctr.ShortID(), ips[0], port.HostPort)
	}

	return nil
}

func (d *DiscoveryService) DiscoverContainers(timeout time.Duration) ([]DiscoveredContainer, error) {
	entries := make(chan *mdns.ServiceEntry)
	var containers []DiscoveredContainer
	var mu sync.Mutex

	go func() {
		for entry := range entries {
			if entry.InfoFields == nil {
				continue
			}

			ctr := parseEntry(entry)
			if ctr != nil {
				mu.Lock()
				containers = append(containers, *ctr)
				mu.Unlock()
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := &mdns.QueryParam{
		Service:             "_budgie._tcp",
		Domain:             "local",
		Timeout:            timeout,
		Entries:            entries,
		WantUnicastResponse: false,
	}

	if err := mdns.Query(params); err != nil {
		return nil, fmt.Errorf("mDNS query failed: %w", err)
	}

	<-ctx.Done()

	return containers, nil
}

func (d *DiscoveryService) Shutdown() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var lastErr error
	for _, server := range d.servers {
		if server != nil {
			if err := server.Shutdown(); err != nil {
				logrus.Errorf("Failed to shutdown mDNS server: %v", err)
				lastErr = err
			}
		}
	}

	return lastErr
}

type DiscoveredContainer struct {
	ID      string
	Name    string
	NodeID  string
	Image   string
	IPs     []string
	Port    int
	NameTag string
}

func parseEntry(entry *mdns.ServiceEntry) *DiscoveredContainer {
	if entry.Typev4 == nil {
		return nil
	}

	infoMap := make(map[string]string)
	for _, field := range entry.InfoFields {
		key, val, ok := parseTxtField(field)
		if ok {
			infoMap[key] = val
		}
	}

	containerID, ok := infoMap["container_id"]
	if !ok {
		return nil
	}

	ctr := &DiscoveredContainer{
		ID:      containerID,
		Name:    infoMap["container_name"],
		NodeID:  infoMap["node_id"],
		Image:   infoMap["image"],
		Port:    entry.Port,
		NameTag: entry.Name,
	}

	if len(entry.AddrV4) > 0 {
		ctr.IPs = make([]string, 0, len(entry.AddrV4))
		for _, addr := range entry.AddrV4 {
			ctr.IPs = append(ctr.IPs, addr.String())
		}
	}

	return ctr
}

func parseTxtField(field string) (string, string, bool) {
	for i, c := range field {
		if c == '=' {
			return field[:i], field[i+1:], true
		}
	}
	return "", "", false
}

func getLocalIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Errorf("Failed to get interfaces: %v", err)
		return []string{"127.0.0.1"}
	}

	var ips []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}

	if len(ips) == 0 {
		return []string{"127.0.0.1"}
	}

	return ips
}
