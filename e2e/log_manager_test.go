package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLogManagerLoadsSeededLogs(t *testing.T) {
	body := waitTable(t, newTab(t), adminURL("log-manager"))
	if !strings.Contains(body, seedLogMessage) {
		t.Fatal("expected seeded 'E2E Seed Log' rows in log manager")
	}
}

// TestLogManagerFilterModal verifies the "Filters" button opens the
// filter modal.
func TestLogManagerFilterModal(t *testing.T) {
	ctx := newTab(t)

	// Waiting for the second rendered table row guarantees Vue has
	// mounted; the buttons are only bound after that.
	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("log-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("log manager did not load: %v", err)
	}

	// The "Filters" button is the only .btn-primary on the filter bar
	// before the modal opens.
	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-primary`, chromedp.ByQuery),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("filter modal did not open: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading log manager failed: %v", err)
	}
	if !strings.Contains(body, "Filter Logs") {
		t.Fatal("expected 'Filter Logs' modal title")
	}
}

// TestLogShowContext verifies the per-row eye button opens the context
// modal with the log's details.
func TestLogShowContext(t *testing.T) {
	ctx := newTab(t)

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("log-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("log manager did not load: %v", err)
	}

	// The eye button is .btn-info in each row's Actions cell.
	err = chromedp.Run(ctx,
		chromedp.Click(`table tbody tr .btn-info`, chromedp.ByQuery),
		chromedp.WaitReady(`.modal .modal-title`, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("context modal did not open: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading context modal failed: %v", err)
	}
	if !strings.Contains(body, "Log Context") {
		t.Fatal("expected 'Log Context' modal title")
	}
	if !strings.Contains(body, seedLogMessage) {
		t.Fatal("expected seeded log message in context modal")
	}
}

// TestLogDeleteFlow verifies the per-row delete button removes a log
// after the SweetAlert2 confirmation.
func TestLogDeleteFlow(t *testing.T) {
	ctx := newTab(t)

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("log-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("log manager did not load: %v", err)
	}

	// Capture the first row's message so we can assert it is gone
	// after the delete + refresh.
	var firstMessage string
	err = chromedp.Run(ctx,
		chromedp.Text(`table tbody tr td:nth-of-type(4)`, &firstMessage, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("reading first row failed: %v", err)
	}

	// The trash button is .btn-danger in each row's Actions cell; the
	// confirmation is a SweetAlert2 dialog (.swal2-confirm).
	err = chromedp.Run(ctx,
		chromedp.Click(`table tbody tr .btn-danger`, chromedp.ByQuery),
		chromedp.WaitReady(`.swal2-confirm`, chromedp.ByQuery),
		chromedp.Click(`.swal2-confirm`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("delete flow failed: %v", err)
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.Text("body", &body)); err != nil {
		t.Fatalf("reading log manager failed: %v", err)
	}
	if strings.Contains(body, firstMessage) {
		t.Fatalf("expected log %q to be deleted", firstMessage)
	}
}
