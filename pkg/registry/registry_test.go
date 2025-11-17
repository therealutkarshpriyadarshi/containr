package registry

import (
	"testing"
)

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedRegistry string
		expectedRepo     string
		expectedTag      string
	}{
		{
			name:             "simple image with tag",
			input:            "alpine:3.14",
			expectedRegistry: "docker.io",
			expectedRepo:     "library/alpine",
			expectedTag:      "3.14",
		},
		{
			name:             "simple image without tag",
			input:            "alpine",
			expectedRegistry: "docker.io",
			expectedRepo:     "library/alpine",
			expectedTag:      "latest",
		},
		{
			name:             "user image with tag",
			input:            "user/myimage:v1.0",
			expectedRegistry: "docker.io",
			expectedRepo:     "user/myimage",
			expectedTag:      "v1.0",
		},
		{
			name:             "custom registry",
			input:            "gcr.io/project/image:tag",
			expectedRegistry: "gcr.io",
			expectedRepo:     "project/image",
			expectedTag:      "tag",
		},
		{
			name:             "image with latest tag",
			input:            "nginx:latest",
			expectedRegistry: "docker.io",
			expectedRepo:     "library/nginx",
			expectedTag:      "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseImageReference(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if ref.Registry != tt.expectedRegistry {
				t.Errorf("Expected registry %s, got %s", tt.expectedRegistry, ref.Registry)
			}

			if ref.Repository != tt.expectedRepo {
				t.Errorf("Expected repository %s, got %s", tt.expectedRepo, ref.Repository)
			}

			if ref.Tag != tt.expectedTag {
				t.Errorf("Expected tag %s, got %s", tt.expectedTag, ref.Tag)
			}
		})
	}
}

func TestImageReferenceString(t *testing.T) {
	tests := []struct {
		name     string
		ref      ImageReference
		expected string
	}{
		{
			name: "docker hub official image",
			ref: ImageReference{
				Registry:   "docker.io",
				Repository: "library/alpine",
				Tag:        "latest",
			},
			expected: "library/alpine:latest",
		},
		{
			name: "custom registry",
			ref: ImageReference{
				Registry:   "gcr.io",
				Repository: "project/image",
				Tag:        "v1.0",
			},
			expected: "gcr.io/project/image:v1.0",
		},
		{
			name: "image with digest",
			ref: ImageReference{
				Registry:   "docker.io",
				Repository: "library/alpine",
				Digest:     "sha256:abc123",
			},
			expected: "library/alpine@sha256:abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ref.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.baseURL != "https://registry-1.docker.io" {
		t.Errorf("Expected baseURL 'https://registry-1.docker.io', got %s", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}
