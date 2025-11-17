package build

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// Instruction represents a Dockerfile instruction
type Instruction struct {
	Command   string   // FROM, RUN, COPY, etc.
	Args      []string // Arguments to the command
	RawLine   string   // Original line from Dockerfile
	LineNum   int      // Line number in Dockerfile
	Flags     map[string]string // Flags (e.g., --from for COPY)
}

// Dockerfile represents a parsed Dockerfile
type Dockerfile struct {
	Instructions []*Instruction
	Stages       []*BuildStage
	Args         map[string]string // Build arguments
}

// BuildStage represents a build stage in multi-stage builds
type BuildStage struct {
	Name         string
	BaseImage    string
	Instructions []*Instruction
	Index        int
}

// Parser parses Dockerfiles
type Parser struct {
	log *logger.Logger
}

// NewParser creates a new Dockerfile parser
func NewParser() *Parser {
	return &Parser{
		log: logger.New("dockerfile-parser"),
	}
}

// ParseFile parses a Dockerfile from a file path
func (p *Parser) ParseFile(path string) (*Dockerfile, error) {
	p.log.Infof("Parsing Dockerfile: %s", path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Dockerfile: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses a Dockerfile from a reader
func (p *Parser) Parse(file *os.File) (*Dockerfile, error) {
	dockerfile := &Dockerfile{
		Instructions: make([]*Instruction, 0),
		Stages:       make([]*BuildStage, 0),
		Args:         make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentStage *BuildStage
	var continuedLine string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle line continuation
		if strings.HasSuffix(strings.TrimSpace(line), "\\") {
			continuedLine += strings.TrimSuffix(strings.TrimSpace(line), "\\") + " "
			continue
		}

		if continuedLine != "" {
			line = continuedLine + line
			continuedLine = ""
		}

		// Skip empty lines and comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse instruction
		instruction, err := p.parseInstruction(line, lineNum)
		if err != nil {
			p.log.Warnf("Line %d: %v", lineNum, err)
			continue
		}

		dockerfile.Instructions = append(dockerfile.Instructions, instruction)

		// Handle special instructions
		switch instruction.Command {
		case "FROM":
			// Start new build stage
			stage := &BuildStage{
				Index:        len(dockerfile.Stages),
				Instructions: make([]*Instruction, 0),
			}

			if len(instruction.Args) > 0 {
				stage.BaseImage = instruction.Args[0]
			}

			// Check for stage name (FROM image AS name)
			if len(instruction.Args) >= 3 && strings.ToUpper(instruction.Args[1]) == "AS" {
				stage.Name = instruction.Args[2]
			}

			dockerfile.Stages = append(dockerfile.Stages, stage)
			currentStage = stage

		case "ARG":
			// Store build argument
			if len(instruction.Args) > 0 {
				parts := strings.SplitN(instruction.Args[0], "=", 2)
				key := parts[0]
				value := ""
				if len(parts) > 1 {
					value = parts[1]
				}
				dockerfile.Args[key] = value
			}
		}

		// Add instruction to current stage
		if currentStage != nil {
			currentStage.Instructions = append(currentStage.Instructions, instruction)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Dockerfile: %w", err)
	}

	p.log.Infof("Parsed Dockerfile: %d instructions, %d stages", len(dockerfile.Instructions), len(dockerfile.Stages))
	return dockerfile, nil
}

// parseInstruction parses a single Dockerfile instruction
func (p *Parser) parseInstruction(line string, lineNum int) (*Instruction, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty instruction")
	}

	instruction := &Instruction{
		Command: strings.ToUpper(parts[0]),
		Args:    make([]string, 0),
		RawLine: line,
		LineNum: lineNum,
		Flags:   make(map[string]string),
	}

	// Parse arguments
	if len(parts) > 1 {
		// Handle flags (e.g., COPY --from=stage)
		argStart := 1
		for i := 1; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "--") {
				// Parse flag
				flagParts := strings.SplitN(parts[i], "=", 2)
				flagName := strings.TrimPrefix(flagParts[0], "--")
				flagValue := ""
				if len(flagParts) > 1 {
					flagValue = flagParts[1]
				}
				instruction.Flags[flagName] = flagValue
				argStart = i + 1
			} else {
				break
			}
		}

		// Remaining parts are arguments
		if argStart < len(parts) {
			instruction.Args = parts[argStart:]
		}
	}

	// Validate instruction
	if !p.isValidCommand(instruction.Command) {
		return nil, fmt.Errorf("unknown instruction: %s", instruction.Command)
	}

	return instruction, nil
}

// isValidCommand checks if a command is a valid Dockerfile instruction
func (p *Parser) isValidCommand(cmd string) bool {
	validCommands := map[string]bool{
		"FROM":        true,
		"RUN":         true,
		"CMD":         true,
		"LABEL":       true,
		"EXPOSE":      true,
		"ENV":         true,
		"ADD":         true,
		"COPY":        true,
		"ENTRYPOINT":  true,
		"VOLUME":      true,
		"USER":        true,
		"WORKDIR":     true,
		"ARG":         true,
		"ONBUILD":     true,
		"STOPSIGNAL":  true,
		"HEALTHCHECK": true,
		"SHELL":       true,
	}

	return validCommands[cmd]
}

// GetStageByName returns a build stage by name
func (d *Dockerfile) GetStageByName(name string) *BuildStage {
	for _, stage := range d.Stages {
		if stage.Name == name {
			return stage
		}
	}
	return nil
}

// GetFinalStage returns the final build stage
func (d *Dockerfile) GetFinalStage() *BuildStage {
	if len(d.Stages) == 0 {
		return nil
	}
	return d.Stages[len(d.Stages)-1]
}

// String returns a string representation of the instruction
func (i *Instruction) String() string {
	result := i.Command

	// Add flags
	for k, v := range i.Flags {
		result += fmt.Sprintf(" --%s=%s", k, v)
	}

	// Add args
	if len(i.Args) > 0 {
		result += " " + strings.Join(i.Args, " ")
	}

	return result
}
