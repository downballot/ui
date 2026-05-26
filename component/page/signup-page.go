package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type SignupPage struct {
	app.Compo

	name     string
	username string
	error    string
}

func (c *SignupPage) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken != "" {
		ctx.Navigate("/")
	}
}

func (c *SignupPage) Render() app.UI {
	return myui.Page().Body(
		app.Div().
			Body(
				app.H2().Text("Downballot Signup"),
			),
		myui.Input[string]().
			Label("Name").
			Type("text").
			Bind(&c.name),
		myui.Input[string]().
			Label("E-mail address").
			Type("text").
			Bind(&c.username),
		app.If(c.error != "", func() app.UI {
			return myui.StatusBar().
				Text(c.error).
				Bad()
		}),
		app.Div().
			Class("login-page-actions").
			Body(
				app.Span().Style("flex", "1"),
				myui.Button().
					Label("Sign up").
					On("click", func(ctx app.Context, e app.Event) {
						slog.InfoContext(ctx.Context, "Button clicked")
						ctx.Async(func() {
							client := downballotapi.New("/")

							input := downballotapi.RegisterUserRequest{
								Name:     c.name,
								Username: c.username,
								Password: "BOGUS",
							}
							var output downballotapi.RegisterUserResponse
							err := client.Do(ctx.Context, http.MethodPost, "/api/v1/user", input, &output)
							if err != nil {
								app.Log(err)
								c.error = err.Error()
								ctx.Update()
								return
							}
							c.error = ""

							app.Logf("request response: %+v", output)
							ctx.Navigate("/login")
						})
					}),
			),
		app.Hr(),
		app.Div().
			Text("Already have an account?"),
		app.A().
			Href("/login").
			Text("Log in"),
	)
}
