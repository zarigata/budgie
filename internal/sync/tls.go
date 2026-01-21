package sync

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// TLSConfig holds TLS configuration for sync protocol
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
	CAFile   string
	// InsecureSkipVerify disables certificate verification (for testing only)
	InsecureSkipVerify bool
}

// NewTLSConfig creates a TLS configuration from files
func NewTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load certificate and key
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate for client verification
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	tlsConfig.InsecureSkipVerify = cfg.InsecureSkipVerify

	return tlsConfig, nil
}

// GenerateSelfSignedCert generates a self-signed certificate for development/testing
func GenerateSelfSignedCert(certDir string) (certFile, keyFile string, err error) {
	// Ensure directory exists
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	certFile = filepath.Join(certDir, "budgie.crt")
	keyFile = filepath.Join(certDir, "budgie.key")

	// Check if already exists
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			logrus.Debug("Using existing self-signed certificate")
			return certFile, keyFile, nil
		}
	}

	logrus.Info("Generating self-signed certificate...")

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Budgie"},
			CommonName:   hostname,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname, "localhost"},
		IPAddresses:           getLocalIPAddresses(),
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cert file: %w", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certOut.Close()
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}
	certOut.Close()

	// Write private key to file
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to create key file: %w", err)
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		keyOut.Close()
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		keyOut.Close()
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}
	keyOut.Close()

	logrus.Infof("Generated self-signed certificate: %s", certFile)
	return certFile, keyFile, nil
}

func getLocalIPAddresses() []net.IP {
	var ips []net.IP

	// Add localhost
	ips = append(ips, net.ParseIP("127.0.0.1"))
	ips = append(ips, net.ParseIP("::1"))

	// Get all network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

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
				ips = append(ips, ipnet.IP)
			}
		}
	}

	return ips
}

// TLSServer wraps a sync server with TLS
type TLSServer struct {
	*Server
	tlsConfig *tls.Config
}

// NewTLSServer creates a new TLS-enabled sync server
func NewTLSServer(port int, tlsCfg TLSConfig) (*TLSServer, error) {
	tlsConfig, err := NewTLSConfig(tlsCfg)
	if err != nil {
		return nil, err
	}

	var listener net.Listener
	addr := fmt.Sprintf(":%d", port)

	if tlsConfig != nil {
		listener, err = tls.Listen("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS listener: %w", err)
		}
		logrus.Info("Sync server using TLS encryption")
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener: %w", err)
		}
	}

	server := &Server{
		listener:    listener,
		volumePaths: make(map[string]string),
		done:        make(chan struct{}),
	}

	return &TLSServer{
		Server:    server,
		tlsConfig: tlsConfig,
	}, nil
}

// TLSClient provides TLS-enabled sync client capabilities
type TLSClient struct {
	tlsConfig *tls.Config
}

// NewTLSClient creates a new TLS-enabled sync client
func NewTLSClient(tlsCfg TLSConfig) (*TLSClient, error) {
	tlsConfig, err := NewTLSConfig(tlsCfg)
	if err != nil {
		return nil, err
	}

	return &TLSClient{
		tlsConfig: tlsConfig,
	}, nil
}

// Dial connects to a TLS-enabled sync server
func (c *TLSClient) Dial(address string, timeout time.Duration) (net.Conn, error) {
	if c.tlsConfig != nil {
		dialer := &net.Dialer{Timeout: timeout}
		return tls.DialWithDialer(dialer, "tcp", address, c.tlsConfig)
	}

	return net.DialTimeout("tcp", address, timeout)
}

// DialAndSync connects and performs a full volume sync
func (c *TLSClient) DialAndSync(address, localPath string, timeout time.Duration) error {
	conn, err := c.Dial(address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	mgr, err := NewSyncManager(localPath)
	if err != nil {
		return fmt.Errorf("failed to create sync manager: %w", err)
	}

	return mgr.ReceiveVolume(conn)
}
