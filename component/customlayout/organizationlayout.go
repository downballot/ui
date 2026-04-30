package customlayout

import (
	"context"
	"log/slog"

	"github.com/downballot/ui/material"
	"github.com/downballot/ui/routelayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationLayout struct {
	app.Compo
	routelayout.RouterViewComponent

	OrganizationID string `route:"organization_id"`
}

var _ routelayout.RouterViewInterface = (*OrganizationLayout)(nil)

func (c *OrganizationLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationLayout: Render")

	return app.Div().
		Body(
			&material.AppBar{
				Headline: "Organization #" + c.OrganizationID,
			},
			app.Ul().Body(
				app.Li().Body(
					app.A().Href("/organization/"+c.OrganizationID).Text("Organization"),
				),
				app.Li().Body(
					app.A().Href("/organization/"+c.OrganizationID+"/group").Text("Groups"),
				),
			),
			c.RouterView().Render(),
		)
}
