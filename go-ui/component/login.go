package component

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type LoginPage struct {
	app.Compo

	username string
	password string
}

func (c *LoginPage) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken != "" {
		ctx.Navigate("/")
	}
}

func (c *LoginPage) Render() app.UI {
	return app.Div().Body(
		app.Span().Text("Username"),
		app.Br(),
		app.Input().
			Value(c.username).
			OnChange(c.ValueTo(&c.username)),
		app.Br(),
		app.Span().Text("Password"),
		app.Br(),
		app.Input().
			Type("password").
			Value(c.password).
			OnChange(c.ValueTo(&c.password)),
		app.Br(),
		app.Button().
			Text("Log in").
			OnClick(func(ctx app.Context, e app.Event) {
				slog.InfoContext(ctx.Context, "Button clicked")
				ctx.Async(func() {
					client := downballotapi.New("/")

					input := downballotapi.LoginRequest{
						Username: c.username,
						Password: c.password,
					}
					var output downballotapi.LoginResponse
					err := client.Do(ctx.Context, http.MethodPost, "/api/v1/authentication/login", input, &output)
					if err != nil {
						app.Log(err)
						return
					}

					app.Logf("request response: %+v", output)
					ctx.SetState("api-token", output.Token).Persist()
					ctx.Navigate("/")
				})
			}),
	)
}
