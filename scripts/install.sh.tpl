#!/bin/sh
set -e

VERSION="{{VERSION}}"

printf "{{BANNER}}"
printf " \033[38;2;200;200;200mCLI INSTALLER | v%s\033[0m\n\n" "$VERSION"

INSTALL_DIR="/opt/huntcli"
BIN_NAME="huntcli"
SYMLINK_PATH="/usr/local/bin/huntcli"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "[SYSTEM] Unsupported architecture: $ARCH"
    exit 1
fi

DOWNLOAD_URL="https://github.com/MonteChristo46/glitch-hunt-cli/raw/main/build/huntcli-${OS}-${ARCH}"

echo "[SYSTEM] Detected: $OS / $ARCH"

if [ "$(id -u)" -ne 0 ]; then
    echo "[SYSTEM] Not running as root. Installing to user directory."
    INSTALL_DIR="$HOME/.huntcli"
    SYMLINK_PATH="$HOME/.local/bin/huntcli"
    mkdir -p "$HOME/.local/bin"
fi

echo "[CONFIG] Target Directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

TARGET="${INSTALL_DIR}/${BIN_NAME}"

echo "[STATUS] Downloading huntcli..."
curl -fsSL "$DOWNLOAD_URL" -o "$TARGET"

if [ ! -f "$TARGET" ]; then
    echo "[ERROR] Download failed."
    exit 1
fi

chmod +x "$TARGET"

if [ "$(uname -s)" = "Darwin" ]; then
    echo "[STATUS] Applying macOS security fix (ad-hoc signing)..."
    xattr -d com.apple.quarantine "$TARGET" 2>/dev/null || true
    codesign --force --deep -s - "$TARGET" >/dev/null 2>&1
fi

echo "[CONFIG] Creating symlink at $SYMLINK_PATH..."
mkdir -p "$(dirname "$SYMLINK_PATH")"
rm -f "$SYMLINK_PATH"
ln -s "$TARGET" "$SYMLINK_PATH"

echo "[STATUS] Running huntcli install..."
echo "--------------------------------------------------"
if [ -t 0 ]; then
    "$TARGET" install
else
    if [ -c /dev/tty ]; then
        "$TARGET" install < /dev/tty
    else
        echo "[WARN] No TTY detected. Running non-interactive."
        "$TARGET" install --skip-login
    fi
fi

echo "--------------------------------------------------"
echo "[SUCCESS] Installation complete. You can now use 'huntcli'."
