#!/bin/bash
set -e

VERSION="${PORT0_VERSION:-latest}"
INSTALL_DIR="${PORT0_INSTALL_DIR:-/usr/local/bin}"

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) echo "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) echo "unsupported" ;;
    esac
}

main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        echo "Error: Unsupported OS or architecture"
        exit 1
    fi

    echo "Installing port0 for $OS/$ARCH..."

    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/blu3ph4ntom/port0/releases/latest/download/port0-${OS}-${ARCH}"
    else
        DOWNLOAD_URL="https://github.com/blu3ph4ntom/port0/releases/download/${VERSION}/port0-${OS}-${ARCH}"
    fi

    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
        BINARY="port0.exe"
    else
        BINARY="port0"
    fi

    TEMP_FILE=$(mktemp)
    trap "rm -f $TEMP_FILE" EXIT

    echo "Downloading from $DOWNLOAD_URL..."
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "$TEMP_FILE"
    else
        echo "Error: curl or wget required"
        exit 1
    fi

    chmod +x "$TEMP_FILE"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY"
    else
        sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY"
    fi

    echo "✓ port0 installed to $INSTALL_DIR/$BINARY"
    echo ""
    echo "Quick start:"
    echo "  port0 run npm run dev"
    echo "  port0 run python -m http.server"
    echo "  port0 ls"
    echo ""
    echo "For *.web and *.local support, run:"
    echo "  sudo port0 setup"
}

main
