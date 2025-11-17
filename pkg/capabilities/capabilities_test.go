package capabilities

import (
	"testing"
)

func TestDefaultCapabilities(t *testing.T) {
	caps := DefaultCapabilities()

	if len(caps) == 0 {
		t.Fatal("DefaultCapabilities returned empty list")
	}

	// Check that default set includes essential capabilities
	expectedCaps := []Capability{
		CAP_CHOWN,
		CAP_SETUID,
		CAP_SETGID,
		CAP_NET_BIND_SERVICE,
		CAP_KILL,
	}

	for _, expected := range expectedCaps {
		found := false
		for _, cap := range caps {
			if cap == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultCapabilities missing expected capability: %s", expected)
		}
	}
}

func TestAllCapabilities(t *testing.T) {
	caps := AllCapabilities()

	if len(caps) == 0 {
		t.Fatal("AllCapabilities returned empty list")
	}

	// Should have more than default
	if len(caps) <= len(DefaultCapabilities()) {
		t.Error("AllCapabilities should have more capabilities than DefaultCapabilities")
	}

	// Check that dangerous capabilities are included
	dangerousCaps := []Capability{
		CAP_SYS_ADMIN,
		CAP_SYS_MODULE,
		CAP_SYS_BOOT,
	}

	for _, dangerous := range dangerousCaps {
		found := false
		for _, cap := range caps {
			if cap == dangerous {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllCapabilities missing capability: %s", dangerous)
		}
	}
}

func TestConfig_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected int
		wantErr  bool
	}{
		{
			name:     "default configuration",
			config:   &Config{},
			expected: len(DefaultCapabilities()),
			wantErr:  false,
		},
		{
			name: "drop single capability",
			config: &Config{
				Drop: []Capability{CAP_NET_RAW},
			},
			expected: len(DefaultCapabilities()) - 1,
			wantErr:  false,
		},
		{
			name: "add single capability",
			config: &Config{
				Add: []Capability{CAP_SYS_TIME},
			},
			expected: len(DefaultCapabilities()) + 1,
			wantErr:  false,
		},
		{
			name: "drop and add capabilities",
			config: &Config{
				Drop: []Capability{CAP_NET_RAW, CAP_KILL},
				Add:  []Capability{CAP_SYS_TIME, CAP_SYS_NICE},
			},
			expected: len(DefaultCapabilities()),
			wantErr:  false,
		},
		{
			name: "allow all capabilities",
			config: &Config{
				AllowAll: true,
			},
			expected: len(AllCapabilities()),
			wantErr:  false,
		},
		{
			name: "drop all capabilities",
			config: &Config{
				Drop: DefaultCapabilities(),
			},
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps, err := tt.config.Resolve()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(caps) != tt.expected {
				t.Errorf("Config.Resolve() got %d capabilities, want %d", len(caps), tt.expected)
			}
		})
	}
}

func TestConfig_ResolveNoDuplicates(t *testing.T) {
	config := &Config{
		Add: []Capability{CAP_CHOWN, CAP_CHOWN}, // Duplicate
	}

	caps, err := config.Resolve()
	if err != nil {
		t.Fatalf("Config.Resolve() error = %v", err)
	}

	// Check for duplicates
	seen := make(map[Capability]bool)
	for _, cap := range caps {
		if seen[cap] {
			t.Errorf("Config.Resolve() returned duplicate capability: %s", cap)
		}
		seen[cap] = true
	}
}

func TestParseCapability(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Capability
		wantErr bool
	}{
		{
			name:    "with CAP_ prefix",
			input:   "CAP_NET_ADMIN",
			want:    CAP_NET_ADMIN,
			wantErr: false,
		},
		{
			name:    "without CAP_ prefix",
			input:   "NET_ADMIN",
			want:    CAP_NET_ADMIN,
			wantErr: false,
		},
		{
			name:    "lowercase",
			input:   "net_admin",
			want:    CAP_NET_ADMIN,
			wantErr: false,
		},
		{
			name:    "invalid capability",
			input:   "CAP_INVALID",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCapability(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCapability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveCapabilities(t *testing.T) {
	caps := []Capability{
		CAP_CHOWN,
		CAP_NET_ADMIN,
		CAP_SYS_ADMIN,
		CAP_KILL,
	}

	remove := []Capability{
		CAP_NET_ADMIN,
		CAP_SYS_ADMIN,
	}

	result := removeCapabilities(caps, remove)

	if len(result) != 2 {
		t.Errorf("removeCapabilities() got %d capabilities, want 2", len(result))
	}

	for _, cap := range result {
		if cap == CAP_NET_ADMIN || cap == CAP_SYS_ADMIN {
			t.Errorf("removeCapabilities() should have removed %s", cap)
		}
	}
}

func TestAddCapabilities(t *testing.T) {
	caps := []Capability{
		CAP_CHOWN,
		CAP_KILL,
	}

	add := []Capability{
		CAP_NET_ADMIN,
		CAP_SYS_ADMIN,
	}

	result := addCapabilities(caps, add)

	if len(result) != 4 {
		t.Errorf("addCapabilities() got %d capabilities, want 4", len(result))
	}

	expected := map[Capability]bool{
		CAP_CHOWN:     true,
		CAP_KILL:      true,
		CAP_NET_ADMIN: true,
		CAP_SYS_ADMIN: true,
	}

	for _, cap := range result {
		if !expected[cap] {
			t.Errorf("addCapabilities() returned unexpected capability: %s", cap)
		}
	}
}

func TestAddCapabilitiesNoDuplicates(t *testing.T) {
	caps := []Capability{
		CAP_CHOWN,
		CAP_KILL,
	}

	add := []Capability{
		CAP_CHOWN, // Already exists
		CAP_NET_ADMIN,
	}

	result := addCapabilities(caps, add)

	if len(result) != 3 {
		t.Errorf("addCapabilities() got %d capabilities, want 3", len(result))
	}

	// Check for duplicates
	seen := make(map[Capability]bool)
	for _, cap := range result {
		if seen[cap] {
			t.Errorf("addCapabilities() returned duplicate capability: %s", cap)
		}
		seen[cap] = true
	}
}

func TestCapabilityString(t *testing.T) {
	cap := CAP_NET_ADMIN
	if cap.String() != "CAP_NET_ADMIN" {
		t.Errorf("Capability.String() = %s, want CAP_NET_ADMIN", cap.String())
	}
}

func BenchmarkResolveDefault(b *testing.B) {
	config := &Config{}
	for i := 0; i < b.N; i++ {
		_, _ = config.Resolve()
	}
}

func BenchmarkResolveComplex(b *testing.B) {
	config := &Config{
		Drop: []Capability{CAP_NET_RAW, CAP_KILL},
		Add:  []Capability{CAP_SYS_TIME, CAP_SYS_NICE, CAP_SYS_RESOURCE},
	}
	for i := 0; i < b.N; i++ {
		_, _ = config.Resolve()
	}
}
