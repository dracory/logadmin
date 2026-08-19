package shared

import "net/http"

// LogAdminURL returns the log admin base URL from request context
func LogAdminURL(r *http.Request) string {
	value, ok := r.Context().Value(KeyLogAdminURL).(string)
	if !ok {
		return ""
	}
	return value
}
