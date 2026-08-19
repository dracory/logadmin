package shared

import "net/http"

// Links provides URL helpers for logadmin controllers.
// The base URL is read from request context (injected by Handle()),
// not hardcoded. This follows the shopadmin/blogadmin pattern.
type Links struct {
	baseURL string
}

// NewLinks creates a Links helper with the given base URL.
// If baseURL is empty, defaults to "/admin/logs".
func NewLinks(baseURL string) *Links {
	if baseURL == "" {
		baseURL = "/admin/logs"
	}
	return &Links{baseURL: baseURL}
}

// NewLinksFromRequest creates a Links helper using the log admin URL
// from the request context.
func NewLinksFromRequest(r *http.Request) *Links {
	return NewLinks(LogAdminURL(r))
}

// LogManager builds the URL for the log manager controller
func (l *Links) LogManager(params map[string]string) string {
	return l.url(CONTROLLER_LOG_MANAGER, params)
}

// url builds a URL for the given controller. The params map is copied
// before mutation (does not modify caller's map).
func (l *Links) url(controller string, params map[string]string) string {
	return URL(l.baseURL, controller, params)
}
