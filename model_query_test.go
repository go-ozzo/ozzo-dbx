package dbx

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// execRecorder is an Executor that records the executed statements instead of running them.
type execRecorder struct {
	sql  []string
	args [][]interface{}
}

func (r *execRecorder) Exec(query string, args ...interface{}) (sql.Result, error) {
	r.sql = append(r.sql, query)
	r.args = append(r.args, args)
	return nil, nil
}

func (r *execRecorder) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.Exec(query, args...)
}

func (r *execRecorder) Query(string, ...interface{}) (*sql.Rows, error) {
	panic("not expected")
}

func (r *execRecorder) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	panic("not expected")
}

func (r *execRecorder) Prepare(string) (*sql.Stmt, error) {
	panic("not expected")
}

func newRecordingBuilder(driver string, newBuilder BuilderFunc) (Builder, *execRecorder) {
	r := &execRecorder{}
	db := NewFromDB(nil, driver)
	b := newBuilder(db, r)
	db.Builder = b
	return b, r
}

type upsertCustomer struct {
	ID      int
	Email   string
	Name    string
	Created string
}

func (upsertCustomer) TableName() string { return "customer" }

type upsertOrderItem struct {
	OrderID  int `db:"pk"`
	ItemID   int `db:"pk"`
	Quantity int
}

type upsertNoPK struct {
	Name string
}

func TestModelQuery_Upsert(t *testing.T) {
	newCustomer := func(id int) *upsertCustomer {
		return &upsertCustomer{ID: id, Email: "e@example.com", Name: "n", Created: "c"}
	}

	t.Run("mysql", func(t *testing.T) {
		b, r := newRecordingBuilder("mysql", NewMysqlBuilder)

		// existing PK: PK inserted and used as the target, never updated
		assert.Nil(t, b.Model(newCustomer(5)).Upsert())
		// Exclude applies to both the insertion and the update
		assert.Nil(t, b.Model(newCustomer(5)).Exclude("Created").Upsert())
		// zero auto-increment PK is left out of the insertion
		assert.Nil(t, b.Model(newCustomer(0)).Upsert())
		// attrs select the fields; the PK is added regardless
		assert.Nil(t, b.Model(newCustomer(5)).Upsert("Name"))
		// Exclude does not apply to the PK
		assert.Nil(t, b.Model(newCustomer(5)).Exclude("ID", "Created").Upsert())

		assert.Equal(t, []string{
			"INSERT INTO `customer` (`created`, `email`, `id`, `name`) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE `created`=VALUES(`created`), `email`=VALUES(`email`), `name`=VALUES(`name`)",
			"INSERT INTO `customer` (`email`, `id`, `name`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `email`=VALUES(`email`), `name`=VALUES(`name`)",
			"INSERT INTO `customer` (`created`, `email`, `name`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `created`=VALUES(`created`), `email`=VALUES(`email`), `name`=VALUES(`name`)",
			"INSERT INTO `customer` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `name`=VALUES(`name`)",
			"INSERT INTO `customer` (`email`, `id`, `name`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `email`=VALUES(`email`), `name`=VALUES(`name`)",
		}, r.sql)
		assert.Equal(t, []interface{}{"c", "e@example.com", 5, "n"}, r.args[0])
		assert.Equal(t, []interface{}{5, "n"}, r.args[3])
	})

	t.Run("pgsql", func(t *testing.T) {
		b, r := newRecordingBuilder("postgres", NewPgsqlBuilder)
		assert.Nil(t, b.Model(newCustomer(5)).Exclude("Created").Upsert())
		assert.Nil(t, b.Model(newCustomer(0)).Exclude("Created").Upsert())
		assert.Nil(t, b.Model(&upsertOrderItem{OrderID: 1, ItemID: 0, Quantity: 3}).Upsert())
		assert.Equal(t, []string{
			`INSERT INTO "customer" ("email", "id", "name") VALUES ($1, $2, $3) ON CONFLICT ("id") DO UPDATE SET "email"=EXCLUDED."email", "name"=EXCLUDED."name"`,
			`INSERT INTO "customer" ("email", "name") VALUES ($1, $2) ON CONFLICT ("id") DO UPDATE SET "email"=EXCLUDED."email", "name"=EXCLUDED."name"`,
			// composite PK: all PK columns are the target, a zero value is kept
			`INSERT INTO "upsert_order_item" ("item_id", "order_id", "quantity") VALUES ($1, $2, $3) ON CONFLICT ("item_id", "order_id") DO UPDATE SET "quantity"=EXCLUDED."quantity"`,
		}, r.sql)
	})

	t.Run("sqlite", func(t *testing.T) {
		b, r := newRecordingBuilder("sqlite", NewSqliteBuilder)
		assert.Nil(t, b.Model(newCustomer(5)).Upsert("Name"))
		assert.Equal(t, []string{
			"INSERT INTO `customer` (`id`, `name`) VALUES (?, ?) ON CONFLICT (`id`) DO UPDATE SET `name`=EXCLUDED.`name`",
		}, r.sql)
	})

	t.Run("errors", func(t *testing.T) {
		b, r := newRecordingBuilder("mysql", NewMysqlBuilder)
		// nothing left to update besides the PK
		assert.NotNil(t, b.Model(newCustomer(5)).Upsert("ID"))
		assert.Equal(t, MissingPKError, b.Model(&upsertNoPK{Name: "n"}).Upsert())
		var a int
		assert.NotNil(t, b.Model(&a).Upsert())
		assert.Empty(t, r.sql)

		// databases without Builder.Upsert
		b, r = newRecordingBuilder("mssql", NewMssqlBuilder)
		assert.NotNil(t, b.Model(newCustomer(5)).Upsert())
		b2, r2 := newRecordingBuilder("oci8", NewOciBuilder)
		assert.NotNil(t, b2.Model(newCustomer(5)).Upsert())
		assert.Empty(t, r.sql)
		assert.Empty(t, r2.sql)
	})
}
