module github.com/go-ozzo/ozzo-dbx/integration

go 1.22

replace github.com/go-ozzo/ozzo-dbx => ../

require (
	github.com/go-ozzo/ozzo-dbx v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
	github.com/stretchr/testify v1.12.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)
