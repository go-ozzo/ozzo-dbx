package dbx_test

import (
	"os"
	"strings"

	dbx "github.com/go-ozzo/ozzo-dbx"
	_ "github.com/go-sql-driver/mysql"
)

var (
	TestDSN     = getTestDSN()
	FixtureFile = "testdata/mysql.sql"
)

func getTestDSN() string {
	if dsn := os.Getenv("DBX_MYSQL_DSN"); dsn != "" {
		return dsn
	}
	return "travis:@/ozzo_dbx_test?parseTime=true"
}

func getDB() *dbx.DB {
	db, err := dbx.Open("mysql", TestDSN)
	if err != nil {
		panic(err)
	}
	return db
}

func getPreparedDB() *dbx.DB {
	db := getDB()
	s, err := os.ReadFile(FixtureFile)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(s), ";")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := db.NewQuery(line).Execute(); err != nil {
			panic(err)
		}
	}
	return db
}
