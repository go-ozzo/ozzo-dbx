package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderFuncMap_SqliteKeys(t *testing.T) {
	_, ok3 := BuilderFuncMap["sqlite3"]
	assert.True(t, ok3, "sqlite3 key should be registered")

	_, ok := BuilderFuncMap["sqlite"]
	assert.True(t, ok, "sqlite key should be registered for noncgo driver (modernc.org/sqlite)")
}

func TestBuilderFuncMap_AllDrivers(t *testing.T) {
	expected := []string{"sqlite3", "sqlite", "mysql", "postgres", "pgx", "mssql", "oci8"}
	for _, driver := range expected {
		_, ok := BuilderFuncMap[driver]
		assert.True(t, ok, "driver '%s' should be registered in BuilderFuncMap", driver)
	}
}
