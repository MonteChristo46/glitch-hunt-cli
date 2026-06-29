#!/bin/sh

INSTALL_DIR="/opt/huntcli"
SYMLINK_PATH="/usr/local/bin/huntcli"

if [ "$(id -u)" -ne 0 ]; then
    echo "[SYSTEM] Not running as root."
    INSTALL_DIR="$HOME/.huntcli"
    SYMLINK_PATH="$HOME/.local/bin/huntcli"
fi

echo "[STATUS] Removing symlink: $SYMLINK_PATH"
rm -f "$SYMLINK_PATH"

echo "[STATUS] Removing installation directory: $INSTALL_DIR"
rm -rf "$INSTALL_DIR"

echo "[SUCCESS] huntcli uninstalled."
