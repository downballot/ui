package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/demo"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/myui"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	ctx := context.Background()

	godotenv.Load(".env")

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
		var handler slog.Handler
		if app.IsServer {
			//handler = slog.NewTextHandler(os.Stderr, &slogConfig)
			w := os.Stderr
			handler =
				tint.NewHandler(w, &tint.Options{
					NoColor:     !isatty.IsTerminal(w.Fd()),
					AddSource:   slogConfig.AddSource,
					Level:       slogConfig.Level,
					ReplaceAttr: slogConfig.ReplaceAttr,
				})
		} else {
			handler = &JavascriptConsoleLogger{}
		}
		slog.SetDefault(slog.New(slog.NewMultiHandler(
			handler,
		)))
	}

	disableServiceWorker := true
	if value := os.Getenv("DISABLE_SERVICE_WORKER"); value != "" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			slog.ErrorContext(ctx, "Could not parse DISABLE_SERVICE_WORKER", "err", err)
		} else {
			disableServiceWorker = v
		}
	}

	router.Register(ctx,
		router.Route{
			Path:      "/",
			Component: nil,
			Children: []router.Route{
				{
					Path: "/input",
					Component: func() app.Composer {
						return &demo.InputPage{}
					},
				},
			},
		},
	)

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
		Name:        "UI Demo",
		Description: "UI Demo",
		Title:       "UI Demo",
		Styles: []string{
			"/web/main.css",
			"/web/material.css",
			"/web/myui.css",
		},
		RawHeaders: []string{
			`<script src="https://kit.fontawesome.com/a71e001119.js" crossorigin="anonymous"></script>`,
		},
		Env: map[string]string{},
	})
	mux.Handle("/web/material.css", material.CSS())
	mux.Handle("/web/myui.css", myui.CSS())

	slog.InfoContext(ctx, "Disable service worker?", "disableServiceWorker", disableServiceWorker)
	if disableServiceWorker {
		// Disable the service worker by replacing it with an empty one.
		mux.HandleFunc("/app-worker.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Write([]byte(""))
		})
	}

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	if err := http.ListenAndServe(":"+port, wrapper); err != nil {
		slog.ErrorContext(ctx, "Could not start server", "err", err)
		os.Exit(1)
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
	var variables []any
	variables = append(variables, record.Message)
	record.Attrs(func(a slog.Attr) bool {
		variables = append(variables, a.Key, "=", a.Value)
		return true
	})
	switch record.Level {
	case slog.LevelDebug:
		app.Log(variables...) // TODO: Can we do console.debug?
	case slog.LevelInfo:
		app.Log(variables...)
	case slog.LevelWarn:
		app.Log(variables...) // TODO: Can we do console.warn?
	case slog.LevelError:
		app.Log(variables...) // TODO: Can we do console.error?
	default:
		app.Log(variables...)
	}
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
