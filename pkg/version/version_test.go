package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("expected non-empty version")
	}

	if info.GoVersion == "" {
		t.Error("expected non-empty Go version")
	}

	if info.Platform == "" {
		t.Error("expected non-empty platform")
	}
}

func TestString(t *testing.T) {
	info := Get()
	str := info.String()

	if !strings.Contains(str, "containr version") {
		t.Error("expected version string to contain 'containr version'")
	}

	if !strings.Contains(str, info.Version) {
		t.Error("expected version string to contain version number")
	}

	if !strings.Contains(str, info.GoVersion) {
		t.Error("expected version string to contain Go version")
	}
}

func TestShort(t *testing.T) {
	// Set a known commit for testing
	originalCommit := GitCommit
	GitCommit = "1234567890abcdef"
	defer func() { GitCommit = originalCommit }()

	info := Get()
	short := info.Short()

	if !strings.Contains(short, "containr") {
		t.Error("expected short version to contain 'containr'")
	}

	if !strings.Contains(short, info.Version) {
		t.Error("expected short version to contain version number")
	}

	if !strings.Contains(short, "1234567") {
		t.Error("expected short version to contain short commit hash")
	}
}

func TestUserAgent(t *testing.T) {
	info := Get()
	ua := info.UserAgent()

	if !strings.HasPrefix(ua, "containr/") {
		t.Error("expected user agent to start with 'containr/'")
	}

	if !strings.Contains(ua, info.Version) {
		t.Error("expected user agent to contain version")
	}

	if !strings.Contains(ua, info.Platform) {
		t.Error("expected user agent to contain platform")
	}
}

func TestInfoFields(t *testing.T) {
	info := Get()

	tests := []struct {
		name  string
		value string
	}{
		{"Version", info.Version},
		{"GitCommit", info.GitCommit},
		{"BuildDate", info.BuildDate},
		{"GoVersion", info.GoVersion},
		{"Platform", info.Platform},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}
