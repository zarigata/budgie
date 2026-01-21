.PHONY: build clean test install run dev build-all nest

# Build budgie binary for current platform
build:
	@echo "🔨 Building budgie for current platform..."
	@mkdir -p bin
	@if [ "$(GOOS)" = "windows" ]; then \
		go build -o bin/budgie.exe ./cmd/root/main.go; \
	else \
		go build -o bin/budgie ./cmd/root/main.go; \
	fi
	@echo "✅ Build complete: bin/budgie$(if [ "$(GOOS)" = "windows" ]; then echo .exe; fi)"

# Install budgie to GOBIN or GOPATH/bin
install:
	@echo "📦 Installing budgie..."
	@go install -o $(shell go env GOPATH)/bin/budgie ./cmd/root/main.go
	@echo "✅ Installation complete: budgie is now in your PATH"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "✅ Clean complete"

# Run budgie directly
run:
	go run ./cmd/root/main.go $(ARGS)

# Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "✅ All builds complete"
	@echo ""
	@echo "Binaries available in bin/:"
	@ls -lh bin/

# Build for Linux
build-linux:
	@echo "🔨 Building for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/budgie-linux-amd64 ./cmd/root/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/budgie-linux-arm64 ./cmd/root/main.go
	@echo "✅ Linux builds complete"

# Build for macOS (Darwin)
build-darwin:
	@echo "🔨 Building for macOS..."
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/budgie-darwin-amd64 ./cmd/root/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/budgie-darwin-arm64 ./cmd/root/main.go
	@echo "✅ macOS builds complete"

# Build for Windows
build-windows:
	@echo "🔨 Building for Windows..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/budgie-windows-amd64.exe ./cmd/root/main.go
	GOOS=windows GOARCH=arm64 go build -o bin/budgie-windows-arm64.exe ./cmd/root/main.go
	@echo "✅ Windows builds complete"

# Development mode with hot reload (requires air)
dev:
	air

# Format code
fmt:
	@echo "📐 Formatting code..."
	go fmt ./...
	@echo "✅ Format complete"

# Run linter
lint:
	@echo "🔍 Running linter..."
	golangci-lint run

# Run go mod tidy
mod-tidy:
	@echo "📦 Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies tidy"

# Vendor dependencies
vendor:
	go mod vendor

# Create release package
release: build-all
	@echo "📦 Creating release packages..."
	@mkdir -p release
	@cd bin && for f in budgie-*; do tar -czf ../release/$$f.tar.gz $$f; done
	@echo "✅ Release packages created in release/"

# Install nest command (auto-builder)
nest:
	@echo "🏗️  Building budgie with nest command..."
	@go build -o bin/budgie ./cmd/root/main.go
	@echo "✅ Budgie with nest command is ready"
	@echo "   Run: ./bin/budgie nest"

