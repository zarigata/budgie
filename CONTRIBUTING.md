# Contributing to Budgie

Thank you for your interest in contributing to Budgie!

## 📋 How to Contribute

### Reporting Bugs

Before creating bug reports, please:
1. Search for existing issues
2. Include reproduction steps
3. Provide system information:
   - Budgie version: `budgie --version`
   - OS and version
   - Go version: `go version`
   - Architecture (amd64/arm64)

### Suggesting Enhancements

1. Check the [roadmap](GITHUB_README.md#-roadmap) first
2. Create a feature request with:
   - Clear description of the feature
   - Use cases and examples
   - How it would benefit the project

### Submitting Pull Requests

#### Development Workflow

1. **Fork the repository**
```bash
gh repo fork budgie/budgie
```

2. **Clone your fork**
```bash
git clone https://github.com/YOUR_USERNAME/budgie.git
cd budgie
git remote add upstream https://github.com/budgie/budgie.git
```

3. **Create a feature branch**
```bash
git checkout -b feature/your-feature-name
```

4. **Make your changes**
   - Write clean code following style guidelines
   - Add tests for new functionality
   - Update documentation

5. **Commit your changes**
```bash
git add .
git commit -m "Add: your feature description"
```

Commit message format:
- `Add: ` - Adding a new feature
- `Fix: ` - Bug fix
- `Update: ` - Update to dependencies/docs
- `Refactor: ` - Code restructuring without functionality change
- `Docs: ` - Documentation only
- `Test: ` - Adding or updating tests
- `Chore: ` - Maintenance tasks

6. **Push to your fork**
```bash
git push origin feature/your-feature-name
```

7. **Create a Pull Request**
   - Go to GitHub and create PR
   - Reference any related issues
   - Describe your changes clearly
   - Ensure tests pass

#### Code Style Guidelines

**Go Conventions:**
- Follow [Effective Go](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting: `make fmt`
- Handle errors properly, don't panic in production code
- Use meaningful variable names
- Add comments for exported types and complex logic
- Keep functions focused and small

**Testing:**
- Write unit tests for new functionality
- Aim for good test coverage
- Test edge cases
- Use table-driven tests where appropriate

**Documentation:**
- Update README for user-facing changes
- Add code comments for complex algorithms
- Document new configuration options

## 🏗️ Development Setup

### Prerequisites

- Go 1.21 or higher
- containerd (for local testing)
- Git

### Setting Up

```bash
# Clone the repository
git clone https://github.com/budgie/budgie.git
cd budgie

# Install dependencies
go mod download

# Run tests
make test
```

### Project Structure

```
budgie/
├── cmd/              # CLI commands (public API)
│   ├── nest/        # Setup wizard
│   ├── root/        # Root command (main entry point)
│   ├── run/         # Container operations
│   ├── ps/          # Container listing
│   ├── stop/        # Container control
│   └── chirp/       # Network discovery
├── internal/         # Private code (not exported)
│   ├── api/         # Container management
│   ├── bundle/      # .bun file parsing
│   ├── discovery/  # mDNS implementation
│   ├── runtime/      # containerd integration
│   ├── sync/         # File synchronization
│   └── proxy/        # Load balancer
└── pkg/             # Public packages
    └── types/       # Shared types and utilities
```

### Testing

Run all tests:
```bash
make test
```

Run specific test:
```bash
go test ./cmd/nest/...
go test ./internal/...
```

Run with coverage:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

Build for current platform:
```bash
make build
```

Build for all platforms:
```bash
./build-all.sh
```

## 🐛️ Troubleshooting Common Development Issues

### Dependency Issues

**Problem:** `go mod tidy` fails or dependencies don't download

**Solution:**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod tidy
go mod download
```

### Build Errors

**Problem:** Build fails on specific platform

**Solution:**
```bash
# Check Go version
go version

# Update Go if needed
# Install latest from https://go.dev/dl/

# Clean and rebuild
make clean
make build
```

### Containerd Issues

**Problem:** Cannot connect to containerd

**Solution:**
```bash
# Check if containerd is running
sudo systemctl status containerd  # Linux

# Check socket path
ls -la /run/containerd/containerd.sock

# Start containerd if needed
sudo systemctl start containerd  # Linux
```

## 📝 Project Resources

- [Documentation](GITHUB_README.md)
- [Roadmap](GITHUB_README.md#-roadmap)
- [Issue Tracker](https://github.com/budgie/budgie/issues)
- [Discussions](https://github.com/budgie/budgie/discussions)

## 🎓 Learning Resources

If you're new to the project:
1. Read the [README.md](README.md) for an overview
2. Run `budgie nest` to learn the system
3. Check out the `.bun` file format examples
4. Review existing issues and PRs

## 💬 Community

- Join our [Discord](https://discord.gg/budgie) for discussions
- Follow us on [Twitter/X](https://twitter.com/budgie)
- Subscribe to our [YouTube channel](https://youtube.com/@budgie)

## 📜 License

By contributing to Budgie, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to Budgie! 🐦
