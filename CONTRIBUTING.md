# Contributing to ozzo-dbx

Thank you for your interest in contributing to ozzo-dbx! This document covers how to build, test, and submit changes.

## Prerequisites

- **Go 1.21+** ([download](https://go.dev/dl/))
- **MySQL 8.0+** for integration tests (or use unit tests only)

## Building

```bash
go build ./...
```

## Running Tests

```bash
# Unit tests (no database required)
go test -run "TestBuilderFuncMap|TestSqliteBuilder|TestStandardBuilder|TestDefaultFieldMapFunc" ./...

# Full tests (requires MySQL)
export DBX_MYSQL_DSN="root:pass@tcp(127.0.0.1:3306)/ozzo_dbx_test?parseTime=true"
go test -race ./...
```

## Code Style

- Run `go fmt ./...` before every commit. CI enforces this.
- Run `go vet ./...` and fix all issues.
- Follow standard Go naming conventions (`ID`, `URL`, `HTTP` are uppercase).
- Handle every error or explicitly ignore with `_ =` and a comment explaining why.
- Exported types and functions must have doc comments.

## Pull Request Workflow

1. Fork the repository and create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes. Keep commits focused and well-described.
3. Verify locally:
   ```bash
   go fmt ./...
   go build ./...
   go vet ./...
   go test -run "TestBuilderFuncMap|TestSqliteBuilder|TestStandardBuilder" ./...
   ```
4. Push and open a pull request against `master`.
5. Wait for CI to pass. All checks must be green before merge.

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):
```
feat: add batch insert support
fix(sqlite): handle RENAME TABLE correctly
docs: update query building examples
```

## Finding Work

Check the [issues](https://github.com/go-ozzo/ozzo-dbx/issues) page for tasks labeled [`good first issue`](https://github.com/go-ozzo/ozzo-dbx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) if you're looking for a place to start.

## Reporting Bugs

Open an issue with a clear description, steps to reproduce, the Go version, database type and version you're using.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
