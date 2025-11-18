// Package ide provides IDE integration and Language Server Protocol support
package ide

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// LSPServer implements the Language Server Protocol for containr configurations
type LSPServer struct {
	conn   io.ReadWriteCloser
	logger *logger.Logger
	mu     sync.RWMutex

	// Workspace state
	documents map[string]*Document

	// Capabilities
	capabilities ServerCapabilities
}

// ServerCapabilities defines what the LSP server can do
type ServerCapabilities struct {
	TextDocumentSync   int                    `json:"textDocumentSync"`
	CompletionProvider *CompletionOptions     `json:"completionProvider,omitempty"`
	HoverProvider      bool                   `json:"hoverProvider,omitempty"`
	DefinitionProvider bool                   `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider bool               `json:"documentSymbolProvider,omitempty"`
	DiagnosticProvider *DiagnosticOptions     `json:"diagnosticProvider,omitempty"`
}

// CompletionOptions defines completion behavior
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
}

// DiagnosticOptions defines diagnostic behavior
type DiagnosticOptions struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

// Document represents an open document in the workspace
type Document struct {
	URI     string
	Version int
	Content string
	Type    DocumentType
}

// DocumentType represents the type of configuration file
type DocumentType string

const (
	DocumentTypeContainrfile DocumentType = "containrfile"
	DocumentTypeDockerfile   DocumentType = "dockerfile"
	DocumentTypeCompose      DocumentType = "compose"
	DocumentTypeConfig       DocumentType = "config"
)

// Position represents a position in a document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic represents a validation issue
type Diagnostic struct {
	Range    Range    `json:"range"`
	Severity int      `json:"severity"`
	Message  string   `json:"message"`
	Source   string   `json:"source"`
}

// DiagnosticSeverity levels
const (
	DiagnosticSeverityError       = 1
	DiagnosticSeverityWarning     = 2
	DiagnosticSeverityInformation = 3
	DiagnosticSeverityHint        = 4
)

// CompletionItem represents an autocomplete suggestion
type CompletionItem struct {
	Label            string   `json:"label"`
	Kind             int      `json:"kind"`
	Detail           string   `json:"detail,omitempty"`
	Documentation    string   `json:"documentation,omitempty"`
	InsertText       string   `json:"insertText,omitempty"`
	InsertTextFormat int      `json:"insertTextFormat,omitempty"`
}

// CompletionItemKind constants
const (
	CompletionItemKindText     = 1
	CompletionItemKindMethod   = 2
	CompletionItemKindFunction = 3
	CompletionItemKindKeyword  = 14
	CompletionItemKindSnippet  = 15
)

// NewLSPServer creates a new LSP server
func NewLSPServer(conn io.ReadWriteCloser) *LSPServer {
	return &LSPServer{
		conn:      conn,
		logger:    logger.New("lsp"),
		documents: make(map[string]*Document),
		capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full sync
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", "-", " "},
				ResolveProvider:   true,
			},
			HoverProvider:          true,
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
			DiagnosticProvider: &DiagnosticOptions{
				InterFileDependencies: true,
				WorkspaceDiagnostics:  true,
			},
		},
	}
}

// Start starts the LSP server
func (s *LSPServer) Start(ctx context.Context) error {
	s.logger.Info("Starting LSP server")

	// Handle incoming messages
	decoder := json.NewDecoder(s.conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var msg map[string]interface{}
			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					return nil
				}
				s.logger.Error("Failed to decode message", "error", err)
				continue
			}

			if err := s.handleMessage(ctx, msg); err != nil {
				s.logger.Error("Failed to handle message", "error", err)
			}
		}
	}
}

// handleMessage handles an LSP message
func (s *LSPServer) handleMessage(ctx context.Context, msg map[string]interface{}) error {
	method, ok := msg["method"].(string)
	if !ok {
		return fmt.Errorf("message missing method")
	}

	switch method {
	case "initialize":
		return s.handleInitialize(msg)
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	case "textDocument/diagnostic":
		return s.handleDiagnostic(msg)
	default:
		s.logger.Debug("Unhandled method", "method", method)
	}

	return nil
}

// handleInitialize handles the initialize request
func (s *LSPServer) handleInitialize(msg map[string]interface{}) error {
	id := msg["id"]

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"capabilities": s.capabilities,
			"serverInfo": map[string]string{
				"name":    "containr-lsp",
				"version": "1.0.0",
			},
		},
	}

	return s.sendResponse(response)
}

// handleDidOpen handles document open events
func (s *LSPServer) handleDidOpen(msg map[string]interface{}) error {
	params := msg["params"].(map[string]interface{})
	textDocument := params["textDocument"].(map[string]interface{})

	uri := textDocument["uri"].(string)
	version := int(textDocument["version"].(float64))
	text := textDocument["text"].(string)

	s.mu.Lock()
	s.documents[uri] = &Document{
		URI:     uri,
		Version: version,
		Content: text,
		Type:    detectDocumentType(uri),
	}
	s.mu.Unlock()

	// Send diagnostics
	return s.publishDiagnostics(uri)
}

// handleDidChange handles document change events
func (s *LSPServer) handleDidChange(msg map[string]interface{}) error {
	params := msg["params"].(map[string]interface{})
	textDocument := params["textDocument"].(map[string]interface{})
	contentChanges := params["contentChanges"].([]interface{})

	uri := textDocument["uri"].(string)
	version := int(textDocument["version"].(float64))

	if len(contentChanges) > 0 {
		change := contentChanges[0].(map[string]interface{})
		text := change["text"].(string)

		s.mu.Lock()
		if doc, ok := s.documents[uri]; ok {
			doc.Version = version
			doc.Content = text
		}
		s.mu.Unlock()
	}

	// Send diagnostics
	return s.publishDiagnostics(uri)
}

// handleDidClose handles document close events
func (s *LSPServer) handleDidClose(msg map[string]interface{}) error {
	params := msg["params"].(map[string]interface{})
	textDocument := params["textDocument"].(map[string]interface{})
	uri := textDocument["uri"].(string)

	s.mu.Lock()
	delete(s.documents, uri)
	s.mu.Unlock()

	return nil
}

// handleCompletion handles completion requests
func (s *LSPServer) handleCompletion(msg map[string]interface{}) error {
	id := msg["id"]
	params := msg["params"].(map[string]interface{})
	textDocument := params["textDocument"].(map[string]interface{})
	position := params["position"].(map[string]interface{})

	uri := textDocument["uri"].(string)
	line := int(position["line"].(float64))
	character := int(position["character"].(float64))

	items := s.getCompletionItems(uri, line, character)

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  items,
	}

	return s.sendResponse(response)
}

// handleHover handles hover requests
func (s *LSPServer) handleHover(msg map[string]interface{}) error {
	id := msg["id"]
	// TODO: Implement hover support

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  nil,
	}

	return s.sendResponse(response)
}

// handleDefinition handles go-to-definition requests
func (s *LSPServer) handleDefinition(msg map[string]interface{}) error {
	id := msg["id"]
	// TODO: Implement definition support

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  nil,
	}

	return s.sendResponse(response)
}

// handleDiagnostic handles diagnostic requests
func (s *LSPServer) handleDiagnostic(msg map[string]interface{}) error {
	id := msg["id"]
	params := msg["params"].(map[string]interface{})
	textDocument := params["textDocument"].(map[string]interface{})
	uri := textDocument["uri"].(string)

	diagnostics := s.getDiagnostics(uri)

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"kind":  "full",
			"items": diagnostics,
		},
	}

	return s.sendResponse(response)
}

// publishDiagnostics publishes diagnostics for a document
func (s *LSPServer) publishDiagnostics(uri string) error {
	diagnostics := s.getDiagnostics(uri)

	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		},
	}

	return s.sendResponse(notification)
}

// getDiagnostics gets diagnostics for a document
func (s *LSPServer) getDiagnostics(uri string) []Diagnostic {
	s.mu.RLock()
	_, ok := s.documents[uri]
	s.mu.RUnlock()

	if !ok {
		return []Diagnostic{}
	}

	// TODO: Implement validation based on document type
	// For now, return empty diagnostics
	return []Diagnostic{}
}

// getCompletionItems gets completion items for a position
func (s *LSPServer) getCompletionItems(uri string, line, character int) []CompletionItem {
	s.mu.RLock()
	doc, ok := s.documents[uri]
	s.mu.RUnlock()

	if !ok {
		return []CompletionItem{}
	}

	switch doc.Type {
	case DocumentTypeDockerfile, DocumentTypeContainrfile:
		return getDockerfileCompletions()
	case DocumentTypeCompose:
		return getComposeCompletions()
	default:
		return []CompletionItem{}
	}
}

// getDockerfileCompletions returns Dockerfile completion items
func getDockerfileCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:  "FROM",
			Kind:   CompletionItemKindKeyword,
			Detail: "Set the base image",
			Documentation: "FROM instruction initializes a new build stage and sets the Base Image",
			InsertText: "FROM ${1:image}:${2:tag}",
			InsertTextFormat: 2, // Snippet
		},
		{
			Label:  "RUN",
			Kind:   CompletionItemKindKeyword,
			Detail: "Execute command",
			Documentation: "RUN instruction will execute any commands in a new layer",
			InsertText: "RUN ${1:command}",
			InsertTextFormat: 2,
		},
		{
			Label:  "CMD",
			Kind:   CompletionItemKindKeyword,
			Detail: "Default command",
			Documentation: "CMD provides defaults for an executing container",
			InsertText: "CMD [\"${1:executable}\", \"${2:param}\"]",
			InsertTextFormat: 2,
		},
		{
			Label:  "COPY",
			Kind:   CompletionItemKindKeyword,
			Detail: "Copy files",
			Documentation: "COPY instruction copies files or directories from source to destination",
			InsertText: "COPY ${1:src} ${2:dest}",
			InsertTextFormat: 2,
		},
		{
			Label:  "WORKDIR",
			Kind:   CompletionItemKindKeyword,
			Detail: "Set working directory",
			Documentation: "WORKDIR sets the working directory for instructions that follow",
			InsertText: "WORKDIR ${1:/app}",
			InsertTextFormat: 2,
		},
		{
			Label:  "EXPOSE",
			Kind:   CompletionItemKindKeyword,
			Detail: "Expose port",
			Documentation: "EXPOSE informs Docker that the container listens on specified network ports",
			InsertText: "EXPOSE ${1:80}",
			InsertTextFormat: 2,
		},
		{
			Label:  "ENV",
			Kind:   CompletionItemKindKeyword,
			Detail: "Set environment variable",
			Documentation: "ENV sets environment variable",
			InsertText: "ENV ${1:key}=${2:value}",
			InsertTextFormat: 2,
		},
	}
}

// getComposeCompletions returns compose file completion items
func getComposeCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:  "version",
			Kind:   CompletionItemKindKeyword,
			Detail: "Compose file version",
			InsertText: "version: \"${1:3.8}\"",
			InsertTextFormat: 2,
		},
		{
			Label:  "services",
			Kind:   CompletionItemKindKeyword,
			Detail: "Define services",
			InsertText: "services:\n  ${1:service}:\n    image: ${2:image}",
			InsertTextFormat: 2,
		},
		{
			Label:  "image",
			Kind:   CompletionItemKindKeyword,
			Detail: "Container image",
			InsertText: "image: ${1:image}:${2:tag}",
			InsertTextFormat: 2,
		},
		{
			Label:  "build",
			Kind:   CompletionItemKindKeyword,
			Detail: "Build configuration",
			InsertText: "build:\n  context: ${1:.}\n  dockerfile: ${2:Dockerfile}",
			InsertTextFormat: 2,
		},
		{
			Label:  "ports",
			Kind:   CompletionItemKindKeyword,
			Detail: "Port mapping",
			InsertText: "ports:\n  - \"${1:8080}:${2:80}\"",
			InsertTextFormat: 2,
		},
		{
			Label:  "volumes",
			Kind:   CompletionItemKindKeyword,
			Detail: "Volume mounts",
			InsertText: "volumes:\n  - ${1:./data}:${2:/data}",
			InsertTextFormat: 2,
		},
		{
			Label:  "environment",
			Kind:   CompletionItemKindKeyword,
			Detail: "Environment variables",
			InsertText: "environment:\n  ${1:KEY}: ${2:value}",
			InsertTextFormat: 2,
		},
	}
}

// sendResponse sends a response message
func (s *LSPServer) sendResponse(msg map[string]interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Add Content-Length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	if _, err := s.conn.Write([]byte(header)); err != nil {
		return err
	}

	if _, err := s.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// detectDocumentType detects the document type from URI
func detectDocumentType(uri string) DocumentType {
	if contains(uri, "Dockerfile") || contains(uri, "Containrfile") {
		return DocumentTypeDockerfile
	}
	if contains(uri, "docker-compose") || contains(uri, "compose") {
		return DocumentTypeCompose
	}
	return DocumentTypeConfig
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
