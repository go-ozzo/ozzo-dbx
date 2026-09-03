# Contributing to ozzo-dbx

Thank you for your interest in contributing to ozzo-dbx! This document covers how to build, test, and submit changes.

## Prerequisites

- **Go 1.22+** ([download](https://go.dev/dl/))
- **MySQL 8.0+** for integration tests (or use unit tests only)

## Building

```bash
go build ./...
```

## Running Tests

The project has two test suites:

```bash
# Unit tests (no database required — runs from root)
go test -race ./...

# Integration tests (requires MySQL — runs from integration/)
export DBX_MYSQL_DSN="root:pass@tcp(127.0.0.1:3306)/ozzo_dbx_test?parseTime=true"
cd integration && go test -race ./...
```

Unit tests cover SQL generation, quoting, struct mapping, and internal logic. Integration tests cover actual database operations (queries, transactions, CRUD).

## Code Style

- Run `gofmt -w .` and `golangci-lint run --timeout=5m` before every commit. CI enforces both.
- Run `go vet ./...` and fix all issues.
- Follow standard Go naming conventions (`ID`, `URL`, `HTTP` are uppercase).
- Handle every error or explicitly ignore with `_ =` and a comment explaining why.
- Exported types and functions must have doc comments.

## Where to Put Tests

- **Unit tests** (no database needed): `*_test.go` in the root directory, `package dbx`
- **Integration tests** (need MySQL): `integration/*_test.go`, `package dbx_test`

If your test only checks SQL output or struct mapping, it belongs in the root. If it executes queries against a real database, it belongs in `integration/`.

## Pull Request Workflow

1. Fork the repository and create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes. Keep commits focused and well-described.
3. Verify locally:
   ```bash
   go build ./...
   go vet ./...
   gofmt -w . && gofmt -l .
   golangci-lint run --timeout=5m
   go test -race ./...
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
