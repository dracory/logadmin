# e2e

End-to-end tests that drive the logadmin panel in a real headless Chrome
via [chromedp](https://github.com/chromedp/chromedp) (CDP). Unlike go-rod,
chromedp has no `leakless.exe` helper, so it avoids the Windows Defender
false positive entirely.

## How it works

- `TestMain` mounts the admin on `httptest.NewServer` (in-process, random
  port) backed by a seeded in-memory SQLite logstore via
  `testutils.InitStore` — no binary build needed. 30 log entries are
  seeded across info/warning/error levels.
- One shared browser is launched via `chromedp.NewExecAllocator` for the
  whole run; each test opens a fresh tab with a 30s deadline (`newTab`).
- Chrome binary resolution order: `CHROME_BIN` env → Windows Playwright
  `chromium_headless_shell` → WSL-side Windows shell → Linux Playwright
  shell → chromedp's default lookup (Chrome on PATH / install dirs).

## Run

```bash
go test -v ./e2e/     # or: task e2e
```

Note: on WSL, run tests with the Windows Go toolchain (`go.exe`) —
Windows-side Chrome binds 127.0.0.1, which WSL cannot reach (no
WSL→Windows localhost forwarding). The Linux Playwright shell also
requires system libs (libnspr4 etc.) that may be missing.

## Coverage

- `log-manager` — seeded rows + filter modal + context modal + delete
  flow (the only registered controller)

## Gotchas learned

- Prefer `chromedp.WaitReady` over `chromedp.WaitVisible`: on some pages
  `WaitVisible` polls forever even though the DOM is fully loaded.
- Routing is query-param based: `?controller=log-manager`
  (see `shared/consts.go`).
- The log manager renders via Vue after an AJAX fetch — wait for the
  second `tbody tr` before asserting (v-if/v-else and v-for mean the
  data rows only exist after Vue mounts).
- `v-model` is compiled away by Vue — it does NOT appear in the rendered
  DOM. Match form inputs by `type` or `placeholder`, never by the
  `v-model` attribute.
- Modals are Vue-rendered `v-if` blocks (not Bootstrap JS modals) — they
  appear as `.modal.fade.show` once `showFilterModal`/`showContextModal`
  flip to true; wait on `.modal .modal-title`.
- Delete confirmations use SweetAlert2 — confirm by clicking
  `.swal2-confirm`.
- chromedp v0.23 rewrote the API (generic `Run[T]`); this suite pins the
  stable v0.14 line with matching `cdproto`/`go-json-experiment` versions.
