package myui

import (
	_ "embed"
	"net/http"
)

//go:embed embedded/style.css
var embeddedCSS string

// CSS returns an HTTP handler that serves the CSS file.
func CSS() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(embeddedCSS))
	})
}
