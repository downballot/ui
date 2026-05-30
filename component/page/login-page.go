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
		slog.InfoContext(ctx.Context, "LoginPage: OnNav: User is already logged in, navigating to home page")
		ctx.Navigate("/")
	}
}

func (c *LoginPage) Render() app.UI {
	return myui.Page().Body(
		app.Div().
			Body(
				app.H2().Text("Downballot Login"),
			),
		myui.Input[string]().
			Label("E-mail address").
			Name("email").
			Type("text").
			Disabled(c.readyForPassword).
			AutoFocus(!c.readyForPassword).
			Bind(&c.username),
		app.If(c.readyForPassword, func() app.UI {
			return app.If(c.message != "", func() app.UI {
				return myui.StatusBar().
					Text(c.message)
			})
		}),
		app.If(c.readyForPassword, func() app.UI {
			return myui.Input[string]().
				Label("Password").
				Name("password").
				Type("password").
				AutoFocus(true).
				Bind(&c.password)
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
							slog.InfoContext(ctx.Context, "LoginPage: Next button clicked", "username", c.username)

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

								newFunc := app.FuncOf(
									func(this app.Value, args []app.Value) any {
										element := c.JSValue().Call("querySelector", "[autofocus]")
										if !element.IsNull() {
											element.Call("focus")
											slog.InfoContext(ctx.Context, "Defer: Focused on input", "element", element.Get("className").String(), "name", element.Get("name").String())
										}
										return nil
									},
								)
								app.Window().Call("setTimeout", newFunc, 100)
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

								slog.InfoContext(ctx.Context, "LoginPage: Log in button clicked: Navigating to home page")
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
