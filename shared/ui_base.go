package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/logstore"
)

// UiBase is a base struct that implements shared.UiInterface.
// Subcontroller ui structs can embed this to get the Store(),
// Logger(), and Layout() methods for free, following the
// shopadmin/blogadmin pattern.
type UiBase struct {
	StoreField  logstore.StoreInterface
	LoggerField *slog.Logger
	LayoutField func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}

func (u UiBase) Store() logstore.StoreInterface { return u.StoreField }
func (u UiBase) Logger() *slog.Logger           { return u.LoggerField }

func (u UiBase) Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	return u.LayoutField(w, r, webpageTitle, webpageHtml, options)
}

// NewUiBase creates a UiBase from a UiConfig
func NewUiBase(config UiConfig) UiBase {
	return UiBase{
		StoreField:  config.Store,
		LoggerField: config.Logger,
		LayoutField: config.Layout,
	}
}
