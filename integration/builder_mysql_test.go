package dbx_test

import (
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestMysqlBuilder_RenameColumn(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	b := dbx.NewMysqlBuilder(db, db.DB())
	db.Builder = b

	// This requires SHOW CREATE TABLE against real MySQL
	q := b.RenameColumn("customer", "email", "e")
	assert.Equal(t, q.SQL(), "ALTER TABLE `customer` CHANGE `email` `e` varchar(128) NOT NULL")
}
