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

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/component/customlayout"
	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/component/page"
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

	router.Register(ctx,
		router.Route{
			Path: "/",
			Component: func() app.Composer {
				return &layout.CenterLayout{}
			},
			Meta: map[string]string{
				"require-login": "false",
			},
			Children: []router.Route{
				{
					Path: "/login",
					Component: func() app.Composer {
						return &page.LoginPage{}
					},
				},
			},
		},
		router.Route{
			Path: "/",
			Component: func() app.Composer {
				return &customlayout.DownballotLayout{}
			},
			Meta: map[string]string{
				"require-login": "true",
			},
			Children: []router.Route{
				{
					Path: "/",
					Component: func() app.Composer {
						return &page.HomePage{}
					},
					Meta: map[string]string{
						"title": "Home",
					},
				},
				{
					Path: "/organization",
					Component: func() app.Composer {
						return &page.OrganizationPage{}
					},
					Meta: map[string]string{
						"title": "Organizations",
					},
				},
				{
					Path: "/organization/:organization_id",
					PathVariables: func(ctx app.Context, variables map[string]string) {
						//ctx.Dispatch(func(ctx app.Context) {
						var output downballotapi.GetOrganizationResponse
						err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"], nil, &output)
						if err != nil {
							slog.ErrorContext(ctx.Context, "Could not get organization", "err", err)
							return
						}

						variables["organization_name"] = output.Organization.Name
						//})
					},
					Meta: map[string]string{
						"autocrumbs": "true",
					},
					Component: func() app.Composer {
						return &customlayout.OrganizationLayout{}
					},
					Children: []router.Route{
						{
							Path: "/",
							Component: func() app.Composer {
								return &page.OrganizationIDPage{}
							},
							Meta: map[string]string{
								"title": ":organization_name",
							},
						},
						{
							Path: "/group",
							Component: func() app.Composer {
								return &page.OrganizationIDGroupPage{}
							},
							Meta: map[string]string{
								"title": "Groups",
							},
						},
						{
							Path: "/group/new",
							Component: func() app.Composer {
								return &page.OrganizationIDGroupNewPage{}
							},
							Meta: map[string]string{
								"title": "New Group",
							},
						},
						{
							Path:      "/group/:group_id",
							Component: nil,
							PathVariables: func(ctx app.Context, variables map[string]string) {
								var output downballotapi.GetGroupResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/group/"+variables["group_id"], nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
									return
								}

								variables["group_name"] = output.Group.Name
							},
							Meta: map[string]string{
								"title": ":group_name",
							},
							Children: []router.Route{
								{
									Path: "/",
									Component: func() app.Composer {
										return &page.OrganizationIDGroupIDPage{}
									},
									Meta: map[string]string{
										"title": ":group_name",
									},
								},
								{
									Path: "/person",
									Component: func() app.Composer {
										return &page.OrganizationIDGroupIDPersonPage{}
									},
									Meta: map[string]string{
										"title": "Persons",
									},
								},
							},
						},
						{
							Path: "/person/:voter_id",
							Component: func() app.Composer {
								return &page.OrganizationIDPersonIDPage{}
							},
							Meta: map[string]string{
								"title": ":voter_id",
							},
						},
						{
							Path: "/person-field",
							Component: func() app.Composer {
								return &page.OrganizationIDPersonFieldPage{}
							},
							Meta: map[string]string{
								"title": "Person Fields",
							},
						},
						{
							Path: "/person-field/new",
							Component: func() app.Composer {
								return &page.OrganizationIDPersonFieldNewPage{}
							},
							Meta: map[string]string{
								"title": "New Person Field",
							},
						},
						{
							Path: "/person-field/:person_field_id",
							Component: func() app.Composer {
								return &page.OrganizationIDPersonFieldIDPage{}
							},
							Meta: map[string]string{
								"title": ":person_field_name",
							},
							PathVariables: func(ctx app.Context, variables map[string]string) {
								var output downballotapi.GetPersonFieldResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/person-field/"+variables["person_field_id"], nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get person field", "err", err)
									return
								}
								variables["person_field_name"] = output.PersonField.Name
							},
						},
						{
							Path: "/user",
							Component: func() app.Composer {
								return &page.OrganizationIDUserPage{}
							},
							Meta: map[string]string{
								"title": "Users",
							},
						},
						{
							Path: "/user/:user_id",
							Component: func() app.Composer {
								return &page.OrganizationIDUserIDPage{}
							},
							Meta: map[string]string{
								"title": ":user_id",
							},
							PathVariables: func(ctx app.Context, variables map[string]string) {
								var output downballotapi.GetUserResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/user/"+variables["user_id"], nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
									return
								}
								variables["user_name"] = output.User.Name
							},
						},
					},
				},
				{
					Path: "/profile",
					Component: func() app.Composer {
						return &page.ProfilePage{}
					},
					Meta: map[string]string{
						"title": "Profile",
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
	slog.InfoContext(ctx, "main", "GOOGLE_MAPS_API_KEY", os.Getenv("GOOGLE_MAPS_API_KEY"))
	mux.Handle("/", &app.Handler{
		Name:        "Downballot",
		Description: "The official Downballot UI",
		Title:       "Downballot",
		Styles: []string{
			"/web/main.css",
		},
		RawHeaders: []string{
			`<script src="https://kit.fontawesome.com/a71e001119.js" crossorigin="anonymous"></script>`,
		},
		Env: map[string]string{
			"GOOGLE_MAPS_API_KEY": os.Getenv("GOOGLE_MAPS_API_KEY"),
		},
	})

	// Send all API requests to the API server.
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

	// Disable the service worker by replacing it with an empty one.
	mux.HandleFunc("/app-worker.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(""))
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	if err := http.ListenAndServe(":"+port, wrapper); err != nil {
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
