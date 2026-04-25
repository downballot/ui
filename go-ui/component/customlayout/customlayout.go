package customlayout

import (
	"log/slog"

	"github.com/downballot/ui/go-ui/component/layout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type DownballotLayout struct {
	app.Compo

	Content app.UI
}

func (c *DownballotLayout) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		ctx.Navigate("/login")
	}
}

func (c *DownballotLayout) Render() app.UI {
	return &layout.MainLayout{
		Content: c.Content,
		Header: app.Div().Body(
			app.H1().Text("Downballot"),
		),
		Drawer: app.Div().Body(
			app.Ul().Body(
				app.Li().Body(
					app.A().Href("/login").Text("Login"),
				),
				app.Li().Body(
					app.A().Href("/organization").Text("Organization"),
				),
				app.Li().Body(
					app.A().Href("/profile").Text("Profile"),
				),
			),
		),
	}
}
