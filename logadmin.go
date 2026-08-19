// Package logadmin provides a standalone log admin interface following
// the folder-per-controller pattern. Each controller is in its own
// subfolder and handles its own views and AJAX data.
//
// This module is modeled on github.com/dracory/shopadmin and
// github.com/dracory/blogadmin.
package logadmin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/logadmin/log_manager"
	"github.com/dracory/logadmin/shared"
	"github.com/dracory/logstore"
	"github.com/dracory/req"
)

// AdminOptions contains all dependencies and configuration for the log admin.
//
// Store and Logger replace the in-repo version's Registry field (which
// was app.AppInterface). This matches the shopadmin/blogadmin convention
// where stores are passed directly.
//
// FuncLayout is an optional function to render the admin interface
// inside your own layout (branding, menus, etc.). It receives the
// request and response writer so the host project can access request
// context (auth user, locale, etc.) when rendering the layout.
// If nil, a default bare-bones HTML page is used (Bootstrap + Vue CDN).
// Uses anonymous struct to match shopadmin/blogadmin exactly, so
// consumers can reuse their shopadmin/blogadmin layout function for
// logadmin.
type AdminOptions struct {
	// Store is the logstore.StoreInterface (required)
	Store logstore.StoreInterface

	// Logger is required
	Logger *slog.Logger

	// FuncLayout is an optional function to render the admin interface
	// inside your own layout (branding, menus, etc.). It receives the
	// request and response writer so the host project can access request
	// context (auth user, locale, etc.) when rendering the layout.
	FuncLayout func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string

	// AdminHomeURL is the URL for the admin home page (default: "/admin")
	AdminHomeURL string

	// LogAdminURL is the base URL for log admin (default: "/admin/logs")
	LogAdminURL string

	// FileManagerURL is the URL for the file manager (optional)
	FileManagerURL string

	// AuthUserID returns the authenticated user ID from the request.
	// If it returns "", the user is treated as unauthenticated.
	AuthUserID func(r *http.Request) string
}

// AdminInterface defines the interface for the log admin
type AdminInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// admin implements AdminInterface
type admin struct {
	store          logstore.StoreInterface
	logger         *slog.Logger
	funcLayout     func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
	adminHomeURL   string
	logAdminURL    string
	fileManagerURL string
	authUserID     func(r *http.Request) string
	routes         map[string]func(w http.ResponseWriter, r *http.Request)
}

// New creates a new log admin instance.
// Returns ErrStoreRequired if Store is nil, ErrLoggerRequired if Logger is nil.
func New(opts AdminOptions) (AdminInterface, error) {
	if opts.Store == nil {
		return nil, ErrStoreRequired
	}
	if opts.Logger == nil {
		return nil, ErrLoggerRequired
	}

	// Set defaults
	if opts.AdminHomeURL == "" {
		opts.AdminHomeURL = "/admin"
	}
	if opts.LogAdminURL == "" {
		opts.LogAdminURL = "/admin/logs"
	}

	a := &admin{
		store:          opts.Store,
		logger:         opts.Logger,
		funcLayout:     opts.FuncLayout,
		adminHomeURL:   opts.AdminHomeURL,
		logAdminURL:    opts.LogAdminURL,
		fileManagerURL: opts.FileManagerURL,
		authUserID:     opts.AuthUserID,
	}

	// Build routes once at construction time
	a.routes = a.buildRoutes()

	return a, nil
}

// Handle processes all log admin requests.
// Config values are injected into the request context (following the
// shopadmin/blogadmin pattern). Route lookup is map-based.
func (a *admin) Handle(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if a.authUserID != nil && a.authUserID(r) == "" {
		http.Redirect(w, r, a.adminHomeURL, http.StatusSeeOther)
		return
	}

	// Inject config into request context (like shopadmin/blogadmin)
	ctx := context.WithValue(r.Context(), shared.KeyEndpoint, r.URL.Path)
	ctx = context.WithValue(ctx, shared.KeyAdminHomeURL, a.adminHomeURL)
	ctx = context.WithValue(ctx, shared.KeyLogAdminURL, a.logAdminURL)
	ctx = context.WithValue(ctx, shared.KeyFileManagerURL, a.fileManagerURL)
	r = r.WithContext(ctx)

	// Map-based route lookup
	controller := req.GetStringTrimmed(r, "controller")
	if controller == "" {
		controller = shared.CONTROLLER_LOG_MANAGER
	}

	handler, ok := a.routes[controller]
	if !ok {
		handler = a.routes[shared.CONTROLLER_LOG_MANAGER]
	}

	handler(w, r)
}

// buildRoutes creates the handler dispatch map once at construction time.
func (a *admin) buildRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	uiConfig := shared.UiConfig{
		Store:  a.store,
		Logger: a.logger,
		Layout: a.render,
	}

	return map[string]func(w http.ResponseWriter, r *http.Request){
		shared.CONTROLLER_LOG_MANAGER: func(w http.ResponseWriter, r *http.Request) {
			log_manager.UI(uiConfig).LogManager(w, r)
		},
	}
}

// render wraps content in the layout. If FuncLayout is provided and
// returns non-empty, it is used; otherwise the default shared.Layout
// is used (following the shopadmin/blogadmin pattern).
//
// When FuncLayout is set, the default shared.Layout is NOT computed
// (avoids wasted work).
func (a *admin) render(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	// If a custom layout is provided, try it first
	if a.funcLayout != nil {
		custom := a.funcLayout(w, r, webpageTitle, webpageHtml, options)
		if custom != "" {
			if w != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(custom))
				return ""
			}
			return custom
		}
	}

	webpage := shared.Layout(w, r, webpageTitle, webpageHtml, options)

	if w != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(webpage))
		return ""
	}

	return webpage
}
