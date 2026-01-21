package proxy

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/budgie/budgie/pkg/types"
	"github.com/sirupsen/logrus"
)

type LoadBalancer interface {
	AddBackend(containerID, ip string, port int) error
	RemoveBackend(containerID, ip string, port int) error
	GetProxy(containerID string) (http.Handler, error)
	StartHealthCheck(interval time.Duration)
	Shutdown()
}

type LoadBalancerType string

const (
	RoundRobin LoadBalancerType = "round-robin"
	LeastConn  LoadBalancerType = "least-connections"
)

type ContainerProxy struct {
	mu       sync.RWMutex
	pools    map[string]*backendPool
	type     LoadBalancerType
	health   *HealthChecker
	shutdown chan struct{}
}

type backendPool struct {
	backends []*backend
	current  atomic.Uint64
	mu       sync.RWMutex
}

type backend struct {
	URL    *url.URL
	Active atomic.Bool
	Conn   atomic.Int64
}

type HealthChecker struct {
	interval time.Duration
	pools   map[string]*backendPool
	stop     chan struct{}
}

func NewContainerProxy(lbType LoadBalancerType) *ContainerProxy {
	return &ContainerProxy{
		pools:    make(map[string]*backendPool),
		type:     lbType,
		shutdown: make(chan struct{}),
	}
}

func (p *ContainerProxy) AddBackend(containerID, ip string, port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool, exists := p.pools[containerID]
	if !exists {
		pool = &backendPool{
			backends: make([]*backend, 0),
		}
		p.pools[containerID] = pool
	}

	url, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
	if err != nil {
		return fmt.Errorf("failed to parse backend URL: %w", err)
	}

	backend := &backend{
		URL: url,
	}
	backend.Active.Store(true)

	pool.mu.Lock()
	pool.backends = append(pool.backends, backend)
	pool.mu.Unlock()

	logrus.Infof("Added backend %s:%d for container %s", ip, port, containerID[:12])

	return nil
}

func (p *ContainerProxy) RemoveBackend(containerID, ip string, port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool, exists := p.pools[containerID]
	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	targetURL := fmt.Sprintf("http://%s:%d", ip, port)

	pool.mu.Lock()
	defer pool.mu.Unlock()

	for i, backend := range pool.backends {
		if backend.URL.String() == targetURL {
			pool.backends = append(pool.backends[:i], pool.backends[i+1:]...)
			logrus.Infof("Removed backend %s:%d for container %s", ip, port, containerID[:12])
			return nil
		}
	}

	return fmt.Errorf("backend not found")
}

func (p *ContainerProxy) GetProxy(containerID string) (http.Handler, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pool, exists := p.pools[containerID]
	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend := p.selectBackend(pool)
			if backend == nil {
				// Cannot write error here; ModifyResponse will handle it
				return
			}

			req.URL.Scheme = backend.URL.Scheme
			req.URL.Host = backend.URL.Host
			req.URL.Path = backend.URL.Path + req.URL.Path

			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
			req.Header.Set("X-Forwarded-Host", req.Host)

			// Store backend in context for connection tracking
			ctx := context.WithValue(req.Context(), "backend", backend)
			*req = *req.WithContext(ctx)
		},
	}

	// Wrap to handle no-backend case and connection tracking
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend := p.selectBackend(pool)
		if backend == nil {
			http.Error(w, "No backends available", http.StatusServiceUnavailable)
			return
		}

		backend.Conn.Add(1)
		defer backend.Conn.Add(-1)

		proxy.ServeHTTP(w, r)
	}), nil
}

func (p *ContainerProxy) selectBackend(pool *backendPool) *backend {
	switch p.type {
	case RoundRobin:
		return p.roundRobinSelect(pool)
	case LeastConn:
		return p.leastConnSelect(pool)
	default:
		return p.roundRobinSelect(pool)
	}
}

func (p *ContainerProxy) roundRobinSelect(pool *backendPool) *backend {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	if len(pool.backends) == 0 {
		return nil
	}

	activeBackends := make([]*backend, 0)
	for _, backend := range pool.backends {
		if backend.Active.Load() {
			activeBackends = append(activeBackends, backend)
		}
	}

	if len(activeBackends) == 0 {
		return nil
	}

	current := pool.current.Add(1) - 1
	idx := current % uint64(len(activeBackends))
	return activeBackends[idx]
}

func (p *ContainerProxy) leastConnSelect(pool *backendPool) *backend {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	if len(pool.backends) == 0 {
		return nil
	}

	var selected *backend
	minConn := int64(math.MaxInt64)

	for _, backend := range pool.backends {
		if !backend.Active.Load() {
			continue
		}

		conn := backend.Conn.Load()
		if conn < minConn {
			minConn = conn
			selected = backend
		}
	}

	return selected
}

func (p *ContainerProxy) StartHealthCheck(interval time.Duration) {
	p.health = &HealthChecker{
		interval: interval,
		pools:   p.pools,
		stop:     make(chan struct{}),
	}

	go p.health.Run()
}

func (h *HealthChecker) Run() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stop:
			return

		case <-ticker.C:
			h.checkAll()
		}
	}
}

func (h *HealthChecker) checkAll() {
	for containerID, pool := range h.pools {
		pool.mu.RLock()
		for _, backend := range pool.backends {
			go h.checkBackend(containerID, backend)
		}
		pool.mu.RUnlock()
	}
}

func (h *HealthChecker) checkBackend(containerID string, backend *backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", backend.URL.String()+"/_health", nil)
	if err != nil {
		backend.Active.Store(false)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if backend.Active.Load() {
			logrus.Warnf("Backend %s for container %s is unhealthy", backend.URL, containerID[:12])
		}
		backend.Active.Store(false)
		return
	}

	defer resp.Body.Close()

	if !backend.Active.Load() {
		logrus.Infof("Backend %s for container %s is back online", backend.URL, containerID[:12])
	}
	backend.Active.Store(true)
}

func (h *HealthChecker) Shutdown() {
	close(h.stop)
}

func (p *ContainerProxy) Shutdown() {
	close(p.shutdown)
	if p.health != nil {
		p.health.Shutdown()
	}
}
