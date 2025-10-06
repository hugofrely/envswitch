# Development Guide

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - Version control
- **Make** - Build automation (usually pre-installed on macOS/Linux)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/hugofrely/envswitch.git
cd envswitch

# Install dependencies
make deps

# Install git hooks (recommended)
make setup-hooks

# Build the project
make build

# Run tests
make test
```

## 🔧 Git Hooks

### Installation

Git hooks are automatically installed with:

```bash
make setup-hooks
```

Or manually:

```bash
./scripts/setup-hooks.sh
```

### What Hooks Do

The pre-commit hook runs before each commit:

1. ✅ **Code Formatting** - Runs `gofmt` on all Go files
2. ✅ **Go Modules** - Runs `go mod tidy`
3. ✅ **Go Vet** - Checks for common issues
4. ✅ **Tests** - Runs tests in short mode
5. ✅ **Build** - Verifies the code builds

### Skipping Hooks

⚠️ **Use sparingly** - Only when you're sure:

```bash
git commit --no-verify -m "WIP: work in progress"
```

### Alternative: pre-commit Framework

If you prefer the `pre-commit` framework:

```bash
# Install pre-commit (macOS)
brew install pre-commit

# Install pre-commit (Python)
pip install pre-commit

# Setup hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## 🛠️ Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
```

### 2. Make Changes

```bash
# Edit files
vim cmd/mycommand.go

# Format code
make fmt

# Run tests
make test
```

### 3. Verify Changes

```bash
# Quick checks (no linter required)
make check         # Format, vet, build, test

# Full CI checks (requires golangci-lint)
make ci            # Format, vet, lint, test-race

# Or step by step:
make fmt           # Format code
make vet           # Go vet
make lint          # Run linter
make test-race     # Tests with race detector
make build         # Build
```

### 4. Commit

```bash
# Hooks run automatically
git add .
git commit -m "feat: add new feature"

# Hooks will run:
# - Format check
# - Go vet
# - Tests
# - Build
```

### 5. Push

```bash
git push origin feature/my-feature
```

## 🧪 Testing

### Run Tests

```bash
# All tests
make test

# With race detector (like CI)
make test-race

# With coverage
make test-coverage

# Short tests (fast, for pre-commit)
go test -short ./...

# Specific package
go test ./pkg/tools -v

# Specific test
go test ./pkg/tools -run TestGCloudTool
```

### Writing Tests

Tests should follow these patterns:

```go
func TestMyFunction(t *testing.T) {
    t.Run("does something", func(t *testing.T) {
        // Arrange
        input := "test"

        // Act
        result := MyFunction(input)

        // Assert
        assert.Equal(t, "expected", result)
    })
}
```

## 🔍 Code Quality

### Linting

```bash
# Install golangci-lint
make install-lint           # Auto-install (recommended)
# or
brew install golangci-lint  # macOS via Homebrew
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint                   # Check for lint issues

# Auto-fix issues
make lint-fix               # Fix issues automatically when possible
```

### Formatting

```bash
# Format all code
make fmt

# Check formatting
gofmt -l .
```

### Vetting

```bash
# Run go vet
make vet
```

## 📦 Building

### Local Build

```bash
# Build binary
make build

# Run binary
./bin/envswitch --help
```

### Cross-Platform

```bash
# Build for all platforms
make build-all

# Specific platforms
make build-linux
make build-darwin
make build-windows
```

## 🎯 Makefile Targets

```bash
make help              # Show all targets
make build             # Build binary
make test              # Run tests
make test-race         # Tests with race detector
make test-coverage     # Coverage report
make lint              # Run golangci-lint
make lint-fix          # Run golangci-lint with auto-fix
make install-lint      # Install golangci-lint
make fmt               # Format code
make vet               # Go vet
make check             # Quick checks (fmt, vet, build, test)
make ci                # All CI checks (fmt, vet, lint, test-race)
make setup-hooks       # Install git hooks
make clean             # Clean build artifacts
make deps              # Download dependencies
```

## 🐛 Debugging

### Debug Tests

```bash
# Verbose output
go test -v ./pkg/tools

# With prints
go test -v ./pkg/tools | grep "MyTest"

# Debug specific test
dlv test ./pkg/tools -- -test.run TestMyFunction
```

### Debug Binary

```bash
# Build with debug info
go build -gcflags="all=-N -l" -o bin/envswitch

# Run with delve
dlv exec ./bin/envswitch -- init
```

## 📝 Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Tests
- `refactor` - Code refactoring
- `perf` - Performance
- `chore` - Maintenance

**Examples:**

```bash
feat(snapshot): add GCloud snapshot capture
fix(cmd): handle empty environment list correctly
docs: update README with installation steps
test(tools): add tests for AWS tool
```

## 🚨 Troubleshooting

### Tests Failing

```bash
# Clean and retry
make clean
make deps
make test
```

### Build Failing

```bash
# Check Go version
go version  # Should be 1.21+

# Update dependencies
go mod tidy
go mod download
```

### Hooks Not Running

```bash
# Reinstall hooks
make setup-hooks

# Check if installed
ls -la .git/hooks/pre-commit

# Make executable
chmod +x .git/hooks/pre-commit
```

### CI Failing

```bash
# Run CI checks locally
make ci

# Check specific issue
make fmt      # Formatting
make vet      # Go vet
make test-race # Race conditions
```

## 📚 Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Cobra Documentation](https://cobra.dev/)
- [Testing in Go](https://golang.org/pkg/testing/)

## 🤝 Getting Help

- 📖 Read [CONTRIBUTING.md](CONTRIBUTING.md)
- 💬 Open a [Discussion](https://github.com/hugofrely/envswitch/discussions)
- 🐛 Report [Issues](https://github.com/hugofrely/envswitch/issues)

---

**Happy coding! 🎉**
