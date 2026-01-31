# GitHub Copilot Instructions for wallet-backup

## Project Overview

This is a Go CLI application that parses SMS banking notifications and converts them into CSV expense records. The application is actively developed and follows Go best practices.

## Documentation Standards

### README.md Rules

- **README.md should ONLY contain**:
  - Project description and purpose
  - Installation instructions
  - Usage examples
  - Configuration options (if any)
  - License and contribution guidelines

- **README.md should NOT contain**:
  - Project structure diagrams
  - Architecture documentation
  - Design patterns explanations
  - Internal implementation details
  - Development guidelines

### Where to Document What

- **README.md**: User-facing documentation (installation, usage)
- **ARCHITECTURE.md**: Project structure, design patterns, technical decisions
- **CONTRIBUTING.md**: Development setup, coding standards, PR guidelines
- **Package comments**: Go doc comments for public APIs
- **Inline comments**: Complex logic explanations within code

## Code Organization

### Package Structure

```
cmd/           - CLI commands (Cobra)
internal/      - Private application code
  ├── models/     - Data structures and constants
  ├── parser/     - SMS parsing logic (one file per bank)
  ├── categorizer/ - Transaction categorization
  ├── utils/      - Shared helper functions
  └── writer/     - Output file generation
```

### File Naming Conventions

- Use lowercase with underscores for multi-word files: `transaction.go`, `banque_misr.go`
- One concept per file (e.g., separate parsers for different banks)
- Package name should match directory name

## Go Best Practices

### Code Style

- Follow standard Go formatting (use `gofmt` or `goimports`)
- Use meaningful variable names (no single-letter names except in very short scopes)
- Keep functions focused and small (prefer < 50 lines)
- Use early returns to reduce nesting

### Design Patterns

- **Factory Pattern**: Use `New()` constructor functions for all structs
- **Dependency Injection**: Pass dependencies through constructors
- **Separation of Concerns**: Each package has one clear responsibility
- **Single Responsibility**: Each file should have a focused purpose

### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Return errors, don't panic (except in truly exceptional cases)
- Check all errors, even in defer statements

### Package Organization

- Use `internal/` for private packages (prevents external imports)
- Group related functionality together
- Keep public APIs minimal and focused
- Document all exported functions, types, and constants

## Testing Standards

- Write tests in `*_test.go` files
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for high coverage on business logic (parser, categorizer)
- Test edge cases and error paths

## Git Workflow

### Commit Messages

Follow conventional commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

Example: `feat: add support for Bank of Alexandria SMS parsing`

### Branch Strategy

- `main` - Production-ready code
- `master` - Default branch (currently)
- Feature branches: `feature/description`
- Bug fixes: `fix/description`

## Adding New Features

### Adding a New Bank

1. Create new parser file in `internal/parser/` (e.g., `nbe.go`)
2. Implement parsing function following existing pattern
3. Add bank-specific keywords to categorizer if needed
4. Update tests
5. Update README with supported bank list

### Adding New Categories

1. Add constant to `internal/models/transaction.go`
2. Add categorization logic to `internal/categorizer/categorizer.go`
3. Add keywords for the new category
4. Update tests

## Dependencies

- Minimize external dependencies
- Only use well-maintained, popular packages
- Document why each dependency is needed
- Keep `go.mod` clean with `go mod tidy`

## Performance Considerations

- This tool processes SMS backups which can be large (1000s of messages)
- Avoid unnecessary allocations in tight loops
- Use regexp compilation once, not per iteration
- Consider buffering for file I/O operations

## Security Considerations

- Never log or expose sensitive financial data
- Sanitize user inputs (file paths, etc.)
- Be careful with regex to avoid ReDoS attacks
- Validate all external data before processing

## CLI Best Practices

- Use Cobra for consistent command structure
- Always provide `--help` information
- Use meaningful flag names (long and short forms)
- Validate arguments before processing
- Provide clear error messages to users
- Show progress for long-running operations

## Code Review Checklist

- [ ] Code follows Go conventions
- [ ] All exports are documented
- [ ] Error handling is comprehensive
- [ ] Tests are included
- [ ] No sensitive data in logs
- [ ] README updated if user-facing changes
- [ ] Commit messages are descriptive

## Common Patterns in This Project

### Parser Pattern

```go
func parseBankMessage(tx *models.Transaction, body string) {
    // Set target group
    tx.TargetGroup = "Bank_Name"

    // Skip non-transaction messages
    if shouldSkip(body) {
        tx.TargetGroup = ""
        return
    }

    // Parse using regex
    // Set tx.Amount, tx.Currency, tx.Payee, etc.
}
```

### Factory Pattern

```go
type Service struct {
    dependency *Dependency
}

func New(dep *Dependency) *Service {
    return &Service{
        dependency: dep,
    }
}
```

## Useful Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build
go build -o sms-parser

# Run linter (if golangci-lint is installed)
golangci-lint run

# Update dependencies
go mod tidy

# View dependencies
go list -m all
```

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Cobra Documentation](https://github.com/spf13/cobra)
