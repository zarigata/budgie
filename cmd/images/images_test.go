package images

import (
	"testing"
	"time"
)

func TestParseImageName(t *testing.T) {
	tests := []struct {
		input        string
		expectedRepo string
		expectedTag  string
	}{
		// Simple image with tag
		{"nginx:latest", "nginx", "latest"},
		{"nginx:1.21", "nginx", "1.21"},
		{"alpine:3.18", "alpine", "3.18"},

		// Full path with tag
		{"docker.io/library/nginx:latest", "docker.io/library/nginx", "latest"},
		{"ghcr.io/myorg/myimage:v1.0", "ghcr.io/myorg/myimage", "v1.0"},

		// Without tag (should return "latest")
		{"nginx", "nginx", "latest"},
		{"docker.io/library/nginx", "docker.io/library/nginx", "latest"},

		// With digest (should return "<none>")
		{"nginx@sha256:abc123def456", "nginx", "<none>"},

		// Registry with port
		{"myregistry.com:5000/myimage:v1", "myregistry.com:5000/myimage", "v1"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			repo, tag := parseImageName(tc.input)
			if repo != tc.expectedRepo {
				t.Errorf("parseImageName(%q) repo = %q, want %q", tc.input, repo, tc.expectedRepo)
			}
			if tag != tc.expectedTag {
				t.Errorf("parseImageName(%q) tag = %q, want %q", tc.input, tag, tc.expectedTag)
			}
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		contains string
	}{
		{now.Add(-30 * time.Second), "Less than a minute ago"},
		{now.Add(-1 * time.Minute), "1 minute ago"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-3 * time.Hour), "3 hours ago"},
		{now.Add(-24 * time.Hour), "1 day ago"},
		{now.Add(-72 * time.Hour), "3 days ago"},
		{now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{now.Add(-30 * 24 * time.Hour), "1 month ago"},
		{now.Add(-60 * 24 * time.Hour), "2 months ago"},
		{now.Add(-365 * 24 * time.Hour), "1 year ago"},
		{now.Add(-730 * 24 * time.Hour), "2 years ago"},
		{time.Time{}, "N/A"},
	}

	for _, tc := range tests {
		t.Run(tc.contains, func(t *testing.T) {
			result := formatTimeAgo(tc.time)
			if result != tc.contains {
				t.Errorf("formatTimeAgo() = %q, want %q", result, tc.contains)
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
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{2048, "2.0 KB"},
		{1048576, "1.0 MB"},
		{10485760, "10.0 MB"},
		{104857600, "100.0 MB"},
		{1073741824, "1.0 GB"},
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
