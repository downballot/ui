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
		slog.InfoContext(ctx.Context, "SignupPage: OnNav: User is already logged in, navigating to home page")
		ctx.Navigate("/")
	}
}

func (c *SignupPage) Render() app.UI {
	return myui.Page().Body(
		app.Div().
			Body(
				app.H2().Text("Downballot Signup"),
			),
		myui.Form().
			Body(
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
			).
			SubmitLabel("Sign up").
			SubmitFunction(func(ctx app.Context) {
				ctx.PreventUpdate()

				ctx.Async(func() {
					client := downballotapi.New("/")

					input := downballotapi.RegisterUserRequest{
						Name:     c.name,
						Username: c.username,
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

					slog.InfoContext(ctx.Context, "SignupPage: Sign up button clicked: Navigating to login page")
					ctx.Navigate("/login")
				})
			}),
		app.Hr(),
		app.Div().
			Text("Already have an account?"),
		app.A().
			Href("/login").
			Text("Log in"),
	)
}
