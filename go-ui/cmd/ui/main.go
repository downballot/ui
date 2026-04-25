package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/downballot/ui/go-ui/component"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// hello is a component that displays a simple "Hello World!". A component is a
// customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type hello struct {
	app.Compo
}

func (h *hello) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		ctx.Navigate("/login")
	}
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (h *hello) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Hello, world 1!"),
		app.A().Href("/login").Text("Login"),
	)
}

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	{
		logLevel := "info"
		if value, ok := os.LookupEnv("LOG_LEVEL"); ok {
			logLevel = value
		}
		slogConfig := slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					source, _ := a.Value.Any().(*slog.Source)
					if source != nil {
						source.File = filepath.Base(source.File)
					}
				}
				return a
			},
		}
		switch strings.ToLower(logLevel) {
		case "debug":
			slogConfig.Level = slog.LevelDebug
		case "info":
			slogConfig.Level = slog.LevelInfo
		case "warn":
			slogConfig.Level = slog.LevelWarn
		case "error":
			slogConfig.Level = slog.LevelError
		}
		slog.SetDefault(slog.New(slog.NewMultiHandler(
			slog.NewTextHandler(os.Stderr, &slogConfig),
			&JavascriptConsoleLogger{},
		)))
	}

	// The first thing to do is to associate the hello component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", func() app.Composer { return &hello{} })
	app.Route("/login", func() app.Composer { return &component.LoginPage{} })

	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	mux := http.NewServeMux()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	mux.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		reverseProxy := &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(&url.URL{
					Scheme: "http",
					Host:   "localhost:8888",
				})
			},
		}
		slog.InfoContext(r.Context(), "Reverse proxy request", "method", r.Method, "url", r.URL.String())
		reverseProxy.ServeHTTP(w, r)
	})

	wrapper := http.NewServeMux()
	wrapper.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "[in] Request", "method", r.Method, "url", r.URL.String())
		wrappedResponseWriter := &ResponseWriter{
			writer: w,
		}
		wrappedResponseWriter.Info.StatusCode = http.StatusOK
		mux.ServeHTTP(wrappedResponseWriter, r)
		slog.InfoContext(r.Context(), "[out] Request", "method", r.Method, "url", r.URL.String(), "code", wrappedResponseWriter.Info.StatusCode)
	})

	if err := http.ListenAndServe(":8000", wrapper); err != nil {
		log.Fatal(err)
	}
}

type ResponseWriter struct {
	writer http.ResponseWriter

	Info struct {
		StatusCode   int
		BytesWritten uint64
	}
}

var _ http.ResponseWriter = (*ResponseWriter)(nil)

func (w *ResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *ResponseWriter) Write(contents []byte) (int, error) {
	count, err := w.writer.Write(contents)
	if err != nil {
		return count, err
	}
	w.Info.BytesWritten += uint64(count)
	return count, err
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.Info.StatusCode = statusCode
	w.writer.WriteHeader(statusCode)
}

type JavascriptConsoleLogger struct {
	attrs []slog.Attr
	group string
}

var _ slog.Handler = (*JavascriptConsoleLogger)(nil)

func (h *JavascriptConsoleLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}
func (h *JavascriptConsoleLogger) Handle(ctx context.Context, record slog.Record) error {
	if app.IsServer {
		return nil
	}
	app.Log(record.Message)
	return nil
}
func (h *JavascriptConsoleLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := JavascriptConsoleLogger{
		group: h.group,
	}
	newHandler.attrs = append(newHandler.attrs, h.attrs...)
	newHandler.attrs = append(newHandler.attrs, attrs...)

	return &newHandler
}
func (h *JavascriptConsoleLogger) WithGroup(name string) slog.Handler {
	newHandler := JavascriptConsoleLogger{
		group: name,
	}
	newHandler.attrs = append(newHandler.attrs, h.attrs...)

	return &newHandler
}
