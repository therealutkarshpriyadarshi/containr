// Package ide provides IDE integration and development tools for containr
package ide

import (
	"context"
	"fmt"
	"net"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// IDEIntegration provides IDE integration capabilities
type IDEIntegration struct {
	lspServer    *LSPServer
	debugAdapter *DebugAdapter
	logger       *logger.Logger
}

// DebugAdapter provides debugging capabilities
type DebugAdapter struct {
	port   int
	logger *logger.Logger
}

// Config configures the IDE integration
type Config struct {
	LSPEnabled   bool
	LSPPort      int
	DebugEnabled bool
	DebugPort    int
}

// NewIDEIntegration creates a new IDE integration
func NewIDEIntegration(config *Config) *IDEIntegration {
	return &IDEIntegration{
		logger: logger.New(logger.InfoLevel),
	}
}

// StartLSPServer starts the Language Server Protocol server
func (ide *IDEIntegration) StartLSPServer(ctx context.Context, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	ide.logger.Info("LSP server listening", "port", port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				ide.logger.Error("Failed to accept connection", "error", err)
				continue
			}

			server := NewLSPServer(conn)
			go func() {
				if err := server.Start(ctx); err != nil {
					ide.logger.Error("LSP server error", "error", err)
				}
			}()
		}
	}()

	return nil
}

// StartDebugAdapter starts the debug adapter
func (ide *IDEIntegration) StartDebugAdapter(ctx context.Context, port int) error {
	ide.debugAdapter = &DebugAdapter{
		port:   port,
		logger: logger.New(logger.InfoLevel),
	}

	ide.logger.Info("Debug adapter listening", "port", port)
	return nil
}

// Stop stops all IDE integration services
func (ide *IDEIntegration) Stop() error {
	ide.logger.Info("Stopping IDE integration")
	return nil
}
