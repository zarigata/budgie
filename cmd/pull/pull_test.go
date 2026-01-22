package pull

import (
	"testing"
)

func TestNormalizeImageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple image name
		{"nginx", "docker.io/library/nginx:latest"},
		{"alpine", "docker.io/library/alpine:latest"},
		{"redis", "docker.io/library/redis:latest"},

		// Image with tag
		{"nginx:1.21", "docker.io/library/nginx:1.21"},
		{"alpine:3.18", "docker.io/library/alpine:3.18"},

		// User/image format
		{"myuser/myimage", "docker.io/myuser/myimage:latest"},
		{"myuser/myimage:v1.0", "docker.io/myuser/myimage:v1.0"},

		// Full registry path
		{"docker.io/library/nginx", "docker.io/library/nginx:latest"},
		{"docker.io/library/nginx:1.21", "docker.io/library/nginx:1.21"},

		// Other registries
		{"ghcr.io/myorg/myimage", "ghcr.io/myorg/myimage:latest"},
		{"ghcr.io/myorg/myimage:v1.0", "ghcr.io/myorg/myimage:v1.0"},
		{"gcr.io/myproject/myimage", "gcr.io/myproject/myimage:latest"},
		{"quay.io/myorg/myimage:latest", "quay.io/myorg/myimage:latest"},

		// Private registry with port (port makes it look like has tag already)
		{"myregistry.com:5000/myimage", "myregistry.com:5000/myimage"},
		{"myregistry.com:5000/myimage:v1", "myregistry.com:5000/myimage:v1"},

		// Digest (adds docker.io/library/ prefix but preserves digest)
		{"nginx@sha256:abc123", "docker.io/library/nginx@sha256:abc123"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeImageName(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeImageName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatSize(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tc.bytes, result, tc.expected)
			}
		})
	}
}
