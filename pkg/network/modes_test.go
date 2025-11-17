package network

import (
	"testing"
)

func TestParseNetworkMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		want    NetworkMode
		wantErr bool
	}{
		{
			name:    "none mode",
			mode:    "none",
			want:    NetworkModeNone,
			wantErr: false,
		},
		{
			name:    "host mode",
			mode:    "host",
			want:    NetworkModeHost,
			wantErr: false,
		},
		{
			name:    "bridge mode",
			mode:    "bridge",
			want:    NetworkModeBridge,
			wantErr: false,
		},
		{
			name:    "empty defaults to bridge",
			mode:    "",
			want:    NetworkModeBridge,
			wantErr: false,
		},
		{
			name:    "container mode",
			mode:    "container:abc123",
			want:    NetworkModeContainer,
			wantErr: false,
		},
		{
			name:    "invalid mode",
			mode:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNetworkMode(tt.mode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseNetworkMode() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseNetworkMode() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseNetworkMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNetworkNamespaceFlags(t *testing.T) {
	tests := []struct {
		name    string
		mode    NetworkMode
		want    bool
		wantErr bool
	}{
		{
			name:    "none mode creates namespace",
			mode:    NetworkModeNone,
			want:    true,
			wantErr: false,
		},
		{
			name:    "host mode no namespace",
			mode:    NetworkModeHost,
			want:    false,
			wantErr: false,
		},
		{
			name:    "bridge mode creates namespace",
			mode:    NetworkModeBridge,
			want:    true,
			wantErr: false,
		},
		{
			name:    "container mode no namespace",
			mode:    NetworkModeContainer,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNetworkNamespaceFlags(tt.mode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetNetworkNamespaceFlags() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetNetworkNamespaceFlags() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("GetNetworkNamespaceFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
