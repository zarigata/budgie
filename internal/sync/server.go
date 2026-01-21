package sync

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultSyncPort is the default port for sync server
	DefaultSyncPort = 18733
)

// Server handles incoming sync requests
type Server struct {
	listener   net.Listener
	volumePaths map[string]string // containerID -> volume path
	mu         sync.RWMutex
	done       chan struct{}
}

// NewServer creates a new sync server
func NewServer(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	return &Server{
		listener:    listener,
		volumePaths: make(map[string]string),
		done:        make(chan struct{}),
	}, nil
}

// RegisterVolume registers a container's volume for sync
func (s *Server) RegisterVolume(containerID, volumePath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.volumePaths[containerID] = volumePath
	logrus.Infof("Registered volume for container %s at %s", containerID[:12], volumePath)
}

// UnregisterVolume removes a container's volume from sync
func (s *Server) UnregisterVolume(containerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.volumePaths, containerID)
}

// Start starts accepting connections
func (s *Server) Start() {
	logrus.Infof("Sync server listening on %s", s.listener.Addr())

	for {
		select {
		case <-s.done:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				logrus.Errorf("Failed to accept connection: %v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single sync connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	logrus.Infof("New sync connection from %s", remoteAddr)

	// Get the default volume path (first registered)
	s.mu.RLock()
	var volumePath string
	for _, path := range s.volumePaths {
		volumePath = path
		break
	}
	s.mu.RUnlock()

	if volumePath == "" {
		logrus.Warn("No volumes registered for sync")
		return
	}

	// Create sync manager and receive data
	mgr, err := NewSyncManager(volumePath)
	if err != nil {
		logrus.Errorf("Failed to create sync manager: %v", err)
		return
	}

	if err := mgr.ReceiveVolume(conn); err != nil {
		logrus.Errorf("Failed to receive volume data: %v", err)
		return
	}

	logrus.Infof("Sync completed from %s", remoteAddr)
}

// handleConnectionWithProtocol handles connection using the sync protocol
func (s *Server) handleConnectionWithProtocol(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	logrus.Infof("New sync connection from %s", remoteAddr)

	proto := NewProtocol(conn)

	// Receive initial message
	msg, err := proto.Receive()
	if err != nil {
		logrus.Errorf("Failed to receive message: %v", err)
		return
	}

	switch msg.Type {
	case MsgSignatureRequest:
		s.handleSignatureRequest(proto, msg.Payload.(SignatureRequest))
	case MsgDeltaRequest:
		s.handleDeltaRequest(proto, msg.Payload.(DeltaRequest))
	default:
		proto.SendError(400, "unexpected message type")
	}
}

// handleSignatureRequest handles a signature request
func (s *Server) handleSignatureRequest(proto *Protocol, req SignatureRequest) {
	s.mu.RLock()
	volumePath, ok := s.volumePaths[req.ContainerID]
	s.mu.RUnlock()

	if !ok {
		proto.SendError(404, "container not found")
		return
	}

	fullPath := filepath.Join(volumePath, req.VolumePath)
	mgr, err := NewSyncManager(fullPath)
	if err != nil {
		proto.SendError(500, err.Error())
		return
	}

	signatures, err := mgr.collectSignatures()
	if err != nil {
		proto.SendError(500, err.Error())
		return
	}

	proto.SendSignatures(signatures)
}

// handleDeltaRequest handles a delta/file request
func (s *Server) handleDeltaRequest(proto *Protocol, req DeltaRequest) {
	// Get first registered volume
	s.mu.RLock()
	var volumePath string
	for _, path := range s.volumePaths {
		volumePath = path
		break
	}
	s.mu.RUnlock()

	if volumePath == "" {
		proto.SendError(404, "no volumes registered")
		return
	}

	for _, relPath := range req.Files {
		fullPath := filepath.Join(volumePath, relPath)
		data, err := readFileData(fullPath)
		if err != nil {
			logrus.Errorf("Failed to read file %s: %v", relPath, err)
			continue
		}

		proto.SendFile(FileSignature{Path: relPath, Size: int64(len(data))}, data)
	}

	proto.SendAck(true, "transfer complete")
}

// readFileData reads entire file into memory
func readFileData(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Stop stops the sync server
func (s *Server) Stop() error {
	close(s.done)
	return s.listener.Close()
}

// Addr returns the server's address
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// StartDefaultServer starts a sync server on the default port
func StartDefaultServer() (*Server, error) {
	server, err := NewServer(DefaultSyncPort)
	if err != nil {
		return nil, err
	}

	go server.Start()
	return server, nil
}
