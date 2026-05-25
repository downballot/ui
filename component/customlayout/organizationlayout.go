package customlayout

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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

	if c.RouterViewComponent.RouterView() != nil {
		if navigator, ok := c.RouterViewComponent.RouterView().(app.Navigator); ok {
			navigator.OnNav(ctx)
		}
	}
}

func (c *OrganizationLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationLayout: Render", "OrganizationID", c.OrganizationID, "OrganizationName", c.OrganizationName)

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

	mainLayout := &layout.MainLayout{
		Header: &material.AppBar{
			Leading: app.Div().
				Class("organizationlayout-header-leading").
				Style("cursor", "pointer").
				OnClick(func(ctx app.Context, e app.Event) {
					slog.InfoContext(ctx.Context, "OrganizationLayout: Header: Leading: OnClick")
					ctx.PreventUpdate()

					mainLayoutElement := e.Get("target").Call("closest", ".main-layout")
					slog.InfoContext(ctx.Context, "OrganizationLayout: Header: Leading: OnClick", "mainLayoutElement", mainLayoutElement)
					if !mainLayoutElement.IsNull() {
						drawerElement := mainLayoutElement.Call("querySelector", ".main-layout-drawer")
						slog.InfoContext(ctx.Context, "OrganizationLayout: Header: Leading: OnClick", "drawerElement", drawerElement)
						if !drawerElement.IsNull() {
							drawerElement.Get("classList").Call("toggle", "visible")
						}
					}
				}).
				Body(
					myui.Icon().Icon("bars"),
				),
			Headline:   c.OrganizationName,
			HeadlineUI: headline,
		},
		Drawer: app.Div().
			Class("organizationlayout-menu").
			Body(
				myui.Item().
					Icon("house").
					Name("Home").
					To("/organization/"+c.OrganizationID),
				myui.Item().
					Icon("people-group").
					Name("Groups").
					To("/organization/"+c.OrganizationID+"/group"),
				myui.Item().
					Icon("filter").
					Name("Filters").
					To("/organization/"+c.OrganizationID+"/filter"),
				myui.Item().
					Icon("user-gear").
					Name("Person Fields").
					To("/organization/"+c.OrganizationID+"/person-field"),
				myui.Item().
					Icon("user").
					Name("Users").
					To("/organization/"+c.OrganizationID+"/user"),
			),
	}
	mainLayout.SetRouterView(c.RouterView())

	return mainLayout.Render()
}
