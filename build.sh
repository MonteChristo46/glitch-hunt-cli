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
echo "Generating install scripts..."

BANNER_B64=$(base64 < assets/banner.txt | tr -d '\n')
VERSION_B64=$(echo -n "$VERSION" | base64 | tr -d '\n')

python3 -c "
import base64

banner = base64.b64decode('$BANNER_B64').decode()
version = base64.b64decode('$VERSION_B64').decode()

with open('scripts/install.sh.tpl') as f:
    content = f.read()
content = content.replace('{{VERSION}}', version)
content = content.replace('{{BANNER}}', banner)
with open('scripts/install.sh', 'w') as f:
    f.write(content)
"

chmod +x "scripts/install.sh"

python3 -c "
import base64

banner = base64.b64decode('$BANNER_B64').decode()
version = base64.b64decode('$VERSION_B64').decode()

with open('scripts/install.ps1.tpl') as f:
    content = f.read()
content = content.replace('{{VERSION}}', version)
content = content.replace('{{BANNER}}', banner)
with open('scripts/install.ps1', 'w') as f:
    f.write(content)
"

echo ""
echo "Done. Builds in ./build/"
ls -lh build/
