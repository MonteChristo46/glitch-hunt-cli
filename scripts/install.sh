#!/bin/sh
set -e

VERSION="0.1.0-alpha"

printf "\033[38;2;156;39;176m██╗  ██╗██╗   ██╗███╗   ██╗████████╗     ██████╗██╗     ██╗\033[0m\n"
printf "\033[38;2;125;61;168m██║  ██║██║   ██║████╗  ██║╚══██╔══╝    ██╔════╝██║     ██║\033[0m\n"
printf "\033[38;2;94;83;160m███████║██║   ██║██╔██╗ ██║   ██║       ██║     ██║     ██║\033[0m\n"
printf "\033[38;2;63;105;152m██╔══██║██║   ██║██║╚██╗██║   ██║       ██║     ██║     ██║\033[0m\n"
printf "\033[38;2;32;127;144m██║  ██║╚██████╔╝██║ ╚████║   ██║       ╚██████╗███████╗██║\033[0m\n"
printf "\033[38;2;0;150;136m╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝        ╚═════╝╚══════╝╚═╝\033[0m\n"
printf " \033[38;2;200;200;200mCLI INSTALLER | v%s\033[0m\n\n" "$VERSION"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "[ERROR] Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Default: user install (no sudo)
INSTALL_DIR="$HOME/.local/bin"
BIN_NAME="huntcli"

# If running as root, use system-wide path
if [ "$(id -u)" -eq 0 ]; then
  INSTALL_DIR="/usr/local/bin"
fi

mkdir -p "$INSTALL_DIR"
TARGET="$INSTALL_DIR/$BIN_NAME"
DOWNLOAD_URL="https://raw.githubusercontent.com/MonteChristo46/glitch-hunt-cli/main/dist/${BIN_NAME}-${OS}-${ARCH}"

echo "[INFO] Target: $TARGET"

# Try to download pre-built binary from GitHub releases
if command -v curl >/dev/null 2>&1; then
  HTTP_CODE=$(curl -sfL -w "%{http_code}" "$DOWNLOAD_URL" -o "$TARGET" 2>/dev/null || echo "000")
  if [ "$HTTP_CODE" = "200" ]; then
    echo "[OK] Downloaded binary for $OS/$ARCH."
    chmod +x "$TARGET"
  else
    echo "[INFO] No pre-built binary available (HTTP $HTTP_CODE). Building from source..."
    curl=""
  fi
elif command -v wget >/dev/null 2>&1; then
  if wget -q "$DOWNLOAD_URL" -O "$TARGET" 2>/dev/null; then
    echo "[OK] Downloaded binary for $OS/$ARCH."
    chmod +x "$TARGET"
  else
    echo "[INFO] No pre-built binary available. Building from source..."
    wget=""
  fi
fi

# If download failed, build from source
if [ ! -f "$TARGET" ] || [ ! -s "$TARGET" ]; then
  if ! command -v go >/dev/null 2>&1; then
    echo "[ERROR] No pre-built binary found and Go is not installed."
    echo ""
    echo "To install Go:"
    echo "  macOS: brew install go"
    echo "  Linux: see https://go.dev/doc/install"
    echo ""
    echo "Then re-run this script, or download manually from:"
    echo "  https://github.com/MonteChristo46/glitch-hunt-cli"
    exit 1
  fi

  TMP_DIR=$(mktemp -d)
  echo "[INFO] Cloning repository..."
  git clone --depth 1 https://github.com/MonteChristo46/glitch-hunt-cli.git "$TMP_DIR" 2>/dev/null || {
    echo "[ERROR] Failed to clone repository."
    rm -rf "$TMP_DIR"
    exit 1
  }

  echo "[INFO] Building huntcli..."
  cd "$TMP_DIR"
  go build -o "$TARGET" ./cmd/huntcli/ 2>/dev/null || {
    echo "[ERROR] Build failed."
    rm -rf "$TMP_DIR"
    exit 1
  }
  rm -rf "$TMP_DIR"
  echo "[OK] Built from source."
fi

chmod +x "$TARGET"

if [ "$(uname -s)" = "Darwin" ]; then
  echo "[INFO] Applying macOS security fix..."
  xattr -d com.apple.quarantine "$TARGET" 2>/dev/null || true
  codesign --force --deep -s - "$TARGET" >/dev/null 2>&1 || true
fi

echo ""
echo "[OK] Installed to: $TARGET"
echo ""

# Check PATH and ask to add it
IN_PATH=false
case ":$PATH:" in
  *:"$INSTALL_DIR":*) IN_PATH=true ;;
esac

if [ "$IN_PATH" = false ]; then
  echo "Note: $INSTALL_DIR is not in your PATH."
  printf "Add it to your PATH now? [Y/n]: "
  read -r ADD_PATH < /dev/tty 2>/dev/null || read -r ADD_PATH
  case "$ADD_PATH" in
    n|N|no|NO)
      echo "Skipping PATH update."
      echo ""
      echo "To add it manually, run:"
      echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
      echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.$(basename "${SHELL:-sh}")rc"
      echo ""
      ;;
    *)
      SHELL_NAME="$(basename "${SHELL:-sh}")"
      case "$SHELL_NAME" in
        zsh) RC_FILE="$HOME/.zshrc" ;;
        bash)
          if [ -f "$HOME/.bash_profile" ]; then
            RC_FILE="$HOME/.bash_profile"
          elif [ -f "$HOME/.bashrc" ]; then
            RC_FILE="$HOME/.bashrc"
          else
            RC_FILE="$HOME/.profile"
          fi
          ;;
        fish) RC_FILE="$HOME/.config/fish/config.fish" ;;
        *) RC_FILE="$HOME/.profile" ;;
      esac

      if grep -q "export PATH=.*$INSTALL_DIR.*" "$RC_FILE" 2>/dev/null; then
        echo "[OK] $INSTALL_DIR already configured in $RC_FILE."
      else
        echo "" >> "$RC_FILE"
        echo "# Added by huntcli installer" >> "$RC_FILE"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$RC_FILE"
        echo "[OK] Added to PATH in $RC_FILE."
        echo "     Restart your terminal or run: source $RC_FILE"
      fi
      echo ""
      ;;
  esac
fi

echo "Now run 'huntcli install' to complete setup:"
echo "  $TARGET install"
echo ""
echo "Or authenticate directly:"
echo "  $TARGET login"
echo "  $TARGET listen --forward-to http://localhost:8080/webhooks"
