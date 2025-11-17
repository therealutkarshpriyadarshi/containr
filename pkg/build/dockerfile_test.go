package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDockerfile(t *testing.T) {
	// Create temporary Dockerfile
	tmpDir, err := os.MkdirTemp("", "containr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfileContent := `# Test Dockerfile
FROM alpine:latest AS builder
RUN apk add --no-cache gcc
COPY --from=previous /app /app
WORKDIR /app

FROM alpine:latest
ARG VERSION=1.0.0
ENV APP_VERSION=$VERSION
COPY --from=builder /app /app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
CMD ["--port", "8080"]
`

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Parse Dockerfile
	parser := NewParser()
	dockerfile, err := parser.ParseFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to parse Dockerfile: %v", err)
	}

	// Test number of stages
	if len(dockerfile.Stages) != 2 {
		t.Errorf("Number of stages = %d, want 2", len(dockerfile.Stages))
	}

	// Test first stage
	stage1 := dockerfile.Stages[0]
	if stage1.Name != "builder" {
		t.Errorf("Stage 1 name = %s, want builder", stage1.Name)
	}
	if stage1.BaseImage != "alpine:latest" {
		t.Errorf("Stage 1 base image = %s, want alpine:latest", stage1.BaseImage)
	}

	// Test second stage
	stage2 := dockerfile.Stages[1]
	if stage2.BaseImage != "alpine:latest" {
		t.Errorf("Stage 2 base image = %s, want alpine:latest", stage2.BaseImage)
	}

	// Test ARG parsing
	if version, ok := dockerfile.Args["VERSION"]; !ok || version != "1.0.0" {
		t.Errorf("ARG VERSION = %s, want 1.0.0", version)
	}

	// Test total instructions
	if len(dockerfile.Instructions) < 5 {
		t.Errorf("Number of instructions = %d, want at least 5", len(dockerfile.Instructions))
	}
}

func TestParseInstruction(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		line      string
		wantCmd   string
		wantArgs  []string
		wantFlags map[string]string
		wantErr   bool
	}{
		{
			name:     "FROM instruction",
			line:     "FROM alpine:latest",
			wantCmd:  "FROM",
			wantArgs: []string{"alpine:latest"},
		},
		{
			name:     "FROM with AS",
			line:     "FROM ubuntu:20.04 AS builder",
			wantCmd:  "FROM",
			wantArgs: []string{"ubuntu:20.04", "AS", "builder"},
		},
		{
			name:     "RUN instruction",
			line:     "RUN apk add gcc",
			wantCmd:  "RUN",
			wantArgs: []string{"apk", "add", "gcc"},
		},
		{
			name:      "COPY with flag",
			line:      "COPY --from=builder /app /app",
			wantCmd:   "COPY",
			wantArgs:  []string{"/app", "/app"},
			wantFlags: map[string]string{"from": "builder"},
		},
		{
			name:     "ENV instruction",
			line:     "ENV PATH=/usr/local/bin",
			wantCmd:  "ENV",
			wantArgs: []string{"PATH=/usr/local/bin"},
		},
		{
			name:     "EXPOSE instruction",
			line:     "EXPOSE 8080",
			wantCmd:  "EXPOSE",
			wantArgs: []string{"8080"},
		},
		{
			name:    "invalid instruction",
			line:    "INVALID instruction",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr, err := parser.parseInstruction(tt.line, 1)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseInstruction() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseInstruction() unexpected error: %v", err)
				return
			}

			if instr.Command != tt.wantCmd {
				t.Errorf("Command = %s, want %s", instr.Command, tt.wantCmd)
			}

			if len(instr.Args) != len(tt.wantArgs) {
				t.Errorf("Args length = %d, want %d", len(instr.Args), len(tt.wantArgs))
			} else {
				for i, arg := range instr.Args {
					if arg != tt.wantArgs[i] {
						t.Errorf("Args[%d] = %s, want %s", i, arg, tt.wantArgs[i])
					}
				}
			}

			if tt.wantFlags != nil {
				for k, v := range tt.wantFlags {
					if instr.Flags[k] != v {
						t.Errorf("Flags[%s] = %s, want %s", k, instr.Flags[k], v)
					}
				}
			}
		})
	}
}

func TestGetStageByName(t *testing.T) {
	dockerfile := &Dockerfile{
		Stages: []*BuildStage{
			{Name: "builder", BaseImage: "alpine"},
			{Name: "final", BaseImage: "scratch"},
		},
	}

	stage := dockerfile.GetStageByName("builder")
	if stage == nil {
		t.Error("GetStageByName(builder) returned nil")
	} else if stage.BaseImage != "alpine" {
		t.Errorf("Stage base image = %s, want alpine", stage.BaseImage)
	}

	stage = dockerfile.GetStageByName("nonexistent")
	if stage != nil {
		t.Error("GetStageByName(nonexistent) should return nil")
	}
}

func TestGetFinalStage(t *testing.T) {
	dockerfile := &Dockerfile{
		Stages: []*BuildStage{
			{Name: "builder", BaseImage: "alpine"},
			{Name: "final", BaseImage: "scratch"},
		},
	}

	stage := dockerfile.GetFinalStage()
	if stage == nil {
		t.Error("GetFinalStage() returned nil")
	} else if stage.Name != "final" {
		t.Errorf("Final stage name = %s, want final", stage.Name)
	}

	// Test empty Dockerfile
	emptyDockerfile := &Dockerfile{
		Stages: []*BuildStage{},
	}

	stage = emptyDockerfile.GetFinalStage()
	if stage != nil {
		t.Error("GetFinalStage() on empty Dockerfile should return nil")
	}
}
