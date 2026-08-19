package shared

import "net/http"

// FileManagerURL returns the file manager URL from request context
func FileManagerURL(r *http.Request) string {
	value, ok := r.Context().Value(KeyFileManagerURL).(string)
	if !ok {
		return ""
	}
	return value
}
