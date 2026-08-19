package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/logstore"
)

// UiConfig holds the dependencies passed to subcontroller UI factories.
// The Layout function uses an anonymous struct for options to match
// shopadmin/blogadmin exactly, allowing consumers to reuse their
// existing layout function for logadmin.
type UiConfig struct {
	Store  logstore.StoreInterface
	Logger *slog.Logger
	Layout func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
