# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.7.0] - 2026-10-26

### Added
- `ModelQuery.Upsert(attrs...)`: inserts a model or updates the row with the same primary key, on MySQL, PostgreSQL and SQLite ([#95](https://github.com/go-ozzo/ozzo-dbx/issues/95)). Fields are chosen as for `Insert()`/`Update()`; fields passed to `Exclude()` are neither inserted nor updated; the primary key is never updated; a zero single-column primary key is left out of the insertion (auto-increment). The primary key field is not filled after the call
- `Upsert()` for SQLite (`INSERT ... ON CONFLICT (...) DO UPDATE SET ...`, SQLite 3.24.0+)
- Locking reads: `SelectQuery.ForUpdate()`, `ForShare()` and `ForUpdateSkipLocked()` append `FOR UPDATE`, `FOR SHARE` or `FOR UPDATE SKIP LOCKED` after `ORDER BY`/`LIMIT`/`OFFSET`. A locking read must be built from a transaction (`tx.Select()`): from a `DB` the query returns an error on every database, SQLite included, since the lock would be released as soon as the statement ends. Inside a transaction the clause is left out for SQLite (no row locks); an error for SQL Server and Oracle, and when combined with `Union()`, `GroupBy()` or `Distinct()`. Optional `LockQueryBuilder` interface for custom query builders (not added to `QueryBuilder`, so existing implementations keep compiling)
- `Errors.Unwrap() []error`, so `errors.Is` and `errors.As` see each error of the list (e.g. the callback's error when `Transactional()` also failed to roll back)

### Fixed
- **Data loss:** `Transactional()` and `TransactionalContext()` returned `nil` when the context was canceled while the callback ran. database/sql had already rolled the transaction back, so the callback's writes were lost while the caller was told they were committed. They now return the context error (`context.Canceled` / `context.DeadlineExceeded`). A callback that ends the transaction itself with `tx.Commit()` or `tx.Rollback()` and returns `nil` still gets `nil`
- MySQL `Upsert()` executed a plain `INSERT`: the `ON DUPLICATE KEY UPDATE` clause and its parameters were shown by `SQL()` but never sent to the server, so an existing key failed with a duplicate-key error (the same bug was fixed for PostgreSQL in [#70](https://github.com/go-ozzo/ozzo-dbx/pull/70))
- PostgreSQL `Upsert()` without constraint columns generated `ON CONFLICT DO UPDATE` without a conflict target, which PostgreSQL rejects; it now returns an error before executing (SQLite likewise)
- `Upsert()` with no column left to update besides the constraint columns generated an empty update clause; it now returns an error before executing
- `GetTableName()` (the default `DB.TableMapper`) ignored a `TableName()` method with a pointer receiver when given a struct value or a slice of struct values, e.g. `db.Select().All(&[]Customer{})` without `From()`, and fell back to the snake-case type name. Code that relied on the fallback name will now query the table returned by `TableName()`
- SQLite: `Like()` and the other LIKE expressions never matched an escaped `%` or `_` (e.g. `Like("name", "50%")` did not match "50% off"): the values are escaped with a backslash, but SQLite has no default LIKE escape character. An `ESCAPE '\'` clause is now added on SQLite only (MySQL and PostgreSQL use the backslash by default)

### Changed
- `Upsert()` leaves the constraint columns out of the update part on every database. On MySQL `constraints` was ignored before: now a row matched through another unique index keeps its own values in the constraint columns (e.g. its primary key). On PostgreSQL the conflict columns were assigned their own values, a no-op except for `UPDATE OF` triggers and case-insensitive collations

## [1.6.0] - 2026-09-03

### Added
- Support for scanning into pointer slices (`[]*Struct`) via `Query.All()` ([#48](https://github.com/go-ozzo/ozzo-dbx/issues/48), originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- `"sqlite"` driver key in `BuilderFuncMap` for [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (CGo-free SQLite driver) (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- `RenameTable()` for SQLite builder (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- `StructInfo` and `FieldInfo` types exported with getter methods for struct-to-column mapping inspection ([#106](https://github.com/go-ozzo/ozzo-dbx/pull/106), [@BourgeoisBear](https://github.com/BourgeoisBear))
- `LogBinaryFormatter` for controlling `[]byte` serialization in query logging ([#106](https://github.com/go-ozzo/ozzo-dbx/pull/106), [@BourgeoisBear](https://github.com/BourgeoisBear))
- GitHub Actions CI with MySQL 8.0 service, golangci-lint, Codecov (OIDC)
- `CONTRIBUTING.md`, `CODEOWNERS`, `CODE_OF_CONDUCT.md`, `SECURITY.md`

### Fixed
- `ScanStruct` and `All()` now strip table alias prefix from column names (e.g., `src.qualified_name` → `qualified_name`) before matching struct tags, preventing silent data loss when drivers return prefixed column names ([#111](https://github.com/go-ozzo/ozzo-dbx/pull/111))
- SQLite `DropColumn()` and `RenameColumn()` now generate standard `ALTER TABLE` SQL instead of returning errors (requires SQLite 3.25.0+ for rename, 3.35.0+ for drop column) (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- pgx driver compatibility: skip `LastInsertId()` for drivers that don't support it ([#94](https://github.com/go-ozzo/ozzo-dbx/issues/94))
- Broken example test function names (`ExampleSchemaBuilder`, `ExampleDB_Open`)
- 66 golangci-lint issues resolved across the codebase (errcheck, staticcheck, govet)

### Changed
- **Minimum Go version: 1.13 → 1.22** (Go 1.21 or older is no longer supported)
- **Removed `go-sql-driver/mysql` from main `go.mod`** — integration tests moved to `integration/` subdirectory with its own module. Consumers no longer pull MySQL driver or `filippo.io/edwards25519` into their dependency graph
- `stretchr/testify` v1.4.0 → v1.12.1
- Removed `google.golang.org/appengine` indirect dependency
- Modernized stdlib usage: `os.ReadFile`, `strings.ReplaceAll`, `time.Since`, `reflect.PointerTo`
- Test DSN configurable via `DBX_MYSQL_DSN` environment variable (replaces hardcoded Travis CI credentials)
- CI: split into unit tests (no DB) and integration tests (MySQL 8.0), `actions/setup-go@v7`, `golangci-lint-action@v9`

## [1.5.0] - 2018-12-17

_Last release by original author [@qiangxue](https://github.com/qiangxue)._
