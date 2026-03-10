#!/bin/sh
set -eu

REPO="santif/openspec-go"
BINARY="openspec"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/${REPO}/releases"

say() {
    printf '[openspec] %s\n' "$@" >&2
}

err() {
    say "ERROR: $1"
    exit 1
}

main() {
    if ! command -v curl >/dev/null 2>&1; then
        err "curl is required but not found. Please install curl and try again."
    fi

    # Detect OS
    OS=$(uname -s)
    case "$OS" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      err "Unsupported operating system: $OS (supported: Linux, macOS)" ;;
    esac

    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              err "Unsupported architecture: $ARCH (supported: amd64, arm64)" ;;
    esac

    # Get version (allow override via VERSION env var)
    if [ -n "${VERSION:-}" ]; then
        TAG="v${VERSION#v}"
        VERSION="${TAG#v}"
        say "Using specified version: ${VERSION}"
    else
        say "Fetching latest release..."
        TAG=$(curl -fsSL "${GITHUB_API}/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' \
            | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/')

        if [ -z "$TAG" ]; then
            err "Failed to determine latest release version"
        fi

        VERSION="${TAG#v}"
        say "Latest version: ${VERSION}"
    fi

    # Build download URLs
    ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="${GITHUB_RELEASES}/download/${TAG}/${ARCHIVE}"
    CHECKSUMS_URL="${GITHUB_RELEASES}/download/${TAG}/checksums.txt"

    # Create temp directory with cleanup trap
    WORK_DIR=$(mktemp -d)
    trap 'rm -rf "$WORK_DIR"' EXIT INT TERM

    # Download archive and checksums
    say "Downloading ${ARCHIVE}..."
    if ! curl -fsSL -o "${WORK_DIR}/${ARCHIVE}" "$DOWNLOAD_URL"; then
        err "Failed to download ${ARCHIVE}. Check that version ${VERSION} exists for ${OS}/${ARCH}."
    fi

    say "Verifying checksum..."
    if ! curl -fsSL -o "${WORK_DIR}/checksums.txt" "$CHECKSUMS_URL"; then
        err "Failed to download checksums.txt"
    fi

    # Determine SHA256 tool
    if command -v sha256sum >/dev/null 2>&1; then
        SHASUM_CMD="sha256sum"
    elif command -v shasum >/dev/null 2>&1; then
        SHASUM_CMD="shasum -a 256"
    else
        err "No SHA256 tool found (need sha256sum or shasum)"
    fi

    # Verify checksum
    EXPECTED=$(grep -F "${ARCHIVE}" "${WORK_DIR}/checksums.txt" | awk '{print $1}')
    if [ -z "$EXPECTED" ]; then
        err "Archive ${ARCHIVE} not found in checksums.txt"
    fi

    ACTUAL=$(cd "$WORK_DIR" && $SHASUM_CMD "${ARCHIVE}" | awk '{print $1}')
    if [ "$EXPECTED" != "$ACTUAL" ]; then
        err "Checksum verification failed (expected: ${EXPECTED}, got: ${ACTUAL})"
    fi

    say "Checksum verified."

    # Extract binary
    tar -xzf "${WORK_DIR}/${ARCHIVE}" -C "${WORK_DIR}" "${BINARY}"

    # Determine install directory
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=""

    if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
        USE_SUDO=""
    elif command -v sudo >/dev/null 2>&1; then
        USE_SUDO="sudo"
        say "Installing to ${INSTALL_DIR} (requires sudo)..."
    else
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    # Install binary
    if [ -n "$USE_SUDO" ]; then
        sudo install -m 755 "${WORK_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        install -m 755 "${WORK_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    # Warn if install dir is not on PATH
    case ":${PATH}:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            say ""
            say "Add ${INSTALL_DIR} to your PATH:"
            say "  export PATH=\"${INSTALL_DIR}:\$PATH\""
            say ""
            ;;
    esac

    say "Installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"
}

main "$@"
