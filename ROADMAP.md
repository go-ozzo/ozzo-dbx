# Roadmap

## Current State: v1.7.0

Working upserts on MySQL, PostgreSQL and SQLite, model-level upsert, locking reads, and a transaction data-loss fix.

## Completed (v1.7.0)

- [x] `ModelQuery.Upsert()` — [#95](https://github.com/go-ozzo/ozzo-dbx/issues/95)
- [x] MySQL `Upsert()` executes the full `ON DUPLICATE KEY UPDATE` statement; PostgreSQL/SQLite require a conflict target
- [x] `Upsert()` for SQLite
- [x] Locking reads: `ForUpdate()`, `ForShare()`, `ForUpdateSkipLocked()`
- [x] `Transactional()` no longer reports success for a transaction rolled back by a canceled context
- [x] SQLite `LIKE` escaping (`ESCAPE '\'`)
- [x] Pointer-receiver `TableName()` honoured for values and slices
- [x] `Errors.Unwrap()` for `errors.Is` / `errors.As`

## Completed (v1.6.0)

- [x] Add `"sqlite"` driver key for modernc.org/sqlite (CGo-free)
- [x] Enable `DropColumn`, `RenameColumn`, `RenameTable` for modern SQLite
- [x] Support `[]*Struct` scanning via `Query.All()`
- [x] `ScanStruct`/`All()` strip table alias prefix from column names
- [x] `StructInfo`/`FieldInfo` exported for struct-to-column mapping inspection
- [x] `LogBinaryFormatter` for `[]byte` serialization in query logging
- [x] GitHub Actions CI (unit + integration), golangci-lint, Codecov
- [x] Go 1.22, dependency updates, 66 lint issues resolved
- [x] Integration tests separated into own module — clean `go.mod` for consumers
- [x] Review and merge community PRs (#106, #108)
- [x] pgx `LastInsertId()` compatibility

## Next (v1.8.0+)

- [ ] Subquery support (`SelectQuery` as `Expression`) — [#32](https://github.com/go-ozzo/ozzo-dbx/issues/32)
- [ ] PostgreSQL and SQLite service containers in CI
- [ ] Performance benchmarks

## Long Term

- [ ] Context-first API (all methods accept `context.Context`)
- [ ] `Exists()` / `Count()` convenience methods
- [ ] Statement cache
- [ ] Batch insert
- [ ] Transaction helper with auto-commit/rollback

## Related Projects

- [coregx/relica](https://github.com/coregx/relica) — Modern type-safe SQL query builder inspired by ozzo-dbx

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get involved. Issues labeled [`good first issue`](https://github.com/go-ozzo/ozzo-dbx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) are a great place to start.
