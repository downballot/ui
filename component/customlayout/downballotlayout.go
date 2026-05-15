package customlayout

import (
	"context"
	"log/slog"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type DownballotLayout struct {
	app.Compo
	router.RouterViewComponent
}

var _ router.RouterViewInterface = (*DownballotLayout)(nil)

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
			Leading: app.Div().
				Class("mainlayout-header-leading").
				OnClick(func(ctx app.Context, e app.Event) {
					slog.InfoContext(ctx.Context, "DownballotLayout: Header: Leading: OnClick")
				}).
				Body(
					myui.Icon().Icon("bars"),
				),
			Headline: "Downballot",
		},
		Drawer: app.Div().
			Body(
				myui.Item().
					Icon("arrow-right-to-bracket").
					Name("Login").
					To("/login"),
				myui.Item().
					Icon("building").
					Name("Organizations").
					To("/organization"),
				myui.Item().
					Icon("user").
					Name("Profile").
					To("/profile"),
			),
	}
	mainLayout.SetRouterView(c.RouterView())
	return mainLayout.Render()
}
