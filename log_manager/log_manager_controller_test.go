package log_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dracory/logadmin/shared"
	"github.com/dracory/logadmin/testutils"
	"github.com/dracory/logstore"
	_ "modernc.org/sqlite"
)

func newTestUI(t *testing.T) *ui {
	t.Helper()
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	return &ui{UiBase: shared.NewUiBase(shared.UiConfig{
		Store:  store,
		Logger: logger(),
		Layout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			return body
		},
	})}
}

func logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func seedLogs(t *testing.T, store logstore.StoreInterface, ctx context.Context) {
	t.Helper()
	l1 := logstore.NewLog()
	l1.SetMessage("Test Message 1")
	l1.SetLevel("info")
	if err := store.LogCreate(ctx, l1); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	l2 := logstore.NewLog()
	l2.SetMessage("Test Message 2")
	l2.SetLevel("error")
	if err := store.LogCreate(ctx, l2); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}
}

func TestLogManagerController_RenderPage(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyAdminHomeURL, "/admin"))
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyLogAdminURL, "/admin/logs"))
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyEndpoint, "/admin/logs"))

	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "Log Manager") {
		t.Error("expected Log Manager in response")
	}
}

func TestLogManagerController_HandleLoadLogs(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	loadData := map[string]any{
		"page":     0,
		"per_page": 10,
	}
	body, _ := json.Marshal(loadData)
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?action="+actionLoadLogs, bytes.NewBuffer(body))
	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "success") {
		t.Error("expected success in response")
	}
	if !strings.Contains(resp, "Test Message 1") {
		t.Error("expected Test Message 1 in response")
	}
	if !strings.Contains(resp, "Test Message 2") {
		t.Error("expected Test Message 2 in response")
	}
}

func TestLogManagerController_HandleLogShowContext(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	logs, _ := u.Store().LogList(ctx, logstore.LogQuery())
	logID := logs[0].GetID()

	showData := map[string]string{
		"log_id": logID,
	}
	body, _ := json.Marshal(showData)
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?action="+actionLogShowContext, bytes.NewBuffer(body))
	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "success") {
		t.Error("expected success in response")
	}
	if !strings.Contains(resp, "context") {
		t.Error("expected context in response")
	}
}

func TestLogManagerController_HandleLogDelete(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	logs, _ := u.Store().LogList(ctx, logstore.LogQuery())
	logID := logs[0].GetID()

	deleteData := map[string]string{
		"log_id": logID,
	}
	body, _ := json.Marshal(deleteData)
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?action="+actionLogDelete, bytes.NewBuffer(body))
	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "success") {
		t.Error("expected success in response")
	}

	// Verify deletion
	l, _ := u.Store().LogFindByID(ctx, logID)
	if l != nil {
		t.Error("expected log to be nil after deletion")
	}
}

func TestLogManagerController_HandleLogDeleteSelected(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	// Create a log to delete
	l := logstore.NewLog()
	l.SetMessage("Delete selected")
	if err := u.Store().LogCreate(ctx, l); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	logs, _ := u.Store().LogList(ctx, logstore.LogQuery())
	var logID string
	for _, log := range logs {
		if log.GetMessage() == "Delete selected" {
			logID = log.GetID()
			break
		}
	}

	deleteData := map[string][]string{
		"bulk_log_ids": {logID},
	}
	body, _ := json.Marshal(deleteData)
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?action="+actionLogDeleteSelected, bytes.NewBuffer(body))
	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "success") {
		t.Error("expected success in response")
	}

	// Verify deletion
	lFound, _ := u.Store().LogFindByID(ctx, logID)
	if lFound != nil {
		t.Error("expected log to be nil after deletion")
	}
}

func TestLogManagerController_HandleLogDeleteAll(t *testing.T) {
	u := newTestUI(t)
	ctx := context.Background()
	seedLogs(t, u.Store(), ctx)

	// Create another log
	l3 := logstore.NewLog()
	l3.SetMessage("Delete me")
	if err := u.Store().LogCreate(ctx, l3); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	deleteData := map[string]string{}
	body, _ := json.Marshal(deleteData)
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?action="+actionLogDeleteAll, bytes.NewBuffer(body))
	resp := u.Handler(httptest.NewRecorder(), req)
	if !strings.Contains(resp, "success") {
		t.Error("expected success in response")
	}

	// Verify deletion
	count, _ := u.Store().LogCount(ctx, logstore.LogQuery())
	if int(count) != 0 {
		t.Errorf("expected 0 logs, got %d", count)
	}
}

func TestLogManagerController_RendersVueApp(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	config := shared.UiConfig{
		Store:  store,
		Logger: logger(),
		Layout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			// Return the body with the script URLs appended so we can
			// assert on CDN includes.
			out := body
			for _, u := range opts.ScriptURLs {
				out += `<script src="` + u + `"></script>`
			}
			return out
		},
	}

	controller := UI(config)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyAdminHomeURL, "/admin"))
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyLogAdminURL, "/admin/logs"))
	req = req.WithContext(context.WithValue(req.Context(), shared.KeyEndpoint, "/admin/logs"))

	controller.LogManager(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	html := rr.Body.String()
	// Page should render Vue app mount point
	if !strings.Contains(html, "logs-app") {
		t.Error("expected logs-app in HTML")
	}
	// Page should include Vue CDN
	if !strings.Contains(html, "vue.global.js") {
		t.Error("expected Vue CDN script in HTML")
	}
	// Page should include SweetAlert2
	if !strings.Contains(html, "sweetalert2") {
		t.Error("expected SweetAlert2 in HTML")
	}
}

func TestLogManagerController_LoadLogsAction(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	// Seed some logs
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		l := logstore.NewLog()
		l.SetMessage("test log")
		l.SetLevel("info")
		store.LogCreate(ctx, l)
	}

	config := shared.UiConfig{
		Store:  store,
		Logger: logger(),
		Layout: func(w http.ResponseWriter, r *http.Request, title, body string, opts struct {
			Styles     []string
			StyleURLs  []string
			Scripts    []string
			ScriptURLs []string
		}) string {
			return body
		},
	}

	controller := UI(config)

	queryParams := url.Values{}
	queryParams.Set("action", "load-logs")

	requestBody := map[string]any{
		"page":       0,
		"per_page":   100,
		"sort_order": "desc",
		"sort_by":    "time",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/logs?"+queryParams.Encode(), bytes.NewBuffer(bodyBytes))
	controller.LogManager(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	body := rr.Body.String()
	// Should return JSON response
	if !strings.Contains(body, `"status"`) {
		t.Error("expected JSON response with status field")
	}
	if !strings.Contains(body, `"logs"`) {
		t.Error("expected JSON response with logs field")
	}
}
