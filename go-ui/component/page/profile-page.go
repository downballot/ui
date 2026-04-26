package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ProfilePage struct {
	app.Compo

	AuthenticatedUser *downballotapi.AuthenticationStatusUser
}

func (c *ProfilePage) OnUpdate(ctx app.Context) {
	ctx.Async(func() {
		var output downballotapi.AuthenticationStatusResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/authentication/status", nil, &output)
		if err != nil {
			slog.InfoContext(ctx.Context, "Could not get authenticated user", "err", err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.AuthenticatedUser = output.User
		})
	})
}

func (c *ProfilePage) Render() app.UI {
	return app.Div().Body(
		app.If(c.AuthenticatedUser == nil, func() app.UI {
			return app.Div().Body(
				app.Span().Text("Not logged in."),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Span().Text("Username: "),
				app.Span().Text(c.AuthenticatedUser.Email),
				app.Br(),
				app.Span().Text("Name: "),
				app.Span().Text(c.AuthenticatedUser.Name),
				app.Br(),
				app.Button().
					Text("Log out").
					OnClick(func(ctx app.Context, e app.Event) {
						ctx.DelState("api-token")

						err := api.Do(ctx, http.MethodGet, "/api/v1/authentication/status", nil, nil)
						if err != nil {
							slog.InfoContext(ctx.Context, "Could not get (un)authenticated user", "err", err)
						}
					}),
			)
		}),
	)
}
