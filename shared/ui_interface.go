package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/logstore"
)

// UiInterface defines the methods every subcontroller UI must implement.
// This follows the shopadmin/blogadmin pattern.
type UiInterface interface {
	Store() logstore.StoreInterface
	Logger() *slog.Logger

	Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
