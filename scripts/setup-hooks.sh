#!/bin/bash
set -e

# Setup Git Hooks for EnvSwitch
# This script installs pre-commit hooks to ensure code quality

echo "🔧 Setting up Git hooks for EnvSwitch..."

# Create hooks directory if it doesn't exist
HOOKS_DIR=".git/hooks"
if [ ! -d "$HOOKS_DIR" ]; then
    echo "Error: Not a git repository"
    exit 1
fi

# Create pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# EnvSwitch Pre-commit Hook
# Runs before each commit to ensure code quality

set -e

echo "🔍 Running pre-commit checks..."

# 1. Format code
echo "📝 Formatting code..."
UNFORMATTED=$(gofmt -l . 2>&1)
if [ -n "$UNFORMATTED" ]; then
    echo "⚠️  The following files need formatting:"
    echo "$UNFORMATTED"
    echo ""
    echo "Formatting now..."
    gofmt -w .
    git add -u
    echo "✓ Code formatted and staged. Please commit again."
    exit 1
fi

# 2. Run go mod tidy
echo "📦 Tidying go modules..."
go mod tidy

# 3. Run go vet
echo "🔎 Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet failed. Please fix issues before committing."
    exit 1
fi

# 4. Run tests (short mode for speed)
echo "🧪 Running tests..."
if ! go test -short ./...; then
    echo "❌ Tests failed. Please fix before committing."
    exit 1
fi

# 5. Check build
echo "🏗️  Checking build..."
if ! go build ./...; then
    echo "❌ Build failed. Please fix before committing."
    exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0
EOF

# Make pre-commit executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully!"
echo ""
echo "📝 The following checks will run before each commit:"
echo "   - Code formatting (gofmt)"
echo "   - Go modules tidy"
echo "   - Go vet"
echo "   - Tests (short mode)"
echo "   - Build validation"
echo ""
echo "💡 To skip hooks for a commit (use sparingly):"
echo "   git commit --no-verify"
echo ""
echo "🚀 You're all set!"
