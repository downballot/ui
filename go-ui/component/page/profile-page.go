package page

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/downballot/ui/go-ui/component/customlayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ProfilePage struct {
	app.Compo

	authenticatedUser downballotapi.AuthenticationStatusResponse
}

func (c *ProfilePage) OnNav(ctx app.Context) {
	err := api.Do(ctx, http.MethodGet, "/api/v1/authentication/status", nil, &c.authenticatedUser)
	if err != nil {
		slog.InfoContext(ctx.Context, "Could not get authenticated user", "err", err)
	}
}

func (c *ProfilePage) Render() app.UI {
	return &customlayout.DownballotLayout{
		Content: app.If(c.authenticatedUser.User == nil, func() app.UI {
			return app.Div().Body(
				app.Span().Text("Not logged in."),
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Span().Text("Username: "),
				app.Span().Text(c.authenticatedUser.User.Email),
				app.Br(),
				app.Span().Text("Name: "),
				app.Span().Text(c.authenticatedUser.User.Name),
				app.Br(),
			)
		}),
	}
}
