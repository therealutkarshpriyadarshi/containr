# IDE Integration Package

This package provides IDE integration capabilities for containr, including Language Server Protocol (LSP) support and debug adapter functionality.

## Features

- **Language Server Protocol (LSP)**: IntelliSense, autocompletion, and validation for Dockerfiles and containr configurations
- **Debug Adapter Protocol (DAP)**: Interactive debugging support for containers
- **Syntax Highlighting**: Real-time syntax validation
- **Go-to-Definition**: Navigate to image and volume definitions
- **Hover Information**: Contextual help for instructions and configurations

## Usage

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/ide"

// Create IDE integration
config := &ide.Config{
    LSPEnabled:   true,
    LSPPort:      4389,
    DebugEnabled: true,
    DebugPort:    4711,
}

integration := ide.NewIDEIntegration(config)

// Start LSP server
err := integration.StartLSPServer(ctx, 4389)

// Start debug adapter
err = integration.StartDebugAdapter(ctx, 4711)
```

## LSP Features

### Autocompletion

Provides intelligent autocompletion for:
- Dockerfile instructions (FROM, RUN, COPY, etc.)
- Compose file keywords
- Containr configuration options

### Diagnostics

Real-time validation and error checking for:
- Syntax errors
- Invalid instruction usage
- Deprecated instructions
- Security best practices

### Hover Information

Contextual documentation when hovering over:
- Instructions
- Image names
- Configuration keys

## IDE Support

Compatible with:
- Visual Studio Code
- Neovim (with LSP support)
- Emacs (with lsp-mode)
- Any IDE supporting LSP
