package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type ProfilePage struct {
	app.Compo
	myui.EmbeddedPage

	Loaded bool

	AuthenticatedUser *downballotapi.AuthenticationStatusUser
}

func (c *ProfilePage) OnNav(ctx app.Context) {
	ctx.Async(func() {
		var output downballotapi.AuthenticationStatusResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/authentication/status", nil, &output)
		if err != nil {
			slog.InfoContext(ctx.Context, "Could not get authenticated user", "err", err)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true

			c.AuthenticatedUser = output.User
		})
	})
}

func (c *ProfilePage) Render() app.UI {
	if !c.Loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.AuthenticatedUser == nil {
		return c.EmbeddedPage.Wrap(
			myui.StatusBar().
				Text("Not logged in.").
				Bad(),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().Body(
			app.Span().Text("Username: "),
			app.Span().Text(c.AuthenticatedUser.Email),
		),
		app.Div().Body(
			app.Span().Text("Name: "),
			app.Span().Text(c.AuthenticatedUser.Name),
		),
		myui.Form().
			Spacer(false).
			Action(myui.FormAction{
				Name: "Log out",
				Function: func(ctx app.Context) {
					ctx.DelState("api-token")

					err := api.Do(ctx, http.MethodGet, "/api/v1/authentication/status", nil, nil)
					if err != nil {
						slog.InfoContext(ctx.Context, "Could not get (un)authenticated user", "err", err)
					}
				},
			}),
	)
}
