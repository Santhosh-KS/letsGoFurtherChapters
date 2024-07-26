package models

import (
	"database/sql"
	"os"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	// Establis a sql.DB conection pool for our test database. Because our setup and teardown scripts contains multiple SQL statements, we need to use the "multiStatemens=true: pareameter in our DSN. This instructs our MYSQL database driver to support executing multiple SQL statements in one db.Exec() cll"

	db, err := sql.Open("mysql", "test_web:pass@/test_snippetbox?parseTime-true&multistatements=true")
	if err != nil {
		t.Fatal(err)
	}

	// Read the setup SQL script from the file and execute the statements, closing the connection pool and calling t.Fatal() in the event of an error.

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	// Use t.Cleanup() to register a fucntion*which will automatically be called by Go when the current test (or sub-test) which callsnewTestDB() has finished*. In this function we read and execute the teardown script, and close the database connection pool.

	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")

		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
	})
	return db
}
