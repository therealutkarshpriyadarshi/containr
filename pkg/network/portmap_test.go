package network

import (
	"testing"
)

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		name      string
		portStr   string
		wantHost  int
		wantCont  int
		wantProto string
		wantErr   bool
	}{
		{
			name:      "simple port",
			portStr:   "8080",
			wantHost:  8080,
			wantCont:  8080,
			wantProto: "tcp",
			wantErr:   false,
		},
		{
			name:      "host:container",
			portStr:   "8080:80",
			wantHost:  8080,
			wantCont:  80,
			wantProto: "tcp",
			wantErr:   false,
		},
		{
			name:      "with tcp protocol",
			portStr:   "8080:80/tcp",
			wantHost:  8080,
			wantCont:  80,
			wantProto: "tcp",
			wantErr:   false,
		},
		{
			name:      "with udp protocol",
			portStr:   "53:53/udp",
			wantHost:  53,
			wantCont:  53,
			wantProto: "udp",
			wantErr:   false,
		},
		{
			name:    "invalid protocol",
			portStr: "8080:80/sctp",
			wantErr: true,
		},
		{
			name:    "invalid port",
			portStr: "abc:80",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := ParsePortMapping(tt.portStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePortMapping() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePortMapping() unexpected error: %v", err)
				return
			}

			if pm.HostPort != tt.wantHost {
				t.Errorf("HostPort = %d, want %d", pm.HostPort, tt.wantHost)
			}

			if pm.ContainerPort != tt.wantCont {
				t.Errorf("ContainerPort = %d, want %d", pm.ContainerPort, tt.wantCont)
			}

			if pm.Protocol != tt.wantProto {
				t.Errorf("Protocol = %s, want %s", pm.Protocol, tt.wantProto)
			}
		})
	}
}
