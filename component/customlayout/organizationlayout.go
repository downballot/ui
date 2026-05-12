package customlayout

import (
	"context"
	"log/slog"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/material"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationLayout struct {
	app.Compo
	router.RouterViewComponent

	OrganizationID   string `route:"organization_id"`
	OrganizationName string `route:"organization_name"`
}

var _ router.RouterViewInterface = (*OrganizationLayout)(nil)
var _ app.Navigator = (*OrganizationLayout)(nil)

func (c *OrganizationLayout) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "url", ctx.Page().URL())

	activeRoute := router.GetActiveRoute(ctx)
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "activeRoute", activeRoute)
}

func (c *OrganizationLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationLayout: Render")

	return app.Div().
		Body(
			&material.AppBar{
				Headline: c.OrganizationName,
			},
			app.Ul().Body(
				app.Li().Body(
					app.A().Href("/organization/"+c.OrganizationID).Text("Organization"),
				),
				app.Li().Body(
					app.A().Href("/organization/"+c.OrganizationID+"/group").Text("Groups"),
				),
				app.Li().Body(
					app.A().Href("/organization/"+c.OrganizationID+"/person-field").Text("Person Fields"),
				),
			),
			c.RouterView().Render(),
		)
}
