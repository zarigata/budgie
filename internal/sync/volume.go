package sync

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/zarigata/budgie/pkg/types"
)

// FileSignature represents a file's sync signature
type FileSignature struct {
	Path     string
	Size     int64
	ModTime  int64
	Checksum []byte
}

// SyncManager handles volume synchronization between nodes
type SyncManager struct {
	localPath string
}

// NewSyncManager creates a new sync manager for the given path
func NewSyncManager(localPath string) (*SyncManager, error) {
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", localPath)
	}
	return &SyncManager{
		localPath: localPath,
	}, nil
}

// SendVolume sends volume signatures and data over the connection
func (s *SyncManager) SendVolume(conn net.Conn) error {
	encoder := gob.NewEncoder(conn)

	// First, send all file signatures
	signatures, err := s.collectSignatures()
	if err != nil {
		return fmt.Errorf("failed to collect signatures: %w", err)
	}

	// Send signature count
	if err := encoder.Encode(len(signatures)); err != nil {
		return fmt.Errorf("failed to send signature count: %w", err)
	}

	// Send each signature
	for _, sig := range signatures {
		if err := encoder.Encode(sig); err != nil {
			return fmt.Errorf("failed to send signature for %s: %w", sig.Path, err)
		}
		logrus.Debugf("Sent signature for %s", sig.Path)
	}

	// Receive list of files needed
	decoder := gob.NewDecoder(conn)
	var neededFiles []string
	if err := decoder.Decode(&neededFiles); err != nil {
		return fmt.Errorf("failed to receive needed files list: %w", err)
	}

	// Send each needed file
	for _, relPath := range neededFiles {
		fullPath := filepath.Join(s.localPath, relPath)
		if err := s.sendFile(conn, fullPath, relPath); err != nil {
			return fmt.Errorf("failed to send file %s: %w", relPath, err)
		}
		logrus.Debugf("Sent file %s", relPath)
	}

	return nil
}

// collectSignatures walks the directory and collects file signatures
func (s *SyncManager) collectSignatures() ([]FileSignature, error) {
	var signatures []FileSignature

	err := filepath.Walk(s.localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.localPath, path)
		if err != nil {
			return err
		}

		// Simple checksum using first/last bytes and size
		checksum, err := quickChecksum(path, info.Size())
		if err != nil {
			return err
		}

		sig := FileSignature{
			Path:     relPath,
			Size:     info.Size(),
			ModTime:  info.ModTime().UnixNano(),
			Checksum: checksum,
		}
		signatures = append(signatures, sig)

		return nil
	})

	return signatures, err
}

// quickChecksum creates a fast checksum for file comparison
func quickChecksum(path string, size int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read first 1KB
	first := make([]byte, 1024)
	n1, _ := f.Read(first)
	first = first[:n1]

	// Read last 1KB
	last := make([]byte, 1024)
	if size > 1024 {
		f.Seek(-1024, io.SeekEnd)
	}
	n2, _ := f.Read(last)
	last = last[:n2]

	// Combine for checksum
	checksum := make([]byte, 0, len(first)+len(last)+8)
	checksum = append(checksum, first...)
	checksum = append(checksum, last...)

	return checksum, nil
}

// sendFile sends a single file over the connection
func (s *SyncManager) sendFile(conn net.Conn, fullPath, relPath string) error {
	f, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(conn)

	// Send file metadata
	meta := FileSignature{
		Path:    relPath,
		Size:    info.Size(),
		ModTime: info.ModTime().UnixNano(),
	}
	if err := encoder.Encode(meta); err != nil {
		return err
	}

	// Send file content
	_, err = io.Copy(conn, f)
	return err
}

// ReceiveVolume receives volume data from a sync sender
func (s *SyncManager) ReceiveVolume(conn net.Conn) error {
	decoder := gob.NewDecoder(conn)

	// Receive signature count
	var count int
	if err := decoder.Decode(&count); err != nil {
		return fmt.Errorf("failed to receive signature count: %w", err)
	}

	// Receive signatures and determine which files we need
	var neededFiles []string
	for i := 0; i < count; i++ {
		var sig FileSignature
		if err := decoder.Decode(&sig); err != nil {
			return fmt.Errorf("failed to receive signature: %w", err)
		}

		// Check if we need this file
		localPath := filepath.Join(s.localPath, sig.Path)
		if needsUpdate(localPath, sig) {
			neededFiles = append(neededFiles, sig.Path)
		}
	}

	// Send list of needed files
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(neededFiles); err != nil {
		return fmt.Errorf("failed to send needed files list: %w", err)
	}

	// Receive each needed file
	for range neededFiles {
		if err := s.receiveFile(conn); err != nil {
			return fmt.Errorf("failed to receive file: %w", err)
		}
	}

	return nil
}

