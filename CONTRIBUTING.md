# Contributing to OVHcloud Driver for Rancher

Thank you for your interest in contributing! We welcome contributions of all kinds.

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+** — [Install Go](https://golang.org/doc/install)
- **Docker** — [Install Docker](https://docs.docker.com/get-docker/)
- **Docker Machine** — [Install Docker Machine](https://docs.docker.com/machine/install-machine/)
- **OVH Account** — [Sign up](https://www.ovhcloud.com/)
- **Make** — Usually pre-installed on Linux/macOS

### Development Setup

1. **Fork and clone the repository:**

```bash
git clone https://github.com/YOUR_USERNAME/docker-machine-driver-ovh.git
cd docker-machine-driver-ovh
```

2. **Install dependencies:**

```bash
go mod download
go mod vendor
```

3. **Build the driver:**

```bash
make build
```

4. **Install locally for testing:**

```bash
make install
# Or manually:
ln -sf $(pwd)/docker-machine-driver-ovh /usr/local/bin/
```

5. **Configure OVH credentials:**

Create `~/.ovh.conf` or set environment variables:

```bash
export OVH_APPLICATION_KEY="your_key"
export OVH_APPLICATION_SECRET="your_secret"
export OVH_CONSUMER_KEY="your_consumer_key"
```

6. **Test your build:**

```bash
docker-machine create -d ovh --ovh-region GRA1 test-node
docker-machine rm -f test-node
```

## 🎨 Code Style

We follow standard Go conventions with additional guidelines:

### Go Code Standards

- **Formatting:** All code must pass `gofmt -s`
- **Linting:** Code must pass `golangci-lint` checks
- **Error handling:** Always check and handle errors explicitly
- **Comments:** Public functions and types must have doc comments
- **Naming:** Use descriptive names; avoid abbreviations unless widely known

### Running Linters

We use [golangci-lint](https://golangci-lint.run/) with project-specific configuration:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters (uses .golangci.yml)
golangci-lint run

# Auto-fix where possible
golangci-lint run --fix
```

Our configuration: [.golangci.yml](.golangci.yml)

### Code Formatting

```bash
# Format all Go files
gofmt -s -w .

# Check formatting without changes
gofmt -s -l .
```

## 🧪 Testing

### Running Tests

```bash
# Run unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Requirements

- **Unit tests** are required for all new features and bug fixes
- Tests must pass before merging
- Aim for >80% code coverage on new code
- Mock external API calls using test fixtures
- Use table-driven tests for multiple scenarios

### Writing Tests

Example test structure:

```go
func TestDriverFeature(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid case", "input", "expected", false},
		{"error case", "bad", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DriverFunction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
```

### Integration Testing

For testing against real OVH infrastructure:

```bash
# Set test project
export OVH_PROJECT="test-project-id"

# Run integration tests (requires OVH credentials)
go test -tags=integration ./...
```

⚠️ **Note:** Integration tests will create real resources and may incur costs.

## 📝 Pull Request Process

### Before Submitting

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-description
   ```

2. **Make your changes:**
   - Write code following our style guidelines
   - Add/update tests
   - Update documentation if needed

3. **Run checks locally:**
   ```bash
   # Format code
   gofmt -s -w .
   
   # Run linters
   golangci-lint run
   
   # Run tests
   go test ./...
   
   # Build to ensure it compiles
   make build
   ```

4. **Commit with clear messages:**
   ```bash
   git commit -m "feat: add support for X"
   # or
   git commit -m "fix: resolve issue with Y"
   ```

   **Commit message format:**
   - `feat:` — New features
   - `fix:` — Bug fixes
   - `docs:` — Documentation changes
   - `test:` — Test additions/changes
   - `refactor:` — Code refactoring
   - `chore:` — Build/tooling changes

### Submitting Your PR

1. **Push your branch:**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub

3. **Fill out the PR template:**
   - Describe what changed and why
   - Reference related issues (`Fixes #123`)
   - Note any breaking changes
   - Add screenshots for UI changes

4. **Respond to review feedback:**
   - Address comments promptly
   - Update your branch as needed
   - Request re-review when ready

### PR Requirements

✅ **Your PR must:**
- Pass all CI checks (linting, tests, build)
- Include tests for new functionality
- Update documentation if behavior changes
- Follow our code style guidelines
- Have a clear, descriptive title
- Reference any related issues

❌ **Your PR will be blocked if:**
- CI checks fail
- Code coverage decreases significantly
- Code style violations exist
- Tests are missing or failing

## 🐛 Reporting Bugs

Found a bug? Please [open an issue](https://github.com/sneederco/docker-machine-driver-ovh/issues/new) with:

**Required information:**
- **Description:** Clear description of the bug
- **Steps to reproduce:** Numbered steps to trigger the issue
- **Expected behavior:** What should happen
- **Actual behavior:** What actually happens
- **Environment:**
  - Driver version: `docker-machine-driver-ovh --version`
  - Docker Machine version: `docker-machine --version`
  - OS: `uname -a`
  - Go version (if building from source): `go version`
- **Logs:** Relevant error messages or logs (use `-D` flag for debug mode)

**Example:**

```bash
# Run with debug output
docker-machine -D create -d ovh --ovh-region GRA1 test-node
```

## 💡 Feature Requests

Have an idea? [Open an issue](https://github.com/sneederco/docker-machine-driver-ovh/issues/new) with:

- **Use case:** Why is this feature needed?
- **Proposed solution:** How should it work?
- **Alternatives:** Other approaches you've considered
- **Additional context:** Mockups, examples, references

## 📖 Documentation

Documentation improvements are always welcome!

**Documentation locations:**
- Main README: `README.md`
- Detailed guides: `docs/*.md`
- Code comments: Inline in `.go` files
- API docs: Generated from code comments

**To improve docs:**
1. Edit the relevant Markdown file
2. Preview locally (use any Markdown viewer)
3. Submit a PR with your changes

## 🤝 Code of Conduct

Be respectful, inclusive, and constructive. We're all here to build something great together.

**Expected behavior:**
- Use welcoming and inclusive language
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the project

**Unacceptable behavior:**
- Harassment, insults, or derogatory comments
- Trolling or inflammatory remarks
- Public or private harassment
- Publishing others' private information

## ❓ Questions?

- **General questions:** [Open a discussion](https://github.com/sneederco/docker-machine-driver-ovh/discussions)
- **Bug reports:** [Open an issue](https://github.com/sneederco/docker-machine-driver-ovh/issues)
- **Security issues:** Contact the maintainers privately

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to the OVHcloud Driver for Rancher! 🎉
