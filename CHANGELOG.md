# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- `go-sql-driver/mysql` v1.4.1 → v1.9.3
- `stretchr/testify` v1.4.0 → v1.12.1
- Removed `google.golang.org/appengine` indirect dependency
- `ioutil.ReadFile` → `os.ReadFile` (deprecated since Go 1.16)
- `reflect.PtrTo` → `reflect.PointerTo`
- `strings.Replace(..., -1)` → `strings.ReplaceAll`
- `time.Now().Sub(start)` → `time.Since(start)`
- Test DSN configurable via `DBX_MYSQL_DSN` environment variable (replaces hardcoded Travis CI credentials)
- CI: `actions/setup-go@v7`, `golangci-lint-action@v9`

## [1.5.0] - 2018-12-17

_Last release by original author [@qiangxue](https://github.com/qiangxue)._
