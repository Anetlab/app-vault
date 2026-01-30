# Contributing to SecureVault

Thank you for your interest in contributing to SecureVault! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and encourage diverse perspectives
- Focus on what is best for the community

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/app-vault.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes: `git commit -am 'Add new feature'`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Create a Pull Request

## Development Setup

### Prerequisites

- Go 1.23 or later
- PostgreSQL 12 or later
- Docker (optional, for running PostgreSQL)

### Setup

```bash
# Clone the repository
git clone https://github.com/securevault/app-vault.git
cd app-vault

# Install dependencies
go mod download

# Start PostgreSQL (using Docker)
docker-compose up -d

# Run the server
go run cmd/server/main.go
```

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) and the guidelines in `.github/instructions/go.instructions.md`.

Key points:
- Use `gofmt` to format your code
- Keep functions small and focused
- Write clear, descriptive variable names
- Add comments for exported functions and complex logic
- Use early returns to reduce nesting
- Handle all errors properly

### Security Requirements

**Critical security rules** (from `.github/instructions/copilot-instructions.md`):

1. **Never log secrets or plaintext** - No sensitive data in logs, errors, or traces
2. **Use constant-time comparisons** - For tokens, hashes, and derived keys
3. **Validate all inputs** - Length limits, character sets, prevent injection
4. **Parameterized queries only** - No string concatenation in SQL
5. **Zero sensitive buffers** - Clear memory after use (best effort in Go)
6. **Use AEAD encryption** - XChaCha20-Poly1305 only, no custom crypto

### Testing

- Write unit tests for all new functionality
- Ensure tests pass: `go test ./...`
- Aim for high code coverage
- Test error cases and edge conditions
- No secrets or credentials in test code

Example test structure:

```go
func TestMyFunction(t *testing.T) {
    // Setup
    input := "test input"
    
    // Execute
    result, err := MyFunction(input)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

## Pull Request Process

1. **Update documentation** - Update README.md or relevant docs if needed
2. **Add tests** - Ensure new code has adequate test coverage
3. **Run tests** - All tests must pass: `go test ./...`
4. **Run linters** - `go vet ./...` must pass
5. **Write clear commit messages** - Describe what and why
6. **Keep PRs focused** - One feature or fix per PR
7. **Respond to reviews** - Address feedback promptly and professionally

### Commit Message Format

```
type: brief description

Detailed explanation of what changed and why.

Fixes #123
```

Types:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

## What to Contribute

### Good First Issues

Look for issues labeled `good first issue` or `help wanted`.

### Ideas for Contributions

- **Features**: Service principal management UI, rate limiting improvements
- **Documentation**: Tutorials, deployment guides, architecture diagrams
- **Tests**: Increase test coverage, add integration tests
- **Performance**: Optimize database queries, caching strategies
- **Security**: Security audits, penetration testing reports

### What We're NOT Looking For

- Breaking API changes (discuss first in an issue)
- Features that violate zero-knowledge principles
- Custom cryptography implementations
- UI/frontend code (backend only for now)

## Architecture Guidelines

### Package Structure

```
cmd/server/         - Application entry point
internal/
  api/             - HTTP handlers and middleware
  service/         - Business logic layer
  db/              - Database repository layer
  crypto/          - Cryptography operations
  models/          - Shared data structures
migrations/        - SQL schema migrations
```

### Layer Responsibilities

- **API Layer**: HTTP request/response handling, validation, authentication
- **Service Layer**: Business logic, orchestration, transactions
- **Database Layer**: Data persistence, queries, transactions
- **Crypto Layer**: Encryption, decryption, key derivation

### Adding a New Endpoint

1. Add handler in `internal/api/handlers.go`
2. Add business logic in appropriate service file
3. Add database operations if needed in `internal/db/database.go`
4. Register route in `cmd/server/main.go`
5. Update API documentation in README.md
6. Add tests

## Database Migrations

### Creating a Migration

Create a new SQL file in `migrations/` with naming pattern: `NNN_description.sql`

Example: `002_add_sessions_table.sql`

```sql
-- Add session management
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

Migrations must be:
- Idempotent (can run multiple times safely)
- Forward-compatible (no breaking changes to existing data)

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/crypto

# With coverage
go test ./... -cover

# Verbose output
go test ./... -v
```

### Test Coverage

Aim for:
- Critical paths: 90%+ coverage
- Crypto operations: 100% coverage
- Business logic: 80%+ coverage
- HTTP handlers: 70%+ coverage

## Documentation

- Update README.md for user-facing changes
- Update EXAMPLES.md for new API endpoints
- Add godoc comments for exported functions
- Keep comments up-to-date with code changes

## Questions?

- Open an issue for questions about contributing
- Check existing issues and PRs for similar work
- Join discussions in GitHub Discussions (if enabled)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
