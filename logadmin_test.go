package logadmin

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dracory/logadmin/testutils"
	_ "modernc.org/sqlite"
)

func TestNew_ValidOptions(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	options := AdminOptions{
		Store:          store,
		Logger:         slog.New(slog.NewTextHandler(os.Stderr, nil)),
		AdminHomeURL:   "/admin",
		LogAdminURL:    "/admin/logs",
		FileManagerURL: "/admin/files",
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if a == nil {
		t.Errorf("Expected admin to be created, got nil")
	}
}

func TestNew_MissingStore(t *testing.T) {
	options := AdminOptions{
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	a, err := New(options)
	if err == nil {
		t.Errorf("Expected error when store is missing")
	}
	if a != nil {
		t.Errorf("Expected nil admin when store is missing")
	}
	if !strings.Contains(err.Error(), ErrStoreRequired.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", ErrStoreRequired.Error(), err.Error())
	}
}

func TestNew_MissingLogger(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	options := AdminOptions{
		Store: store,
	}
	a, err := New(options)
	if err == nil {
		t.Errorf("Expected error when logger is missing")
	}
	if a != nil {
		t.Errorf("Expected nil admin when logger is missing")
	}
	if !strings.Contains(err.Error(), ErrLoggerRequired.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", ErrLoggerRequired.Error(), err.Error())
	}
}

func TestNew_Defaults(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	options := AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Verify defaults were set
	admin := a.(*admin)
	if admin.adminHomeURL != "/admin" {
		t.Errorf("Expected default adminHomeURL '/admin', got '%s'", admin.adminHomeURL)
	}
	if admin.logAdminURL != "/admin/logs" {
		t.Errorf("Expected default logAdminURL '/admin/logs', got '%s'", admin.logAdminURL)
	}
}

func TestHandle_LogManagerController(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Log Manager") {
		t.Errorf("Expected body to contain 'Log Manager'")
	}
}

func TestHandle_UnknownControllerFallsBackToLogManager(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/logs?controller=nonexistent", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Log Manager") {
		t.Errorf("Expected fallback to log manager page with 'Log Manager'")
	}
}

func TestHandle_AuthUserIDRedirect(t *testing.T) {
	store, err := testutils.InitStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}

	a, err := New(AdminOptions{
		Store:      store,
		Logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
		AuthUserID: func(r *http.Request) string { return "" },
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect status 303, got %d", rr.Code)
	}
}
