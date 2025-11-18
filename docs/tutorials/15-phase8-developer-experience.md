# Tutorial: Phase 8 - Developer Experience & Advanced Tooling

## Introduction

This tutorial guides you through Phase 8's developer experience features, covering IDE integration, advanced debugging, security scanning, GitOps deployment, hot reload development, and container testing.

## Prerequisites

- Containr installed and working
- Basic understanding of containers
- Text editor or IDE
- Git installed

## Part 1: IDE Integration

### Setting Up LSP

1. **Start the LSP Server**:
```bash
# Start LSP server on default port
containr ide lsp --port 4389
```

2. **Configure VS Code** (create `.vscode/settings.json`):
```json
{
  "containr.lsp.enabled": true,
  "containr.lsp.port": 4389
}
```

3. **Test Autocomplete**:
Create a `Dockerfile`:
```dockerfile
# Type "FROM" and see suggestions
FROM

# Type "RUN" and see snippets
RUN

# Full example with autocomplete
FROM alpine:latest
WORKDIR /app
COPY . .
RUN apk add --no-cache nodejs npm
CMD ["node", "server.js"]
```

### Using Real-time Validation

The LSP provides real-time error checking:

```dockerfile
# This will show an error (invalid base image)
FROM nonexistent:latest

# This will show a warning (deprecated instruction)
MAINTAINER john@example.com

# This will pass validation
FROM alpine:latest
LABEL maintainer="john@example.com"
```

## Part 2: Advanced Debugging

### Attaching the Debugger

1. **Start a Container**:
```bash
containr run --name myapp -d nginx
```

2. **Attach Debugger**:
```bash
containr debug attach myapp
```

3. **Interactive Debug Session**:
```bash
containr debug interactive myapp

# In the debug session:
(containr-debug) breakpoint add syscall:open
(containr-debug) breakpoint add syscall:write
(containr-debug) continue
```

### System Call Tracing

```bash
# Trace all syscalls
containr debug trace myapp --syscalls all --output trace.log

# Trace specific syscalls
containr debug trace myapp --syscalls open,write,read

# Trace with filters
containr debug trace myapp --syscalls all --filter "path=/etc/*"
```

### Performance Profiling

```bash
# CPU profiling for 30 seconds
containr debug profile cpu myapp --duration 30s --output cpu.prof

# Memory profiling
containr debug profile memory myapp --output mem.prof

# Analyze with pprof
go tool pprof -http=:8080 cpu.prof
```

## Part 3: SBOM & Security Scanning

### Generating SBOM

1. **Generate SBOM in SPDX Format**:
```bash
containr sbom generate myapp:latest --format spdx --output sbom.json
```

2. **Generate in CycloneDX Format**:
```bash
containr sbom generate myapp:latest --format cyclonedx --output sbom.xml
```

3. **View SBOM**:
```bash
cat sbom.json | jq '.packages[] | {name, version, licenses}'
```

### Vulnerability Scanning

1. **Scan Image**:
```bash
containr scan myapp:latest --scanner trivy
```

2. **Filter by Severity**:
```bash
containr scan myapp:latest --min-severity high
```

3. **Generate Scan Report**:
```bash
containr scan myapp:latest --output scan-report.json --format json
```

### Compliance Checking

1. **Create Compliance Policy** (`compliance-policy.yaml`):
```yaml
licenses:
  allowed:
    - MIT
    - Apache-2.0
    - BSD-3-Clause
  denied:
    - GPL
    - AGPL

security:
  max_severity: medium
  block_on_critical: true
```

2. **Run Compliance Check**:
```bash
containr compliance check myapp:latest --policy compliance-policy.yaml
```

## Part 4: GitOps Deployment

### Setting Up GitOps

1. **Create Deployment Repository**:
```bash
mkdir deployments && cd deployments
git init
mkdir manifests
```

2. **Create Deployment Manifest** (`manifests/myapp.yaml`):
```yaml
apiVersion: containr.io/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  replicas: 3
  image: myapp:v1.0.0
  environment:
    PORT: "8080"
  resources:
    limits:
      cpu: "2"
      memory: "4Gi"
```

3. **Commit and Push**:
```bash
git add manifests/myapp.yaml
git commit -m "Add myapp deployment"
git push origin main
```

### Starting GitOps Controller

```bash
# Start GitOps controller
containr gitops start \
    --repo https://github.com/example/deployments \
    --branch main \
    --interval 30s
```

### Managing Deployments

```bash
# List deployments
containr gitops list

# View deployment status
containr gitops status myapp

# Manual sync
containr gitops sync

# Rollback
containr gitops rollback myapp --revision abc123
```

## Part 5: Hot Reload Development

### Setting Up Development Environment

1. **Create Development Configuration** (`.containr/dev.yaml`):
```yaml
name: myapp-dev
image: node:18-alpine

watch:
  - path: ./src
    target: /app/src
    ignore:
      - "*.tmp"
      - "node_modules"

reload:
  strategy: signal
  signal: SIGHUP
  debounce: 500ms

ports:
  - "3000:3000"
  - "9229:9229"  # Debug port

environment:
  NODE_ENV: development
  DEBUG: "*"
```

2. **Start Development Mode**:
```bash
containr dev myapp \
    --watch ./src:/app/src \
    --reload-strategy signal \
    --ignore "*.tmp,node_modules"
```

### Using Hot Reload

1. **Edit Source Code**:
```javascript
// src/server.js
const express = require('express');
const app = express();

app.get('/', (req, res) => {
    res.send('Hello World!');  // Change this
});

app.listen(3000);
```

2. **Changes Are Automatically Detected and Reloaded**

3. **View Reload Statistics**:
```bash
containr dev stats myapp-dev
```

### Development Templates

