package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
)

// Secret represents an encrypted secret
type Secret struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Data      string            `json:"data"`       // Base64 encoded encrypted data
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// SecretManager manages encrypted secrets
type SecretManager struct {
	secrets   map[string]*Secret
	dataDir   string
	statePath string
	key       []byte
	mu        sync.RWMutex
}

const (
	saltSize    = 32
	keySize     = 32
	iterations  = 100000
	secretsFile = "secrets.json"
	keyFile     = ".secrets.key"
)

// NewSecretManager creates a new secret manager
func NewSecretManager(dataDir string) (*SecretManager, error) {
	sm := &SecretManager{
		secrets:   make(map[string]*Secret),
		dataDir:   dataDir,
		statePath: filepath.Join(dataDir, secretsFile),
	}

	// Ensure secrets directory exists with restricted permissions
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}

	// Load or generate encryption key
	if err := sm.loadOrGenerateKey(); err != nil {
		return nil, fmt.Errorf("failed to initialize encryption key: %w", err)
	}

	// Load existing secrets
	if err := sm.loadState(); err != nil {
		logrus.Warnf("Failed to load secrets state (starting fresh): %v", err)
	}

	return sm, nil
}

// CreateSecret creates a new encrypted secret
func (sm *SecretManager) CreateSecret(name string, data []byte) (*Secret, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.secrets[name]; exists {
		return nil, fmt.Errorf("secret already exists: %s", name)
	}

	// Encrypt the data
	encryptedData, err := sm.encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	secret := &Secret{
		ID:        generateSecretID(),
		Name:      name,
		Data:      base64.StdEncoding.EncodeToString(encryptedData),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    make(map[string]string),
	}

	sm.secrets[name] = secret

	if err := sm.saveState(); err != nil {
		delete(sm.secrets, name)
		return nil, fmt.Errorf("failed to save secret: %w", err)
	}

	logrus.Infof("Created secret %s", name)
	return secret, nil
}

// GetSecret retrieves and decrypts a secret
func (sm *SecretManager) GetSecret(name string) ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secret, exists := sm.secrets[name]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", name)
	}

	// Decode and decrypt
	encryptedData, err := base64.StdEncoding.DecodeString(secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	return sm.decrypt(encryptedData)
}

// UpdateSecret updates an existing secret
func (sm *SecretManager) UpdateSecret(name string, data []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	secret, exists := sm.secrets[name]
	if !exists {
		return fmt.Errorf("secret not found: %s", name)
	}

	// Encrypt the new data
	encryptedData, err := sm.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	secret.Data = base64.StdEncoding.EncodeToString(encryptedData)
	secret.UpdatedAt = time.Now()

	if err := sm.saveState(); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	logrus.Infof("Updated secret %s", name)
	return nil
}

// RemoveSecret removes a secret
func (sm *SecretManager) RemoveSecret(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.secrets[name]; !exists {
		return fmt.Errorf("secret not found: %s", name)
	}

	delete(sm.secrets, name)

	if err := sm.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	logrus.Infof("Removed secret %s", name)
	return nil
}

// ListSecrets returns all secret metadata (without decrypted data)
func (sm *SecretManager) ListSecrets() []*SecretInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	list := make([]*SecretInfo, 0, len(sm.secrets))
	for _, secret := range sm.secrets {
		list = append(list, &SecretInfo{
			ID:        secret.ID,
			Name:      secret.Name,
			CreatedAt: secret.CreatedAt,
			UpdatedAt: secret.UpdatedAt,
		})
	}
	return list
}

// SecretInfo contains non-sensitive secret metadata
type SecretInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetSecretEnv returns secret value formatted as environment variable
func (sm *SecretManager) GetSecretEnv(secretName, envName string) (string, error) {
	data, err := sm.GetSecret(secretName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s=%s", envName, string(data)), nil
}

// encrypt encrypts data using AES-GCM
func (sm *SecretManager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (sm *SecretManager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (sm *SecretManager) loadOrGenerateKey() error {
	keyPath := filepath.Join(sm.dataDir, keyFile)

	// Try to load existing key
	keyData, err := os.ReadFile(keyPath)
	if err == nil && len(keyData) >= saltSize+keySize {
		// Derive key from stored salt
		salt := keyData[:saltSize]
		sm.key = pbkdf2.Key(keyData[saltSize:], salt, iterations, keySize, sha256.New)
		return nil
	}

	// Generate new key
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	masterKey := make([]byte, keySize)
	if _, err := rand.Read(masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	sm.key = pbkdf2.Key(masterKey, salt, iterations, keySize, sha256.New)

	// Save salt + master key with restricted permissions
	keyData = append(salt, masterKey...)
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		return fmt.Errorf("failed to save key file: %w", err)
	}

	return nil
}

func generateSecretID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:12]
}

func (sm *SecretManager) loadState() error {
	data, err := os.ReadFile(sm.statePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var secrets []*Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return err
	}

	for _, secret := range secrets {
		sm.secrets[secret.Name] = secret
	}

	return nil
}

func (sm *SecretManager) saveState() error {
	list := make([]*Secret, 0, len(sm.secrets))
	for _, secret := range sm.secrets {
		list = append(list, secret)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	// Write with restricted permissions
	return os.WriteFile(sm.statePath, data, 0600)
}
