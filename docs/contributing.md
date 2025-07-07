# Contributing to Bedrock Forge

Thank you for your interest in contributing to Bedrock Forge! This guide will help you get started with development and releases.

## Development Setup

### Prerequisites

- Go 1.21+
- Git
- Make (optional)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/your-org/bedrock-forge
cd bedrock-forge

# Install dependencies
go mod download

# Build the binary
go build -o bedrock-forge ./cmd/bedrock-forge

# Run tests
go test ./...

# Verify everything works
./bedrock-forge version
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write code following Go best practices
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Run linting
go vet ./...
gofmt -s -l .

# Test build
go build -o bedrock-forge ./cmd/bedrock-forge

# Test with examples (if available)
./bedrock-forge validate examples/
./bedrock-forge generate examples/ ./test-output
```

### 4. Commit Changes

We use [Conventional Commits](https://www.conventionalcommits.org/) for automatic changelog generation and releases.

#### Commit Message Format

```
type(scope): description

[optional body]

[optional footer(s)]
```

#### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `build`: Build system changes
- `ci`: CI/CD changes
- `deps`: Dependency updates

#### Examples

```bash
# New feature
git commit -m "feat: add support for custom IAM roles"

# Bug fix
git commit -m "fix: handle missing Lambda function error gracefully"

# Documentation
git commit -m "docs: add examples for knowledge base configuration"

# Breaking change
git commit -m "feat!: change YAML schema for action groups

BREAKING CHANGE: action groups now require explicit agent reference"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Create a pull request on GitHub with:
- Clear description of changes
- Link to any related issues
- Screenshots/examples if applicable

## CI/CD Pipeline

### Automated Testing

Our CI pipeline runs:

1. **Unit Tests**: `go test ./...`
2. **Linting**: `go vet`, `gofmt`, `staticcheck`
3. **Security Scan**: `gosec`
4. **Build**: Multi-platform binary builds
5. **Integration Tests**: End-to-end validation
6. **Example Validation**: Validate all example configurations

### Branch Protection

- `main` branch requires PR reviews
- All CI checks must pass
- No direct pushes to `main`

## Release Process

We use [Release Please](https://github.com/googleapis/release-please) for automated releases.

### How Releases Work

1. **Conventional Commits**: Your commit messages drive the release process
2. **Automatic PRs**: Release Please creates PRs with changelogs
3. **Merge to Release**: Merging the PR triggers the release
4. **Multi-Platform Builds**: Automatic binary builds for all platforms
5. **GitHub Action**: Updated action.yml for each release

### Release Types

- `fix:` → Patch release (v1.0.1)
- `feat:` → Minor release (v1.1.0)
- `feat!:` or `BREAKING CHANGE:` → Major release (v2.0.0)

### Manual Release

To trigger a release manually:

```bash
# Create a release commit
git commit -m "chore: release v1.0.0"

# Or use a breaking change
git commit -m "feat!: major API changes

BREAKING CHANGE: updated YAML schema format"
```

## Code Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Handle errors appropriately
- Write testable code

### Testing

- Write unit tests for new functionality
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for good test coverage

### Documentation

- Update README.md for user-facing changes
- Add/update resource documentation in `docs/resources/`
- Include examples for new features
- Update getting started guide if needed

## Project Structure

```
bedrock-forge/
├── cmd/
│   └── bedrock-forge/          # Main CLI application
├── internal/
│   ├── commands/               # CLI commands
│   ├── models/                 # Data models
│   ├── generator/              # Terraform generation
│   ├── parser/                 # YAML parsing
│   ├── registry/               # Resource registry
│   └── validation/             # Validation logic
├── examples/                   # Example configurations
├── docs/                       # Documentation
├── .github/
│   └── workflows/              # CI/CD workflows
└── tests/                      # Integration tests
```

## Adding New Resources

### 1. Define the Model

Create a new model in `internal/models/`:

```go
// internal/models/newresource.go
type NewResource struct {
    Kind       string            `yaml:"kind"`
    Metadata   ResourceMetadata  `yaml:"metadata"`
    Spec       NewResourceSpec   `yaml:"spec"`
}

type NewResourceSpec struct {
    // Define your fields
    Name        string `yaml:"name"`
    Description string `yaml:"description,omitempty"`
}
```

### 2. Add Generator

Create generator in `internal/generator/`:

```go
// internal/generator/new_resource_generator.go
func GenerateNewResource(resource *models.NewResource) (string, error) {
    // Generate Terraform HCL
}
```

### 3. Register Resource

Update `internal/registry/resource_registry.go`:

```go
func (r *ResourceRegistry) RegisterResources() {
    // Add your resource
    r.resourceTypes["NewResource"] = &models.NewResource{}
}
```

### 4. Add Tests

Create tests in appropriate directories:

```go
// internal/generator/new_resource_generator_test.go
func TestGenerateNewResource(t *testing.T) {
    // Test your generator
}
```

### 5. Add Documentation

Create `docs/resources/new-resource.md` with:
- Overview
- Basic example
- Complete example
- Specification
- Best practices

### 6. Add Example

Create example in `examples/`:

```yaml
# examples/new-resource-example.yml
kind: NewResource
metadata:
  name: "example-resource"
spec:
  description: "Example new resource"
```

## Getting Help

- **Questions**: Open a [Discussion](https://github.com/your-org/bedrock-forge/discussions)
- **Bugs**: Open an [Issue](https://github.com/your-org/bedrock-forge/issues)
- **Features**: Open an [Issue](https://github.com/your-org/bedrock-forge/issues) with the enhancement label

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## License

By contributing to Bedrock Forge, you agree that your contributions will be licensed under the same license as the project (MIT License).