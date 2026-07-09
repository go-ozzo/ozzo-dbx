# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Support for scanning into pointer slices (`[]*Struct`) via `Query.All()` ([#48](https://github.com/go-ozzo/ozzo-dbx/issues/48), originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- `"sqlite"` driver key in `BuilderFuncMap` for [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (CGo-free SQLite driver) (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- `RenameTable()` for SQLite builder (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- GitHub Actions CI with MySQL 8.0 service
- Codecov integration (OIDC)
- `CONTRIBUTING.md`
- `CODEOWNERS`

### Fixed
- SQLite `DropColumn()` and `RenameColumn()` now generate standard `ALTER TABLE` SQL instead of returning errors (requires SQLite 3.25.0+ for rename, 3.35.0+ for drop column) (originally suggested by [@ganigeorgiev](https://github.com/ganigeorgiev))
- Broken example test function names (`ExampleSchemaBuilder`, `ExampleDB_Open`)

### Changed
- Minimum Go version: 1.13 → 1.21
- Test DSN configurable via `DBX_MYSQL_DSN` environment variable (replaces hardcoded Travis CI credentials)

## [1.5.0] - 2018-12-17

_Last release by original author [@qiangxue](https://github.com/qiangxue)._
