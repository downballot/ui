package blazarapp

import (
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Plugin is a plugin for the Blazar app.
type Plugin interface {
	// Register the plugin against the go-app handler and the HTTP mux.
	Register(handler *app.Handler, mux *http.ServeMux)
}