// needsUpdate checks if a local file needs to be updated
func needsUpdate(localPath string, remoteSig FileSignature) bool {
	info, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return true
	}

	// Compare size and mod time
	if info.Size() != remoteSig.Size {
		return true
	}
	if info.ModTime().UnixNano() < remoteSig.ModTime {
		return true
	}

	return false
}

// receiveFile receives a single file from the connection
func (s *SyncManager) receiveFile(conn net.Conn) error {
	decoder := gob.NewDecoder(conn)

	// Receive file metadata
	var meta FileSignature
	if err := decoder.Decode(&meta); err != nil {
		return err
	}

	// Create directory if needed
	fullPath := filepath.Join(s.localPath, meta.Path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Receive content
	_, err = io.CopyN(f, conn, meta.Size)
	if err != nil {
		return err
	}

	logrus.Debugf("Received file %s (%d bytes)", meta.Path, meta.Size)
	return nil
}

// SyncVolume synchronizes a local volume with a remote node
func (s *SyncManager) SyncVolume(srcPath, dstAddr string) error {
	conn, err := net.Dial("tcp", dstAddr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	logrus.Infof("Syncing volume %s to %s", srcPath, dstAddr)

	if err := s.SendVolume(conn); err != nil {
		return err
	}

	logrus.Infof("Volume sync completed")
	return nil
}

// SyncContainerData synchronizes all rw volumes for a container
func (s *SyncManager) SyncContainerData(ctr *types.Container, remoteIP string) error {
	for _, vol := range ctr.Volumes {
		if vol.Mode == "rw" {
			localVolPath := vol.Source
			if !filepath.IsAbs(localVolPath) {
				cwd, _ := os.Getwd()
				localVolPath = filepath.Join(cwd, localVolPath)
			}

			syncAddr := fmt.Sprintf("%s:18733", remoteIP)

			mgr, err := NewSyncManager(localVolPath)
			if err != nil {
				return fmt.Errorf("failed to create sync manager for %s: %w", vol.Target, err)
			}

			if err := mgr.SyncVolume(localVolPath, syncAddr); err != nil {
				return fmt.Errorf("failed to sync volume %s: %w", vol.Target, err)
			}
		}
	}

	return nil
}

// VolumeWatcher watches for volume changes
type VolumeWatcher struct {
	manager   *SyncManager
	watcher   *fsnotify.Watcher
	done      chan struct{}
	OnChange  func(path string)
}

// NewVolumeWatcher creates a new volume watcher
func NewVolumeWatcher(path string) (*VolumeWatcher, error) {
	manager, err := NewSyncManager(path)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Add all directories recursively
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	})
	if err != nil {
		watcher.Close()
		return nil, err
	}

	vw := &VolumeWatcher{
		manager: manager,
		watcher: watcher,
		done:    make(chan struct{}),
	}

	go vw.watch()

	return vw, nil
}

func (vw *VolumeWatcher) watch() {
	for {
		select {
		case <-vw.done:
			return

		case event, ok := <-vw.watcher.Events:
			if !ok {
				return
			}

			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove) != 0 {
				logrus.Infof("Volume changed: %s", event.Name)
				if vw.OnChange != nil {
					vw.OnChange(event.Name)
				}
			}

			// Add new directories to watch
			if event.Op&fsnotify.Create != 0 {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					vw.watcher.Add(event.Name)
				}
			}

		case err, ok := <-vw.watcher.Errors:
			if !ok {
				return
			}
			logrus.Errorf("Watcher error: %v", err)
		}
	}
}

// Close stops the volume watcher
func (vw *VolumeWatcher) Close() {
	close(vw.done)
	vw.watcher.Close()
}
