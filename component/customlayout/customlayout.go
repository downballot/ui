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
				Leading: app.Div().
					OnClick(func(ctx app.Context, e app.Event) {
						slog.InfoContext(ctx.Context, "DownballotLayout: Header: Leading: OnClick")
					}).
					Body(
						app.Raw(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 640" width="1.5em" height="1.5em"><!--!Font Awesome Free v7.2.0 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2026 Fonticons, Inc.--><path d="M96 160C96 142.3 110.3 128 128 128L512 128C529.7 128 544 142.3 544 160C544 177.7 529.7 192 512 192L128 192C110.3 192 96 177.7 96 160zM96 320C96 302.3 110.3 288 128 288L512 288C529.7 288 544 302.3 544 320C544 337.7 529.7 352 512 352L128 352C110.3 352 96 337.7 96 320zM544 480C544 497.7 529.7 512 512 512L128 512C110.3 512 96 497.7 96 480C96 462.3 110.3 448 128 448L512 448C529.7 448 544 462.3 544 480z"/></svg>`),
					),
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
