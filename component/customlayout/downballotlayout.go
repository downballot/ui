package customlayout

import (
	"context"
	"log/slog"

	"github.com/downballot/ui/component"
	"github.com/downballot/ui/component/layout"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
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

	/*
		if routerView := c.RouterViewComponent.RouterView(); routerView != nil {
			if mounter, ok := routerView.(app.Mounter); ok {
				mounter.OnMount(ctx)
			}
		}
	*/
}

func (c *DownballotLayout) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "DownballotLayout: OnNav", "url", ctx.Page().URL())

	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		slog.InfoContext(ctx.Context, "DownballotLayout: OnNav: User is not logged in, navigating to login page")
		ctx.Navigate("/login")
		return
	}
}

func (c *DownballotLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "DownballotLayout: Render")

	mainLayout := &layout.MainLayout{
		Header: blazar.AppBar().
			HeadlineText("Downballot").
			Subtitle(app.Div().
				Class("downballotlayout-menu").
				Body(
					blazar.Item().
						Icon(component.IconOrganization).
						Label("Organizations").
						To("/organization"),
					app.Span().Style("flex", "1"),
					blazar.Item().
						Icon(component.IconUser).
						Label("Profile").
						To("/profile"),
				),
			),
	}
	mainLayout.SetRouterView(c.RouterView())

	slog.InfoContext(context.TODO(), "DownballotLayout: Render", "mainLayout", mainLayout)

	return mainLayout
}
