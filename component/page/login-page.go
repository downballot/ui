package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/myui"
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
		app.Div().
			Body(
				app.H2().Text("Downballot Login"),
			),
		myui.Input().
			Label("Username").
			Type("text").
			Value(c.username).
			On("change", c.ValueTo(&c.username)),
		myui.Input().
			Label("Password").
			Type("password").
			Value(c.password).
			On("change", c.ValueTo(&c.password)),
		myui.Button().
			Label("Log in").
			On("click", func(ctx app.Context, e app.Event) {
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