```bash
# Create Node.js development environment
containr dev template create nodejs --name myapp-dev

# Create Python development environment
containr dev template create python --name pyapp-dev

# Create Go development environment
containr dev template create golang --name goapp-dev
```

## Part 6: Container Testing

### Writing Container Tests

1. **Create Test File** (`tests/web-server.yaml`):
```yaml
name: web-server-test
image: nginx:latest
timeout: 30s

assertions:
  - type: exit_code
    value: 0

  - type: port_open
    port: 80

  - type: file_exists
    path: /etc/nginx/nginx.conf

  - type: output_contains
    stream: stdout
    text: "nginx"
```

2. **Run Tests**:
```bash
containr test run ./tests/
```

### Behavior-Driven Testing

Create a BDD test (`tests/api_test.go`):
```go
package tests

import (
    "context"
    "testing"
    "github.com/therealutkarshpriyadarshi/containr/pkg/testframework"
)

func TestAPIEndpoint(t *testing.T) {
    runner := testframework.NewBehaviorRunner()

    test := &testframework.BehaviorTest{
        Description: "API endpoint returns user data",
        Given: func(ctx context.Context) error {
            // Setup: Start database and API containers
            return startTestEnvironment()
        },
        When: func(ctx context.Context) error {
            // Action: Make API request
            return makeAPIRequest("/api/users/1")
        },
        Then: func(ctx context.Context) error {
            // Assert: Verify response
            return verifyUserData()
        },
    }

    if err := runner.Run(context.Background(), test); err != nil {
        t.Fatalf("Test failed: %v", err)
    }
}
```

### Integration Testing

Create integration test (`tests/integration_test.go`):
```go
func TestMicroservicesIntegration(t *testing.T) {
    runner := testframework.NewIntegrationRunner()

    test := &testframework.IntegrationTest{
        Name: "microservices-integration",
        Containers: []*testframework.TestContainer{
            {Name: "api", Image: "myapp/api:latest"},
            {Name: "db", Image: "postgres:15"},
            {Name: "cache", Image: "redis:7"},
        },
        Setup: func(ctx context.Context) error {
            return setupDatabase()
        },
        Test: func(ctx context.Context) error {
            return testAPIWithDatabase()
        },
        Teardown: func(ctx context.Context) error {
            return cleanupTestData()
        },
    }

    if err := runner.Run(context.Background(), test); err != nil {
        t.Fatalf("Integration test failed: %v", err)
    }
}
```

## Complete Development Workflow

Here's a complete workflow using all Phase 8 features:

```bash
# 1. Start IDE LSP server
containr ide lsp --port 4389 &

# 2. Create project with hot reload
containr dev myapp \
    --watch ./src:/app/src \
    --reload-strategy signal \
    --debug-port 9229

# 3. Edit code (auto-reload happens)
vim src/server.js

# 4. Debug if needed
containr debug attach myapp
> breakpoint add /app/src/server.js:42
> continue

# 5. Run tests
containr test run ./tests/

# 6. Scan for vulnerabilities
containr scan myapp:dev --min-severity high

# 7. Generate SBOM
containr sbom generate myapp:dev --format spdx

# 8. Check compliance
containr compliance check myapp:dev --policy policy.yaml

# 9. Commit and push (GitOps deploys automatically)
git add .
git commit -m "Add feature"
git push

# 10. Monitor deployment
containr gitops status myapp
```

## Best Practices

1. **IDE Integration**:
   - Use LSP for real-time validation
   - Configure editor snippets for common patterns
   - Enable diagnostics for immediate feedback

2. **Debugging**:
   - Use breakpoints sparingly
   - Prefer syscall tracing for performance debugging
   - Profile regularly to catch performance regressions

3. **Security**:
   - Generate SBOM for all images
   - Scan images before deployment
   - Set up compliance policies
   - Track CVEs and remediate quickly

4. **GitOps**:
   - Keep manifests in version control
   - Use separate branches for environments
   - Enable automatic synchronization
   - Monitor deployment status

5. **Hot Reload**:
   - Use appropriate reload strategy for your stack
   - Configure ignore patterns to avoid unnecessary reloads
   - Set appropriate debounce times
   - Use development-specific configurations

6. **Testing**:
   - Write tests for all containers
   - Use BDD for complex scenarios
   - Run integration tests in CI/CD
   - Generate test reports

## Troubleshooting

### LSP Not Working
```bash
# Check LSP server is running
ps aux | grep containr-lsp

# Restart LSP server
killall containr-lsp
containr ide lsp --port 4389
```

### Debugger Can't Attach
```bash
# Check container is running
containr ps

# Verify container PID
containr inspect myapp | grep Pid

# Try with sudo
sudo containr debug attach myapp
```

### Hot Reload Not Triggering
```bash
# Check watcher status
containr dev stats myapp-dev

# Verify file permissions
ls -la src/

# Check ignore patterns
cat .containr/dev.yaml
```

## Conclusion

You've learned how to use all Phase 8 features for a modern container development workflow. These tools significantly improve developer productivity and code quality.

## Next Steps

- Explore advanced LSP features
- Set up continuous security scanning
- Implement GitOps for all deployments
- Create custom test assertions
- Build development environment templates

For more information, see:
- [IDE Integration Guide](../PHASE8.md#81-ide-integration--lsp-support)
- [Advanced Debugging Guide](../PHASE8.md#82-advanced-debugging--profiling)
- [SBOM & Security Guide](../PHASE8.md#83-sbom-generation--security-scanning)
- [GitOps Guide](../PHASE8.md#84-gitops--cicd-integration)
- [Hot Reload Guide](../PHASE8.md#85-hot-reload--development-workflows)
- [Testing Guide](../PHASE8.md#86-container-testing-framework)
