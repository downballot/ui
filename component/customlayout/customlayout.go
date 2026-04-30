package customlayout

import (
	"context"
	"log/slog"

	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/routelayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type DownballotLayout struct {
	app.Compo
	routelayout.RouterViewComponent
}

var _ routelayout.RouterViewInterface = (*DownballotLayout)(nil)

func (c *DownballotLayout) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		ctx.Navigate("/login")
	}
}

func (c *DownballotLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "DownballotLayout: Render")

	mainLayout := &layout.MainLayout{
		Header: app.Div().Body(
			&material.AppBar{
				Headline: "Downballot",
			},
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
	mainLayout.SetRouterView(c.RouterView())
	return mainLayout.Render()
}
