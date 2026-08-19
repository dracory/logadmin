# logadmin

<img src="https://opengraph.githubassets.com/dracory/logadmin" />

[![Tests Status](https://github.com/dracory/logadmin/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/logadmin/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/logadmin)](https://goreportcard.com/report/github.com/dracory/logadmin)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/logadmin)](https://pkg.go.dev/github.com/dracory/logadmin)

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at [https://www.gnu.org/licenses/agpl-3.0.en.html](https://www.gnu.org/licenses/agpl-3.0.txt)

For commercial use, please use my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Introduction

Admin interface for [`github.com/dracory/logstore`](https://github.com/dracory/logstore).
Provides a ready-to-use admin panel for viewing, filtering, and deleting
application logs.

Modeled after [`github.com/dracory/shopadmin`](https://github.com/dracory/shopadmin)
and [`github.com/dracory/blogadmin`](https://github.com/dracory/blogadmin)
— same folder-per-controller pattern, same `UiConfig`/`UiBase` conventions.

## Features

- **Log viewing** — paginated table with sortable columns
- **Filtering** — filter by level, message contains/not-contains, context
  contains/not-contains, and date range
- **Bulk actions** — delete selected logs or delete all logs matching
  current filters
- **Context viewer** — view full log entry including context data
- **Shareable URLs** — filter and pagination state preserved in URL parameters
- **Custom layouts** — bring your own layout via `FuncLayout`
- **Bootstrap + Vue CDN** — default UI works out of the box

## Installation

```bash
go get github.com/dracory/logadmin
```

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/dracory/logadmin"
    "github.com/dracory/logstore"
)

func main() {
    store, err := logstore.NewStore(logstore.NewStoreOptions{
        DB:                 yourDB,
        LogTableName:       "log",
        AutomigrateEnabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    admin, err := logadmin.New(logadmin.AdminOptions{
        Store:        store,
        Logger:       slog.New(slog.NewTextHandler(os.Stderr, nil)),
        AdminHomeURL: "/admin",
        LogAdminURL:  "/admin/logs",
    })
    if err != nil {
        log.Fatal(err)
    }

    http.Handle("/admin/logs", http.HandlerFunc(admin.Handle))
    http.ListenAndServe(":8080", nil)
}
```

See [`example/`](example/) for a complete runnable server with
in-memory SQLite and seed data.

## Integration with a Router

`logadmin.AdminInterface` exposes `Handle(w, r)`, which is an
`http.HandlerFunc`-compatible method. Wire it into any router that
accepts standard `http.Handler`:

```go
// stdlib
mux.Handle("/admin/logs", http.HandlerFunc(admin.Handle))

// github.com/dracory/rtr
route := rtr.NewRoute().
    SetName("Admin > Logs").
    SetPath("/admin/logs").
    SetHTMLHandler(admin.Handle)
```

## Custom Layout

By default, logadmin renders a bare-bones HTML page with Bootstrap and
Vue from CDN. To embed the admin inside your own layout (branding, menus,
etc.), provide `FuncLayout`:

```go
admin, _ := logadmin.New(logadmin.AdminOptions{
    Store:  store,
    Logger: logger,
    FuncLayout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
        Styles     []string
        StyleURLs  []string
        Scripts    []string
        ScriptURLs []string
    }) string {
        return myLayout(w, r, title, body, opts)
    },
})
```

The anonymous struct matches shopadmin/blogadmin exactly, so you can reuse
your existing shopadmin/blogadmin layout function.

## Testing

```bash
go test ./...
```

Tests use an in-memory SQLite database via `modernc.org/sqlite` — no
external services required.
