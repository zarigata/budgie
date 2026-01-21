#!/bin/bash
# Build script for budgie - requires Go to be installed
# Run this script on a system with Go to build all platform binaries

set -e

VERSION="0.1"
BINARY_NAME="budgie"
OUTPUT_DIR="bin"
RELEASE_DIR="release"

echo "🐦 Building Budgie v${VERSION}"
echo "================================"
echo ""

# Create output directories
mkdir -p "${OUTPUT_DIR}"
mkdir -p "${RELEASE_DIR}"

echo "🔨 Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64" ./cmd/root/main.go
echo "✅ Linux AMD64: ${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64"

echo "🔨 Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64" ./cmd/root/main.go
echo "✅ Linux ARM64: ${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64"

echo "🔨 Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64" ./cmd/root/main.go
echo "✅ macOS AMD64: ${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64"

echo "🔨 Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64" ./cmd/root/main.go
echo "✅ macOS ARM64: ${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64"

echo "🔨 Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-windows-amd64.exe" ./cmd/root/main.go
echo "✅ Windows AMD64: ${OUTPUT_DIR}/${BINARY_NAME}-windows-amd64.exe"

echo "🔨 Building for Windows ARM64..."
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${BINARY_NAME}-windows-arm64.exe" ./cmd/root/main.go
echo "✅ Windows ARM64: ${OUTPUT_DIR}/${BINARY_NAME}-windows-arm64.exe"

echo ""
echo "📦 Creating release packages..."
cd "${OUTPUT_DIR}"

# Linux packages
tar -czf "../${RELEASE_DIR}/${BINARY_NAME}-linux-amd64.tar.gz" "${BINARY_NAME}-linux-amd64"
tar -czf "../${RELEASE_DIR}/${BINARY_NAME}-linux-arm64.tar.gz" "${BINARY_NAME}-linux-arm64"

# macOS packages
tar -czf "../${RELEASE_DIR}/${BINARY_NAME}-darwin-amd64.tar.gz" "${BINARY_NAME}-darwin-amd64"
tar -czf "../${RELEASE_DIR}/${BINARY_NAME}-darwin-arm64.tar.gz" "${BINARY_NAME}-darwin-arm64"

# Windows packages
zip "../${RELEASE_DIR}/${BINARY_NAME}-windows-amd64.zip" "${BINARY_NAME}-windows-amd64.exe"
zip "../${RELEASE_DIR}/${BINARY_NAME}-windows-arm64.zip" "${BINARY_NAME}-windows-arm64.exe"

cd ..

echo ""
echo "✅ Build complete!"
echo ""
echo "📁 Binaries:"
ls -lh "${OUTPUT_DIR}/"
echo ""
echo "📦 Release packages:"
ls -lh "${RELEASE_DIR}/"
echo ""
echo "🚀 You can now distribute these binaries or install them:"
echo "   - Linux/macOS: chmod +x budgie-* && sudo mv budgie-* /usr/local/bin/budgie"
echo "   - Windows: Add budgie-*.exe to your PATH"
