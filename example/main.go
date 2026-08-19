// Example logadmin server.
//
// Boots an HTTP server with an in-memory SQLite database and mounts the
// logadmin panel at /admin/logs. No external services required.
//
// Run from the logadmin module root:
//
//	go run ./example
//
// Then open http://localhost:8080/ in your browser.
package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/dracory/logadmin"
	"github.com/dracory/logstore"
	_ "modernc.org/sqlite"
)

const (
	addr      = ":8080"
	dbFile    = ":memory:"
	adminURL  = "/admin/logs"
	homeURL   = "/admin"
	filesURL  = "/admin/files"
	dbDriver  = "sqlite"
	dsnSuffix = "?parseTime=true"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(dbFile)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := logstore.NewStore(logstore.NewStoreOptions{
		DB:                 db,
		LogTableName:       "log",
		AutomigrateEnabled: true,
	})
	if err != nil {
		logger.Error("failed to create logstore", "err", err)
		os.Exit(1)
	}

	// Seed the in-memory DB with sample data for testing pagination,
	// filtering, and sorting. Skip seeding if a file-based DB is used
	// (persisted data should not be overwritten).
	if dbFile == ":memory:" {
		seedDB(store, logger)
	}

	admin, err := logadmin.New(logadmin.AdminOptions{
		Store:          store,
		Logger:         logger,
		AdminHomeURL:   homeURL,
		LogAdminURL:    adminURL,
		FileManagerURL: filesURL,
		// AuthUserID is intentionally nil so the example is open.
		// Provide a real implementation in production.
	})
	if err != nil {
		logger.Error("failed to create logadmin", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	// logadmin.AdminInterface exposes Handle(w, r) but not ServeHTTP,
	// so wrap it in an http.HandlerFunc.
	mux.Handle(adminURL, http.HandlerFunc(admin.Handle))
	mux.Handle(adminURL+"/", http.HandlerFunc(admin.Handle))

	// Landing page that links into the admin.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, landingHTML, adminURL)
	})

	logger.Info("logadmin example server starting",
		"addr", addr,
		"admin", adminURL,
		"db", dbFile,
	)
	// Print full clickable URLs to stdout (slog escapes URLs, making them
	// unclickable in most terminals).
	fmt.Printf("\n  logadmin example running:\n    Landing:  http://localhost%s/\n    Admin:    http://localhost%s\n\n", portFromAddr(addr), adminURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// openDB opens a SQLite database. For ":memory:" a fresh database is
// created on every run; pass a file path to persist data across restarts.
func openDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" {
		// Start clean so re-runs don't collide with stale schema.
		if err := os.Remove(filepath); err != nil && !strings.Contains(err.Error(), "no such file") {
			return nil, err
		}
	}

	db, err := sql.Open(dbDriver, filepath+dsnSuffix)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// portFromAddr extracts the port from a listen address like ":8080" or
// "127.0.0.1:8080". Falls back to the original addr if parsing fails.
func portFromAddr(addr string) string {
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return addr
	}
	return addr[i:]
}

const landingHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>logadmin example</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body class="bg-light">
  <div class="container py-5">
    <div class="row justify-content-center">
      <div class="col-md-7">
        <div class="card shadow-sm">
          <div class="card-body p-5">
            <h1 class="h3 mb-3">logadmin example</h1>
            <p class="text-muted mb-4">
              Standalone log admin panel running on an in-memory SQLite
              database. Data is reset on every restart.
            </p>
            <a href="%s" class="btn btn-primary">Open Log Admin &rarr;</a>
          </div>
        </div>
      </div>
    </div>
  </div>
</body>
</html>`
