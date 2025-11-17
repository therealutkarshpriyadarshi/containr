#!/bin/bash
# Phase 6 Feature Demonstration Script
# This script demonstrates the new Phase 6 features of containr

set -e

echo "========================================"
echo "Containr Phase 6 Feature Demonstration"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. CRI Commands
echo -e "${BLUE}1. CRI (Container Runtime Interface) Commands${NC}"
echo "---------------------------------------------"
echo ""

echo "Show CRI version:"
sudo containr cri version
echo ""

echo "Check CRI status:"
sudo containr cri status
echo ""

# 2. Plugin Commands
echo -e "${BLUE}2. Plugin Management${NC}"
echo "---------------------"
echo ""

echo "List available plugins:"
sudo containr plugin ls
echo ""

echo "Show plugin info (if any plugins are installed):"
# sudo containr plugin info prometheus-exporter
echo "(Example command: containr plugin info prometheus-exporter)"
echo ""

# 3. Snapshot Commands
echo -e "${BLUE}3. Snapshot Management${NC}"
echo "-----------------------"
echo ""

echo "List snapshots:"
sudo containr snapshot ls
echo ""

echo "Example snapshot commands:"
echo "  - Create snapshot:  containr snapshot create mycontainer snapshot1"
echo "  - Inspect snapshot: containr snapshot inspect snapshot1"
echo "  - Export snapshot:  containr snapshot export snapshot1 -o snapshot.tar.gz"
echo "  - Import snapshot:  containr snapshot import snapshot.tar.gz"
echo ""

# 4. Build Commands
echo -e "${BLUE}4. Image Build Engine${NC}"
echo "----------------------"
echo ""

echo "Example build commands:"
echo "  - Build image:      containr build -t myapp:latest ."
echo "  - Multi-stage:      containr build --target production -t myapp:prod ."
echo "  - With build args:  containr build --build-arg VERSION=1.0 -t myapp:v1 ."
echo "  - No cache:         containr build --no-cache -t myapp:latest ."
echo ""

# Show example Dockerfile
if [ -f "examples/Dockerfile.example" ]; then
    echo "Example Dockerfile available at: examples/Dockerfile.example"
fi
echo ""

echo -e "${GREEN}✅ Phase 6 demonstration complete!${NC}"
echo ""
echo "Key Phase 6 Features:"
echo "  ✅ CRI (Container Runtime Interface) for Kubernetes"
echo "  ✅ Plugin system for extensibility"
echo "  ✅ Snapshot support for fast container operations"
echo "  ✅ Complete build engine with Dockerfile support"
echo ""
echo "For more information, see docs/PHASE6.md"
