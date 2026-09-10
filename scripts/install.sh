#!/usr/bin/env sh
# Usage (1-liner):
#   curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
set -e

OWNER=${OWNER:-Hangell}
REPO=${REPO:-gommit}
API="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
UA="gommit-installer"

echo "Installing gommit..."

uname_s=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$uname_s" in
  linux)  OS=linux ;;
  darwin) OS=darwin ;;
  *) echo "Unsupported OS: $uname_s"; exit 1 ;;
esac

uname_m=$(uname -m)
case "$uname_m" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported arch: $uname_m"; exit 1 ;;
esac

echo "Detected: ${OS}_${ARCH}"

tmp="$(mktemp -d 2>/dev/null || mktemp -d -t gommit)"

# Fetch release info and parse JSON more reliably
echo "Fetching release information..."
release_data=$(curl -fsSL -H "User-Agent: $UA" "$API")

# Extract download URL using a more robust approach
url=$(echo "$release_data" | grep -o '"browser_download_url": *"[^"]*"' | \
      grep "gommit_.*_${OS}_${ARCH}\.tar\.gz" | \
      head -n1 | \
      sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')

if [ -z "$url" ]; then
  echo "No matching asset for ${OS}_${ARCH}"
  echo "Available assets:"
  echo "$release_data" | grep -o '"name": *"[^"]*\.tar\.gz"' | sed 's/.*"name": *"\([^"]*\)".*/  - \1/'
  echo "$release_data" | grep -o '"name": *"[^"]*\.zip"' | sed 's/.*"name": *"\([^"]*\)".*/  - \1/'
  exit 1
fi

echo "Downloading $(basename "$url")"
cd "$tmp"
curl -fsSL -o gommit.tar.gz "$url"

echo "Extracting..."
tar -xzf gommit.tar.gz

# Find the binary more reliably
bin=""
for candidate in $(find . -name gommit -type f 2>/dev/null); do
  if [ -x "$candidate" ] || [ -f "$candidate" ]; then
    bin="$candidate"
    break
  fi
done

if [ -z "$bin" ]; then
  echo "gommit binary not found in package"
  echo "Package contents:"
  find . -type f | head -20
  exit 1
fi

chmod +x "$bin"
echo "Installing..."
"$bin" --install

# Cleanup
rm -rf "$tmp"

echo ""
echo "✅ Installation completed successfully!"
echo "Restart your shell and run: gommit --version"