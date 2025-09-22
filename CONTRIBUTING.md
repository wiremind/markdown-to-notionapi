# Contributing to md2notion

Thank you for your interest in contributing to md2notion! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.25 or later
- Make (optional, for using the Makefile)
- golangci-lint (for linting)
- Container runtime (Docker, Podman, or nerdctl)

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/wiremind/markdown-to-notionapi.git
   cd markdown-to-notionapi
   ```

2. **Set up development environment**:
   ```bash
   make dev-setup
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Build the project**:
   ```bash
   make build
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards below

3. **Run quality checks**:
   ```bash
   make check  # Runs fmt, vet, lint, test, and security checks
   ```

4. **Test your changes**:
   ```bash
   # Unit tests
   make test
   
   # Integration test with container
   export NOTION_TOKEN="your-test-token"
   export TEST_PAGE_ID="your-test-page-id"
   make container-test
   ```

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

### Coding Standards

- **Go formatting**: Use `go fmt` (automated in `make fmt`)
- **Imports**: Use `goimports` for import organization
- **Linting**: All code must pass `golangci-lint` checks
- **Testing**: Add tests for new functionality
- **Documentation**: Update README.md for user-facing changes

### Commit Message Format

We follow conventional commits:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test-related changes
- `refactor:` for code refactoring
- `ci:` for CI/CD changes

Example:
```
feat: add support for table blocks

- Implement table parsing from markdown
- Add table block type to Notion API client
- Include tests for table conversion
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage
```

### Integration Tests

```bash
# Test with your own Notion page
export NOTION_TOKEN="your-integration-token"
export TEST_PAGE_ID="your-page-id"
make run-local
```

### Container Tests

```bash
# Build and test container
make container-test
```

## Code Quality

We maintain high code quality standards:

### Automated Checks

- **golangci-lint**: Comprehensive linting
- **gosec**: Security vulnerability scanning
- **govulncheck**: Known vulnerability detection
- **go vet**: Static analysis
- **Tests**: Unit and integration tests

### Running Checks Locally

```bash
# Run all quality checks
make check

# Individual checks
make fmt      # Format code
make vet      # Run go vet
make lint     # Run golangci-lint
make security # Run security scans
```

## Pull Request Process

1. **Ensure CI passes**: All GitHub Actions must be green
2. **Update documentation**: Update README.md if needed
3. **Add tests**: Include tests for new functionality
4. **Follow conventional commits**: Use proper commit message format
5. **Keep changes focused**: One feature/fix per PR

### PR Checklist

- [ ] Tests pass locally (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Security checks pass (`make security`)
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventional format
- [ ] Changes are backwards compatible (or clearly documented)

## Issue Reporting

### Bug Reports

Use the bug report template and include:

- Version information
- Environment details (binary, container, etc.)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

### Feature Requests

Use the feature request template and include:

- Problem description
- Proposed solution
- Alternative approaches considered
- Use cases and benefits

## Development Tools

### Available Make Targets

```bash
make help          # Show all available commands
make dev-setup     # Set up development environment
make build         # Build binary
make build-all     # Build for all platforms
make test          # Run tests
make coverage      # Generate coverage report
make lint          # Run linter
make security      # Run security checks
make check         # Run all quality checks
make clean         # Clean build artifacts
```

### Pre-commit Hooks (Optional)

Install pre-commit hooks to automatically run checks:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Release Process

Releases are automated via GitHub Actions when tags are pushed:

1. **Create a tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

2. **GitHub Actions will**:
   - Run all CI checks
   - Build multi-platform binaries
   - Create container images
   - Create a GitHub release

## Getting Help

- **Questions**: Open a discussion or issue
- **Chat**: Contact the maintainers
- **Documentation**: Check the README.md

## Code of Conduct

Please be respectful and professional in all interactions. We're all here to build something great together!

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).
