# Versioning System

## Overview

envswitch uses **Git tags** as the source of truth for version information. The version is automatically injected into the binary at build time using Go's linker flags (`-ldflags`).

## How It Works

### Version Information

The binary contains three pieces of version information:

1. **Version**: Git tag (e.g., `v1.0.0` or `v0.1.0-alpha.2-13-g484d911-dirty`)
2. **Commit**: Short git commit hash (e.g., `484d911`)
3. **Build Date**: UTC timestamp (e.g., `2025-10-06_14:58:10`)

### Implementation

In [cmd/root.go](cmd/root.go):

```go
var (
    // Version information - set via ldflags during build
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)
```

These variables are set at **build time** (not compile time) using:

```bash
go build -ldflags="-X github.com/hugofrely/envswitch/cmd.Version=v1.0.0 ..."
```

## Creating a Release

### 1. Create a Git Tag

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 2. GitHub Actions Automatically

When you push a tag starting with `v*`, GitHub Actions will:

1. Run all tests
2. Build binaries for:
   - Linux (AMD64, ARM64)
   - macOS (Intel, Apple Silicon)
   - Windows (AMD64)
3. Generate checksums
4. Create a GitHub Release
5. Upload all binaries as release artifacts

### 3. Download Binaries

Users can download pre-built binaries from:

```
https://github.com/hugofrely/envswitch/releases/latest
```

## Building Locally

### Using Make

```bash
# Build for current platform with version info
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux
make build-darwin
make build-windows
```

### Using Go Directly

```bash
# Get version info from git
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build with version info
go build -ldflags="-X github.com/hugofrely/envswitch/cmd.Version=${VERSION} \
  -X github.com/hugofrely/envswitch/cmd.GitCommit=${COMMIT} \
  -X github.com/hugofrely/envswitch/cmd.BuildDate=${DATE}" \
  -o envswitch
```

### Using the Build Script

```bash
# Simple build
./scripts/build.sh

# Production build (smaller binary)
PRODUCTION=true ./scripts/build.sh

# Custom output directory
OUTPUT_DIR=./dist ./scripts/build.sh
```

## Version Format

### With Git Tags

When building from a tagged commit:

```
v1.0.0
```

When building after a tag (with additional commits):

```
v1.0.0-5-g123abc
```

- `v1.0.0`: Latest tag
- `5`: Number of commits since tag
- `g123abc`: Short commit hash

With uncommitted changes:

```
v1.0.0-5-g123abc-dirty
```

### Without Git Tags

When no tags exist:

```
123abc (commit hash only)
```

### Development Builds

When building without git:

```
dev
```

## Checking Version

### Command Line

```bash
# Full version info
envswitch --version
# Output: envswitch version v1.0.0 (commit: 484d911, built: 2025-10-06_14:58:10)

# Short version
envswitch version
```

### In Code

```go
import "github.com/hugofrely/envswitch/cmd"

fmt.Printf("Version: %s\n", cmd.Version)
fmt.Printf("Commit: %s\n", cmd.GitCommit)
fmt.Printf("Built: %s\n", cmd.BuildDate)
```

## Semantic Versioning

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backwards compatible)
- **PATCH** version for bug fixes (backwards compatible)

### Examples

```bash
# Major release (breaking changes)
git tag -a v2.0.0 -m "Major release with breaking changes"

# Minor release (new features)
git tag -a v1.1.0 -m "Add shell integration feature"

# Patch release (bug fixes)
git tag -a v1.0.1 -m "Fix environment variable escaping bug"

# Pre-release versions
git tag -a v1.0.0-alpha.1 -m "Alpha release for testing"
git tag -a v1.0.0-beta.1 -m "Beta release"
git tag -a v1.0.0-rc.1 -m "Release candidate"
```

## Release Workflow

### For Maintainers

1. **Update CHANGELOG.md** with changes
2. **Run tests** locally: `make test`
3. **Update version** in any docs if needed
4. **Commit changes**: `git commit -am "Prepare for v1.0.0"`
5. **Create tag**: `git tag -a v1.0.0 -m "Release v1.0.0"`
6. **Push changes**: `git push && git push origin v1.0.0`
7. **GitHub Actions** creates the release automatically
8. **Verify release** on GitHub

### For Users

```bash
# Download latest release
curl -LO https://github.com/hugofrely/envswitch/releases/latest/download/envswitch-darwin-arm64

# Make executable
chmod +x envswitch-darwin-arm64

# Move to PATH
sudo mv envswitch-darwin-arm64 /usr/local/bin/envswitch

# Verify installation
envswitch --version
```

## CI/CD Integration

The `.github/workflows/release.yml` workflow:

```yaml
on:
  push:
    tags:
      - "v*" # Triggers on any tag starting with 'v'
```

**Triggered by**:

- Pushing tags like `v1.0.0`, `v0.1.0-alpha.1`, etc.

**Not triggered by**:

- Regular commits
- Branch pushes
- Pull requests

## Troubleshooting

### "dev" version in binary

**Problem**: Binary shows `version dev (commit: unknown, built: unknown)`

**Solution**: You built without git or without using the Makefile. Use:

```bash
make build
```

### Version doesn't match tag

**Problem**: Tagged `v1.0.0` but binary shows `v1.0.0-1-g123abc-dirty`

**Reasons**:

1. **Uncommitted changes**: Commit or stash changes
2. **Additional commits after tag**: You're not on the tagged commit
3. **Tag not pushed**: Run `git push origin v1.0.0`

**Check**:

```bash
# See current position relative to tags
git describe --tags

# See all tags
git tag -l

# See commit of a tag
git show v1.0.0
```

### Release not created automatically

**Problem**: Pushed tag but no GitHub release

**Checks**:

1. Tag must start with `v` (e.g., `v1.0.0`, not `1.0.0`)
2. Check GitHub Actions tab for workflow runs
3. Check workflow has `contents: write` permission
4. Verify `.github/workflows/release.yml` exists

## Best Practices

1. **Always use annotated tags** for releases:

   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"  # Good
   git tag v1.0.0                          # Avoid (lightweight tag)
   ```

2. **Test before tagging**:

   ```bash
   make test
   make build
   ./bin/envswitch --version
   ```

3. **Use pre-release tags** for testing:

   ```bash
   git tag -a v1.0.0-rc.1 -m "Release candidate 1"
   ```

4. **Keep CHANGELOG.md updated** with each release

5. **Never delete published tags** - they're immutable in releases

6. **Use consistent tag format**: `vMAJOR.MINOR.PATCH[-PRERELEASE]`

## Resources

- [Semantic Versioning](https://semver.org/)
- [Git Tagging](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Go linker flags](https://pkg.go.dev/cmd/link)
