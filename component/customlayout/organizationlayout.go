package customlayout

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/material"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationLayout struct {
	app.Compo
	router.RouterViewComponent

	organizationID   string
	organizationName string

	crumbs []Crumb
}

type Crumb struct {
	Name string
	To   string
}

var _ router.RouterViewInterface = (*OrganizationLayout)(nil)
var _ app.Navigator = (*OrganizationLayout)(nil)
var _ app.Mounter = (*OrganizationLayout)(nil)

func (c *OrganizationLayout) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnMount")
}

func (c *OrganizationLayout) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnUpdate")
}

func (c *OrganizationLayout) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "url", ctx.Page().URL())

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("organization_name", &c.organizationName)

	activeRoute := router.GetActiveRoute(ctx)
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "activeRoute", activeRoute)

	var crumbs []Crumb

	path := activeRoute.Path
	pathCount := 0
	for path != "/" {
		pathCount++
		route := router.GetRoute(ctx, path)
		if route != nil {
			slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "pathCount", pathCount, "route", route)
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
		slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "pathCount", pathCount, "path", path)

	}
	slices.Reverse(crumbs)
	if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].To = ""
	}

	c.crumbs = crumbs
	slog.InfoContext(ctx.Context, "OrganizationLayout: OnNav", "crumbs", c.crumbs)
}

func (c *OrganizationLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationLayout: Render", "OrganizationID", c.organizationID, "OrganizationName", c.organizationName)

	bodyItems := []app.UI{}
	for _, crumb := range c.crumbs {
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
			Headline:   c.organizationName,
			HeadlineUI: headline,
		},
		Drawer: app.Div().
			Class("organizationlayout-menu").
			Body(
				myui.Item().
					Icon("house").
					Name("Home").
					To("/organization/"+c.organizationID),
				myui.Item().
					Icon("people-group").
					Name("Groups").
					To("/organization/"+c.organizationID+"/group"),
				myui.Item().
					Icon("filter").
					Name("Filters").
					To("/organization/"+c.organizationID+"/filter"),
				myui.Item().
					Icon("user-gear").
					Name("Person Fields").
					To("/organization/"+c.organizationID+"/person-field"),
				myui.Item().
					Icon("user").
					Name("Users").
					To("/organization/"+c.organizationID+"/user"),
			),
	}
	mainLayout.SetRouterView(c.RouterView())

	return mainLayout
}
