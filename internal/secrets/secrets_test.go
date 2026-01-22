package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretManager_CreateAndGet(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create secret manager
	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	// Test creating a secret
	secretName := "test-secret"
	secretValue := []byte("my-secret-value")

	secret, err := sm.CreateSecret(secretName, secretValue)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	if secret.Name != secretName {
		t.Errorf("Secret name mismatch: got %s, want %s", secret.Name, secretName)
	}

	if secret.ID == "" {
		t.Error("Secret ID should not be empty")
	}

	// Test retrieving the secret
	retrieved, err := sm.GetSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if string(retrieved) != string(secretValue) {
		t.Errorf("Secret value mismatch: got %s, want %s", string(retrieved), string(secretValue))
	}
}

func TestSecretManager_DuplicateSecret(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	secretName := "duplicate-test"
	_, err = sm.CreateSecret(secretName, []byte("value1"))
	if err != nil {
		t.Fatalf("Failed to create first secret: %v", err)
	}

	// Try to create duplicate - should fail
	_, err = sm.CreateSecret(secretName, []byte("value2"))
	if err == nil {
		t.Error("Expected error when creating duplicate secret, got nil")
	}
}

func TestSecretManager_UpdateSecret(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	secretName := "update-test"
	originalValue := []byte("original-value")
	updatedValue := []byte("updated-value")

	_, err = sm.CreateSecret(secretName, originalValue)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Update the secret
	err = sm.UpdateSecret(secretName, updatedValue)
	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	// Verify the update
	retrieved, err := sm.GetSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to get updated secret: %v", err)
	}

	if string(retrieved) != string(updatedValue) {
		t.Errorf("Updated value mismatch: got %s, want %s", string(retrieved), string(updatedValue))
	}
}

func TestSecretManager_RemoveSecret(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	secretName := "remove-test"
	_, err = sm.CreateSecret(secretName, []byte("value"))
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Remove the secret
	err = sm.RemoveSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to remove secret: %v", err)
	}

	// Verify it's gone
	_, err = sm.GetSecret(secretName)
	if err == nil {
		t.Error("Expected error when getting removed secret, got nil")
	}
}

func TestSecretManager_ListSecrets(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	// Create multiple secrets
	secrets := []string{"secret1", "secret2", "secret3"}
	for _, name := range secrets {
		_, err := sm.CreateSecret(name, []byte("value-"+name))
		if err != nil {
			t.Fatalf("Failed to create secret %s: %v", name, err)
		}
	}

	// List secrets
	list := sm.ListSecrets()
	if len(list) != len(secrets) {
		t.Errorf("Expected %d secrets, got %d", len(secrets), len(list))
	}

	// Verify all secrets are in the list
	found := make(map[string]bool)
	for _, s := range list {
		found[s.Name] = true
	}
	for _, name := range secrets {
		if !found[name] {
			t.Errorf("Secret %s not found in list", name)
		}
	}
}

func TestSecretManager_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	secretName := "persist-test"
	secretValue := []byte("persistent-value")

	// Create secret manager and add a secret
	sm1, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create first secret manager: %v", err)
	}

	_, err = sm1.CreateSecret(secretName, secretValue)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Create a new secret manager (simulating restart)
	sm2, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second secret manager: %v", err)
	}

	// Verify the secret persisted
	retrieved, err := sm2.GetSecret(secretName)
	if err != nil {
		t.Fatalf("Failed to get persisted secret: %v", err)
	}

	if string(retrieved) != string(secretValue) {
		t.Errorf("Persisted value mismatch: got %s, want %s", string(retrieved), string(secretValue))
	}
}

func TestSecretManager_GetSecretEnv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm, err := NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	secretName := "env-test"
	secretValue := []byte("env-value")
	envName := "MY_SECRET"

	_, err = sm.CreateSecret(secretName, secretValue)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	envVar, err := sm.GetSecretEnv(secretName, envName)
	if err != nil {
		t.Fatalf("Failed to get secret env: %v", err)
	}

	expected := "MY_SECRET=env-value"
	if envVar != expected {
		t.Errorf("Env var mismatch: got %s, want %s", envVar, expected)
	}
}

func TestSecretManager_FilePermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-secrets-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = NewSecretManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create secret manager: %v", err)
	}

	// Check directory permissions
	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	// Directory should have 0700 permissions
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("Directory permissions: got %o, want 0700", perm)
	}

	// Check key file permissions
	keyPath := filepath.Join(tmpDir, ".secrets.key")
	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}

	keyPerm := keyInfo.Mode().Perm()
	if keyPerm != 0600 {
		t.Errorf("Key file permissions: got %o, want 0600", keyPerm)
	}
}
