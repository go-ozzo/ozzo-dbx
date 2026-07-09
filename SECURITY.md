# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.5.x   | :white_check_mark: |

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues via:

1. **Private Security Advisory** (preferred):
   https://github.com/go-ozzo/ozzo-dbx/security/advisories/new

2. **GitHub Issues** (for less critical issues):
   https://github.com/go-ozzo/ozzo-dbx/issues

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact

### Response Timeline

- **Initial Response**: Within 72 hours
- **Fix & Disclosure**: Coordinated with reporter

## Security Considerations

ozzo-dbx builds SQL queries programmatically. Users should be aware of:

1. **SQL Injection** — Always use parameter binding (`{:name}` placeholders), never string concatenation
2. **Raw Queries** — `NewQuery()` with user input must use `Bind()` for safe parameter injection
3. **Logging** — `LogFunc` may expose query parameters in logs; sanitize in production

## Security Contact

- **GitHub Security Advisory**: https://github.com/go-ozzo/ozzo-dbx/security/advisories/new
- **Public Issues**: https://github.com/go-ozzo/ozzo-dbx/issues
