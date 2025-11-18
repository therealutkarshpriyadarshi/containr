package ide

import (
	"testing"
)

func TestLSPServer_New(t *testing.T) {
	server := NewLSPServer(nil)
	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.capabilities.HoverProvider != true {
		t.Error("Expected hover provider to be enabled")
	}

	if server.capabilities.CompletionProvider == nil {
		t.Error("Expected completion provider to be enabled")
	}
}

func TestGetDockerfileCompletions(t *testing.T) {
	completions := getDockerfileCompletions()

	if len(completions) == 0 {
		t.Fatal("Expected Dockerfile completions")
	}

	// Check for FROM instruction
	found := false
	for _, item := range completions {
		if item.Label == "FROM" {
			found = true
			if item.Kind != CompletionItemKindKeyword {
				t.Error("Expected FROM to be keyword kind")
			}
			break
		}
	}

	if !found {
		t.Error("Expected FROM instruction in completions")
	}
}

func TestGetComposeCompletions(t *testing.T) {
	completions := getComposeCompletions()

	if len(completions) == 0 {
		t.Fatal("Expected compose completions")
	}

	// Check for services keyword
	found := false
	for _, item := range completions {
		if item.Label == "services" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected services keyword in completions")
	}
}

func TestDetectDocumentType(t *testing.T){
	tests := []struct {
		uri      string
		expected DocumentType
	}{
		{"file:///path/to/Dockerfile", DocumentTypeDockerfile},
		{"file:///path/to/Containrfile", DocumentTypeDockerfile},
		{"file:///path/to/docker-compose.yml", DocumentTypeCompose},
		{"file:///path/to/config.yaml", DocumentTypeConfig},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			result := detectDocumentType(tt.uri)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
