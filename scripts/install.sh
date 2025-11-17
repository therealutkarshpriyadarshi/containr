#!/bin/bash
# Containr Installation Script
# This script installs the latest version of containr

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="therealutkarshpriyadarshi/containr"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="containr"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv6l)
            ARCH="arm"
            ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    if [ "$OS" != "linux" ]; then
        echo -e "${RED}Error: This script only supports Linux${NC}"
        exit 1
    fi

    echo -e "${GREEN}Detected platform: $OS/$ARCH${NC}"
}

# Get latest release version
get_latest_version() {
    echo -e "${YELLOW}Fetching latest release...${NC}"
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        echo -e "${RED}Error: Failed to fetch latest version${NC}"
        exit 1
    fi

    echo -e "${GREEN}Latest version: $LATEST_VERSION${NC}"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$LATEST_VERSION/containr-$LATEST_VERSION-$OS-$ARCH.tar.gz"
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/containr.tar.gz"

    echo -e "${YELLOW}Downloading $BINARY_NAME from $DOWNLOAD_URL...${NC}"

    if ! curl -sL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
        echo -e "${RED}Error: Failed to download binary${NC}"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    echo -e "${GREEN}Download complete${NC}"

    # Extract archive
    echo -e "${YELLOW}Extracting archive...${NC}"
    tar -xzf "$TMP_FILE" -C "$TMP_DIR"

    BINARY_PATH="$TMP_DIR/$BINARY_NAME-$OS-$ARCH"

    if [ ! -f "$BINARY_PATH" ]; then
        echo -e "${RED}Error: Binary not found in archive${NC}"
        rm -rf "$TMP_DIR"
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    echo -e "${YELLOW}Verifying checksum...${NC}"

    CHECKSUMS_URL="https://github.com/$REPO/releases/download/v$LATEST_VERSION/checksums.txt"
    CHECKSUMS_FILE="$TMP_DIR/checksums.txt"

    if ! curl -sL "$CHECKSUMS_URL" -o "$CHECKSUMS_FILE"; then
        echo -e "${YELLOW}Warning: Could not download checksums file${NC}"
        return
    fi

    cd "$TMP_DIR"
    if sha256sum -c "$CHECKSUMS_FILE" --ignore-missing 2>/dev/null; then
        echo -e "${GREEN}Checksum verification passed${NC}"
    else
        echo -e "${YELLOW}Warning: Checksum verification failed${NC}"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$TMP_DIR"
            exit 1
        fi
    fi
    cd - > /dev/null
}

# Install binary
install_binary() {
    echo -e "${YELLOW}Installing $BINARY_NAME to $INSTALL_DIR...${NC}"

    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}Sudo access required to install to $INSTALL_DIR${NC}"
        SUDO="sudo"
    else
        SUDO=""
    fi

    # Make binary executable
    chmod +x "$BINARY_PATH"

    # Install binary
    if ! $SUDO install -m 755 "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"; then
        echo -e "${RED}Error: Failed to install binary${NC}"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Clean up
    rm -rf "$TMP_DIR"

    echo -e "${GREEN}Installation complete!${NC}"
}

# Verify installation
verify_installation() {
    echo -e "${YELLOW}Verifying installation...${NC}"

    if ! command -v "$BINARY_NAME" &> /dev/null; then
        echo -e "${RED}Error: $BINARY_NAME not found in PATH${NC}"
        echo -e "${YELLOW}You may need to add $INSTALL_DIR to your PATH${NC}"
        exit 1
    fi

    VERSION=$($BINARY_NAME version --short)
    echo -e "${GREEN}Successfully installed: $VERSION${NC}"
}

# Show usage instructions
show_usage() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Containr installed successfully!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Get started:"
    echo ""
    echo "  # View version"
    echo "  $ containr version"
    echo ""
    echo "  # Pull an image"
    echo "  $ sudo containr pull alpine"
    echo ""
    echo "  # Run a container"
    echo "  $ sudo containr run alpine /bin/sh"
    echo ""
    echo "  # See all commands"
    echo "  $ containr --help"
    echo ""
    echo "Documentation: https://github.com/$REPO/tree/main/docs"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Containr Installation Script           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""

    detect_platform
    get_latest_version
    download_binary
    verify_checksum
    install_binary
    verify_installation
    show_usage
}

# Run main function
main "$@"
