package testutils

import (
	"database/sql"
	"os"
	"strings"

	"github.com/dracory/logstore"
)

// InitStore creates an in-memory logstore for testing.
//
// Note: The SQLite driver (modernc.org/sqlite) must be imported with a
// side-effect import in _test.go files of the consuming package, NOT here.
// This avoids double driver registration panics when a host project that
// already imports modernc.org/sqlite also imports this testutils package.
//
// Example (in your _test.go file):
//
//	import (
//	    "github.com/dracory/logadmin/testutils"
//	    _ "modernc.org/sqlite"
//	)
func InitStore(filepath string) (logstore.StoreInterface, error) {
	db, err := initDB(filepath)
	if err != nil {
		return nil, err
	}

	store, err := logstore.NewStore(logstore.NewStoreOptions{
		DB:                 db,
		LogTableName:       "log",
		AutomigrateEnabled: true,
	})
	if err != nil {
		return nil, err
	}

	return store, nil
}

func initDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" {
		err := os.Remove(filepath)
		if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
			return nil, err
		}
	}

	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
