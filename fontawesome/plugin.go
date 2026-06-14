package fontawesome

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/downballot/ui/blazarapp"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Config is the configuration for the FontAwesome plugin.
type Config struct {
	Location string
	Minify   bool
}

// plugin is the implementation of the FontAwesome plugin.
type plugin struct {
	config Config
}

var _ blazarapp.Plugin = (*plugin)(nil)

// NewPlugin creates a new FontAwesome plugin.
func NewPlugin(config Config) blazarapp.Plugin {
	return &plugin{
		config: config,
	}
}

// Register registers the plugin against the go-app handler and the HTTP mux.
func (p *plugin) Register(handler *app.Handler, mux *http.ServeMux) {
	mux.Handle(p.config.Location, http.StripPrefix(p.config.Location, p.httpHandler()))
	handler.Styles = append(handler.Styles, p.cssFilenames(p.config.Location)...)
}

// cssFilenames returns the CSS filenames for the plugin.
func (p *plugin) cssFilenames(prefix string) []string {
	var cssFiles []string
	if p.config.Minify {
		cssFiles = append(cssFiles, "css/all.min.css")
	} else {
		cssFiles = append(cssFiles, "css/all.css")
	}
	slices.Sort(cssFiles)
	for i, filename := range cssFiles {
		cssFiles[i] = filepath.Join(prefix, filename)
	}
	return cssFiles
}

// httpHandler returns the HTTP handler for the plugin.
func (p *plugin) httpHandler() http.Handler {
	newFS, err := fs.Sub(embeddedFS, "embedded")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(newFS))
}
