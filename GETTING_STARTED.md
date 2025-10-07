# Getting Started with EnvSwitch Development

Welcome to EnvSwitch! This guide will help you set up your development environment and start contributing to this production-ready CLI tool.

## Prerequisites

### Required
- **Go 1.21 or higher** - [Install Go](https://golang.org/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)

### Recommended
- **Make** - For using the Makefile (usually pre-installed on macOS/Linux)
- A code editor (VS Code, GoLand, etc.)

## Setup

### 1. Install Go

**macOS:**
```bash
brew install go
```

**Linux:**
```bash
# Download and install from golang.org
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Windows:**
Download and install from [golang.org](https://golang.org/dl/)

### 2. Clone the Repository

```bash
git clone https://github.com/hugofrely/envswitch.git
cd envswitch
```

### 3. Install Dependencies

```bash
make deps
# or
go mod download
```

### 4. Build the Project

```bash
make build
```

This creates the binary in `bin/envswitch`.

### 5. Test It Works

```bash
./bin/envswitch --version
./bin/envswitch --help
```

## Development Workflow

### Building

```bash
# Build the binary
make build

# Build and install to /usr/local/bin
make install

# Clean build artifacts
make clean
```

### Running

```bash
# Run directly without installing
make run

# Or run the binary
./bin/envswitch init
./bin/envswitch create test --empty
./bin/envswitch list
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./cmd -v

# Run with race detector
make test-race
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if installed)
golangci-lint run
```

## Project Structure Explained

```
envswitch/
│
├── cmd/                        # CLI commands
│   ├── root.go                # Root command and app initialization
│   ├── create.go              # envswitch create
│   ├── list.go                # envswitch list
│   ├── switch.go              # envswitch switch
│   └── ...
│
├── pkg/                        # Public, reusable packages
│   ├── environment/           # Environment management
│   │   └── environment.go     # Core environment logic
│   ├── tools/                 # Tool integrations (gcloud, kubectl, etc.)
│   │   ├── tool.go            # Tool interface
│   │   └── gcloud.go          # GCloud implementation
│   └── ...
│
├── internal/                   # Private packages (not importable)
│   ├── storage/               # File operations
│   ├── config/                # Configuration management
│   └── logger/                # Logging utilities
│
├── docs/                       # Documentation
├── .github/                    # GitHub workflows and templates
├── main.go                     # Application entry point
├── Makefile                    # Build automation
└── go.mod                      # Go module definition
```

## Making Your First Contribution

### 1. Pick an Issue

Look for issues labeled `good first issue` or `help wanted`:
- [Good First Issues](https://github.com/hugofrely/envswitch/labels/good%20first%20issue)
- [Help Wanted](https://github.com/hugofrely/envswitch/labels/help%20wanted)

### 2. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bugfix
```

### 3. Make Your Changes

Edit the relevant files. For example, to add a new command:

```bash
# Create a new command file
touch cmd/mycommand.go
```

### 4. Test Your Changes

```bash
make build
./bin/envswitch mycommand
```

### 5. Commit and Push

```bash
git add .
git commit -m "Add new feature: description"
git push origin feature/my-feature
```

### 6. Create a Pull Request

Go to GitHub and create a PR from your branch.

## Common Development Tasks

### Adding a New Command

1. Create `cmd/newcommand.go`
2. Define the command using Cobra
3. Register it in `init()` function
4. Build and test

Example:
```go
// cmd/mycommand.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description of your command",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Executing mycommand")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

### Adding a New Tool Integration

1. Create `pkg/tools/mytool.go`
2. Implement the `Tool` interface
3. Add tests
4. Update documentation

See `pkg/tools/gcloud.go` for an example.

### Debugging

Use Go's built-in debugging:

```bash
# Print debugging
go run main.go init

# Use delve debugger
dlv debug
```

Or use your IDE's debugger (VS Code, GoLand).

## Resources

### Go Learning
- [A Tour of Go](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### Project Dependencies
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML parsing

### Development Tools
- [GoLand](https://www.jetbrains.com/go/) - Go IDE
- [VS Code](https://code.visualstudio.com/) + [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
- [golangci-lint](https://golangci-lint.run/) - Linter

## Getting Help

- 📖 Read the [full documentation](./README.md)
- 💬 Ask in [GitHub Discussions](https://github.com/hugofrely/envswitch/discussions)
- 🐛 Report bugs in [Issues](https://github.com/hugofrely/envswitch/issues)
- 📧 Contact the maintainers

## Next Steps

1. ✅ Set up your development environment
2. 📖 Read [CONTRIBUTING.md](./CONTRIBUTING.md)
3. 🔍 Browse the [open issues](https://github.com/hugofrely/envswitch/issues)
4. 💻 Start coding!
5. 🚀 Submit your first PR

Happy coding! 🎉
