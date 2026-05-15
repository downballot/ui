package customlayout

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/material"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationLayout struct {
	app.Compo
	router.RouterViewComponent

	OrganizationID   string `route:"organization_id"`
	OrganizationName string `route:"organization_name"`

	Crumbs []Crumb
}

type Crumb struct {
	Name string
	To   string
}

var _ router.RouterViewInterface = (*OrganizationLayout)(nil)
var _ app.Navigator = (*OrganizationLayout)(nil)

func (c *OrganizationLayout) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "url", ctx.Page().URL())

	activeRoute := router.GetActiveRoute(ctx)
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "activeRoute", activeRoute)

	var crumbs []Crumb

	path := activeRoute.Path
	for path != "/" {
		route := router.GetRoute(ctx, path)
		if route != nil {
			slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "route", route)
			if route.Meta["autocrumbs"] != "true" {
				break
			}

			crumb := Crumb{
				Name: route.Meta["title"],
				To:   path,
			}
			if strings.HasPrefix(crumb.Name, ":") {
				newName := activeRoute.Variables[strings.TrimPrefix(crumb.Name, ":")]
				if newName != "" {
					crumb.Name = newName
				}
			}
			crumbs = append(crumbs, crumb)
		}

		index := strings.LastIndex(path, "/")
		if index == -1 {
			break
		}
		path = path[:index]
		if path == "" {
			path = "/"
		}
		slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "path", path)

	}
	slices.Reverse(crumbs)
	if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].To = ""
	}

	c.Crumbs = crumbs
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "crumbs", c.Crumbs)
}

func (c *OrganizationLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationLayout: Render")

	bodyItems := []app.UI{}
	for _, crumb := range c.Crumbs {
		if len(bodyItems) > 0 {
			bodyItems = append(bodyItems,
				app.Text(" / "),
			)
		}
		bodyItems = append(bodyItems,
			app.Span().Body(
				app.A().Href(crumb.To).Text(crumb.Name),
			),
		)
	}
	headline := app.Div().Body(
		bodyItems...,
	)

	return app.Div().
		Body(
			&material.AppBar{
				Headline:   c.OrganizationName,
				HeadlineUI: headline,
			},
			app.Ul().
				Style("list-style-type", "none").
				Style("padding", "0").
				Style("margin", "0").
				Style("display", "flex").
				Style("gap", "20px").
				Body(
					app.Li().
						Style("margin-left", "auto").
						Style("margin-right", "auto").
						Body(
							app.A().Href("/organization/"+c.OrganizationID).Text("Organization"),
						),
					app.Li().
						Style("margin-left", "auto").
						Style("margin-right", "auto").
						Body(
							app.A().Href("/organization/"+c.OrganizationID+"/group").Text("Groups"),
						),
					app.Li().
						Style("margin-left", "auto").
						Style("margin-right", "auto").
						Body(
							app.A().Href("/organization/"+c.OrganizationID+"/person-field").Text("Person Fields"),
						),
				),
			c.RouterView().Render(),
		)
}
