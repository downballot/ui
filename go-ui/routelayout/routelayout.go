package routelayout

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/downballot/ui/go-ui/route"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Layout interface {
	WithComponent(app.Composer) app.Composer
}

type RouteLayout struct {
	Path      string
	Component func() Layout
	Meta      map[string]string
	Children  []RoutePage
}

type RoutePage struct {
	Path      string
	Component func() app.Composer
	Meta      map[string]string
}

func Apply(ctx context.Context, routeLayouts ...RouteLayout) error {
	for _, routeLayout := range routeLayouts {
		for _, routePage := range routeLayout.Children {
			pageMeta := map[string]string{}
			for key, value := range routeLayout.Meta {
				pageMeta[key] = value
			}
			for key, value := range routePage.Meta {
				pageMeta[key] = value
			}

			path := strings.TrimRight(routeLayout.Path, "/") + "/" + strings.TrimLeft(routePage.Path, "/")
			route, err := route.Parse(path)
			if err != nil {
				return fmt.Errorf("could not parse route %q: %w", path, err)
			}
			slog.InfoContext(ctx, "Registering route.", "path", path)
			app.RouteWithRegexp(route.Regexp(), func() app.Composer {
				routeComponent := routeLayout.Component()
				pageComponent := routePage.Component()
				component := routeComponent.WithComponent(pageComponent)

				wrapper := LayoutWrapper{
					Component: component,
					Meta:      pageMeta,
					Route:     *route,
				}
				return &wrapper
			})
		}
	}
	return nil
}

type LayoutWrapper struct {
	app.Compo

	Component app.Composer
	Meta      map[string]string
	Route     route.Route
}

type HasOnNav interface {
	OnNav(ctx app.Context)
}

func (c *LayoutWrapper) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav")
	ctx.SetState("meta", c.Meta)

	matched, variables := c.Route.Match(ctx.Page().URL().Path)
	if !matched {
		slog.WarnContext(ctx.Context, "Could not match route somehow: %s", ctx.Page().URL().Path)
	} else {
		ctx.SetState("route", variables)
	}
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "matched", matched)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "variables", variables)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "Component", fmt.Sprintf("%T", c.Component))

	if v, ok := c.Component.(HasOnNav); ok {
		v.OnNav(ctx)
	}
}

func (c *LayoutWrapper) Render() app.UI {
	return c.Component.Render()
}
