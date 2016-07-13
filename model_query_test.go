package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Item struct {
	ID   int `db:"pk"`
	Name string
}

func TestModelQuery_Insert(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	result, err := db.NewQuery("INSERT INTO item (name) VALUES ('test')").Execute()
	if assert.Nil(t, err) {
		rows, _ := result.RowsAffected()
		assert.Equal(t, rows, int64(1), "Result.RowsAffected()")
		lastID, _ := result.LastInsertId()
		assert.Equal(t, lastID, int64(6), "Result.LastInsertId()")
	}
}
