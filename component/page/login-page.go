package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type LoginPage struct {
	app.Compo

	username         string
	readyForPassword bool
	password         string
	error            string
	message          string
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
	return myui.Page().Body(
		app.Div().
			Body(
				app.H2().Text("Downballot Login"),
			),
		myui.Input().
			Label("E-mail address").
			Type("text").
			Disabled(c.readyForPassword).
			Value(c.username).
			On("change", c.ValueTo(&c.username)),
		app.If(c.readyForPassword, func() app.UI {
			return app.If(c.message != "", func() app.UI {
				return myui.StatusBar().
					Text(c.message)
			})
		}),
		app.If(c.readyForPassword, func() app.UI {
			return myui.Input().
				Label("Password").
				Type("password").
				Value(c.password).
				On("change", c.ValueTo(&c.password))
		}),
		app.If(c.error != "", func() app.UI {
			return myui.StatusBar().
				Text(c.error).
				Bad()
		}),
		app.Div().
			Class("login-page-actions").
			Body(
				app.If(c.readyForPassword, func() app.UI {
					return myui.Button().
						Label("Cancel").
						On("click", func(ctx app.Context, e app.Event) {
							c.error = ""
							c.password = ""
							c.username = ""
							c.readyForPassword = false
							c.message = ""
							ctx.Update()
						})
				}),
				app.Span().Style("flex", "1"),
				app.If(!c.readyForPassword, func() app.UI {
					return myui.Button().
						Label("Next").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.Async(func() {
								client := downballotapi.New("/")

								input := downballotapi.EmailRequest{
									Email: c.username,
								}
								var output downballotapi.EmailResponse
								err := client.Do(ctx.Context, http.MethodPost, "/api/v1/authentication/email", input, &output)
								if err != nil {
									app.Log(err)
									c.error = err.Error()
									ctx.Update()
									return
								}
								c.error = ""
								c.message = output.Message

								c.readyForPassword = true
								ctx.Update()
							})
						})
				}),
				app.If(c.readyForPassword, func() app.UI {
					return myui.Button().
						Label("Log in").
						On("click", func(ctx app.Context, e app.Event) {
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
									c.error = err.Error()
									ctx.Update()
									return
								}
								c.error = ""
								c.message = "Logged in successfully"

								app.Logf("request response: %+v", output)
								ctx.SetState("api-token", output.Token).Persist()
								ctx.Navigate("/")
							})
						})
				}),
			),
		app.Hr(),
		app.Div().
			Text("Don't have an account?"),
		app.A().
			Href("/signup").
			Text("Sign up"),
	)
}
