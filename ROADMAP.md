# Roadmap

## Current Focus

Reviving ozzo-dbx after 6 years of inactivity. Prioritizing stability, modernization, and community contributions.

## Near Term

- [x] Add `"sqlite"` driver key for modernc.org/sqlite (CGo-free)
- [x] Enable `DropColumn`, `RenameColumn`, `RenameTable` for modern SQLite
- [x] Support `[]*Struct` scanning via `Query.All()`
- [x] GitHub Actions CI with MySQL service
- [x] Codecov integration
- [ ] Review and merge pending community PRs (#104, #106)
- [ ] Update go.mod to Go 1.21
- [ ] Fix pre-existing `go vet` issues

## Medium Term

- [ ] Upsert support (ON CONFLICT / ON DUPLICATE KEY) — [#95](https://github.com/go-ozzo/ozzo-dbx/issues/95), [#73](https://github.com/go-ozzo/ozzo-dbx/issues/73)
- [ ] Batch insert — [#63](https://github.com/go-ozzo/ozzo-dbx/issues/63)
- [ ] `errors.Is` / `errors.As` support for query errors
- [ ] Subquery support — [#32](https://github.com/go-ozzo/ozzo-dbx/issues/32)
- [ ] PostgreSQL and SQLite service containers in CI
- [ ] golangci-lint integration
- [ ] Performance benchmarks

## Long Term

- [ ] Context-first API (all methods accept `context.Context`)
- [ ] `Exists()` / `Count()` convenience methods
- [ ] Statement cache
- [ ] `ToSQL()` for query debugging
- [ ] Transaction helper with auto-commit/rollback

## Related Projects

- [coregx/relica](https://github.com/coregx/relica) — Modern type-safe SQL query builder inspired by ozzo-dbx

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get involved. Issues labeled [`good first issue`](https://github.com/go-ozzo/ozzo-dbx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) are a great place to start.
