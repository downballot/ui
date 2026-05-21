package customlayout

import (
	"context"
	"log/slog"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type DownballotLayout struct {
	app.Compo
	router.RouterViewComponent
}

var _ router.RouterViewInterface = (*DownballotLayout)(nil)
var _ app.Mounter = (*DownballotLayout)(nil)

func (c *DownballotLayout) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "DownballotLayout: OnMount")
}

func (c *DownballotLayout) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "DownballotLayout: OnNav", "url", ctx.Page().URL())

	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		ctx.Navigate("/login")
		return
	}

	if component := c.RouterViewComponent.RouterView(); component != nil {
		if v, ok := component.(app.Navigator); ok {
			v.OnNav(ctx)
		}
	}
}

func (c *DownballotLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "DownballotLayout: Render")

	mainLayout := &layout.MainLayout{
		Header: &material.AppBar{
			Headline: "Downballot",
			SubtitleUI: app.Div().
				Class("downballotlayout-menu").
				Body(
					myui.Item().
						Icon("building").
						Name("Organizations").
						To("/organization"),
					app.Span().Style("flex", "1"),
					myui.Item().
						Icon("user").
						Name("Profile").
						To("/profile"),
				),
		},
	}
	mainLayout.SetRouterView(c.RouterView())

	slog.InfoContext(context.TODO(), "DownballotLayout: Render", "mainLayout", mainLayout)

	return mainLayout.Render()
}
