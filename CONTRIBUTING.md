# Contributing to EnvSwitch

Thank you for your interest in contributing to EnvSwitch! 🎉

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
   ```bash
   git clone https://github.com/YOUR_USERNAME/envswitch.git
   cd envswitch
   ```
3. **Install Go** (version 1.21 or higher)
4. **Install dependencies**
   ```bash
   go mod download
   ```

## Development Workflow

### Building

```bash
# Build the binary
make build

# Build and install to /usr/local/bin
make install

# Build for all platforms
make build-all
```

### Testing

```bash
# Run tests
make test

# Run with coverage
go test -cover ./...
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Run `go vet` to check for issues
- Write tests for new functionality

### Making Changes

1. **Create a branch** for your changes
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** and commit them
   ```bash
   git add .
   git commit -m "Add new feature: description"
   ```

3. **Push to your fork**
   ```bash
   git push origin feature/my-new-feature
   ```

4. **Create a Pull Request** on GitHub

### Commit Message Guidelines

- Use clear, descriptive commit messages
- Start with a verb in present tense (Add, Fix, Update, etc.)
- Reference issues when applicable (#123)

Examples:
```
Add GCloud snapshot capture functionality
Fix environment switching bug #45
Update README with installation instructions
```

## Project Structure

```
envswitch/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── create.go          # Create command
│   ├── switch.go          # Switch command
│   └── ...
├── pkg/                    # Public packages
│   ├── environment/        # Environment management
│   ├── tools/             # Tool integrations
│   └── ...
├── internal/              # Private packages
│   ├── config/            # Configuration
│   ├── storage/           # File operations
│   └── ...
└── main.go                # Entry point
```

## Areas We Need Help

### High Priority
- [ ] Implementing tool integrations (GCloud, Kubectl, AWS, etc.)
- [ ] Environment switching logic with backup/restore
- [ ] Snapshot capture and restore functionality
- [ ] Testing and bug fixes

### Medium Priority
- [ ] Shell integration (bash, zsh, fish)
- [ ] Auto-completion scripts
- [ ] Documentation improvements
- [ ] Examples and tutorials

### Future Features
- [ ] Encryption support
- [ ] TUI (Terminal User Interface)
- [ ] Template system
- [ ] Sync with Git
- [ ] Plugin system

## Adding a New Tool Integration

To add support for a new CLI tool:

1. Create a new file in `pkg/tools/` (e.g., `terraform.go`)
2. Implement the `Tool` interface:
   ```go
   type Tool interface {
       Name() string
       IsInstalled() bool
       Snapshot(snapshotPath string) error
       Restore(snapshotPath string) error
       GetMetadata() (map[string]interface{}, error)
       ValidateSnapshot(snapshotPath string) error
       Diff(snapshotPath string) ([]Change, error)
   }
   ```
3. Add tests for your implementation
4. Update documentation

## Questions or Need Help?

- Open an issue on GitHub
- Join our discussions
- Ask in your PR if you need guidance

## Code of Conduct

Be respectful, inclusive, and considerate. We're all here to learn and build something useful together.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🙏
