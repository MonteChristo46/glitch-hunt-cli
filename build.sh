#!/bin/bash
set -euo pipefail

VERSION=$(cat assets/VERSION)
echo "Building huntcli v${VERSION}..."

PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

for plat in "${PLATFORMS[@]}"; do
  GOOS="${plat%/*}"
  GOARCH="${plat#*/}"
  
  ext=""
  if [ "$GOOS" = "windows" ]; then
    ext=".exe"
  fi
  
  output="build/huntcli-${GOOS}-${GOARCH}${ext}"
  
  echo "  -> ${output}"
  GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build -ldflags="-s -w" -o "$output" ./cmd/huntcli/
done

echo ""
echo "Done. Builds in ./build/"
ls -lh build/
