package routelayout

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/downballot/ui/route"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Route struct {
	Path      string
	Component func() app.Composer
	Meta      map[string]string
	Children  []Route
}

// Apply the various routes.
func Apply(ctx context.Context, routes ...Route) error {
	flatRoutes, err := flattenRoutes(ctx, Route{}, routes...)
	if err != nil {
		return fmt.Errorf("could not flatten routes: %w", err)
	}

	for _, r := range flatRoutes {
		route, err := route.Parse(r.Path)
		if err != nil {
			return fmt.Errorf("could not parse route %q: %w", r.Path, err)
		}
		slog.InfoContext(ctx, "Registering route.", "path", r.Path)
		app.RouteWithRegexp(route.Regexp(), func() app.Composer {
			slog.InfoContext(ctx, "RouteLayout: creating component for route.", "route", route)
			routeComponent := r.Component()
			slog.InfoContext(ctx, "RouteLayout:", "routeComponent", routeComponent, "type", fmt.Sprintf("%T", routeComponent))

			wrapper := LayoutWrapper{
				LayoutComponent: routeComponent,
				Meta:            r.Meta,
				Route:           *route,
			}
			return &wrapper
		})
	}
	return nil
}

func composeRoute(ctx context.Context, fs ...func() app.Composer) func() app.Composer {
	goodFunctions := []func() app.Composer{}
	for _, f := range fs {
		if f == nil {
			continue
		}
		goodFunctions = append(goodFunctions, f)
	}

	if len(goodFunctions) == 0 {
		return func() app.Composer {
			return nil
		}
	}

	firstFunction := goodFunctions[0]
	remainingFunctions := goodFunctions[1:]
	if len(remainingFunctions) == 0 {
		return firstFunction
	}
	secondFunction := composeRoute(ctx, remainingFunctions...)

	return func() app.Composer {
		firstComponent := firstFunction()
		secondComponent := secondFunction()
		slog.DebugContext(ctx, "composeRoute: Generated new component.", "firstComponent", fmt.Sprintf("%T", firstComponent))
		slog.DebugContext(ctx, "composeRoute: Generated new component.", "secondComponent", fmt.Sprintf("%T", secondComponent))

		if firstComponent == nil {
			return secondComponent
		}
		if secondComponent == nil {
			return firstComponent
		}

		if hasRouterView, ok := firstComponent.(RouterViewInterface); ok {
			slog.InfoContext(ctx, "RouteLayout: firstComponent is a RouterViewInterface.")
			hasRouterView.SetRouterView(secondComponent)
		}
		return firstComponent
	}
}

// flattenRoutes flattens a list of potentially nested routes.
func flattenRoutes(ctx context.Context, parentRoute Route, routes ...Route) ([]Route, error) {
	var output []Route

	for _, route := range routes {
		newRoute := Route{
			Path:      strings.TrimRight(parentRoute.Path, "/") + "/" + strings.TrimLeft(route.Path, "/"),
			Component: composeRoute(ctx, parentRoute.Component, route.Component),
			Meta:      map[string]string{},
		}
		for key, value := range parentRoute.Meta {
			newRoute.Meta[key] = value
		}
		for key, value := range route.Meta {
			newRoute.Meta[key] = value
		}

		if len(route.Children) == 0 {
			output = append(output, newRoute)
			continue
		}

		flatRoutes, err := flattenRoutes(ctx, newRoute, route.Children...)
		if err != nil {
			return nil, err
		}
		output = append(output, flatRoutes...)
	}

	return output, nil
}

type LayoutWrapper struct {
	app.Compo

	LayoutComponent app.Composer
	Meta            map[string]string
	Route           route.Route
	RouteVariables  map[string]string
}

func (c *LayoutWrapper) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnMount")

	if v, ok := c.LayoutComponent.(app.Mounter); ok {
		v.OnMount(ctx)
	}
}

func (c *LayoutWrapper) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav")
	ctx.SetState("meta", c.Meta)

	matched, variables := c.Route.Match(ctx.Page().URL().Path)
	if !matched {
		slog.WarnContext(ctx.Context, "Could not match route somehow.", "route", c.Route, "path", ctx.Page().URL().Path)
	} else {
		c.RouteVariables = variables

		if v, ok := c.LayoutComponent.(RouterViewInterface); ok {
			slog.DebugContext(ctx, "LayoutWrapper: OnNav: Applying variables.", "LayoutComponent", fmt.Sprintf("%T", c.LayoutComponent))
			err := v.ApplyVariables(variables)
			if err != nil {
				slog.WarnContext(ctx.Context, "Could not apply variables.", "err", err)
			}
		}
		if v, ok := c.LayoutComponent.(app.Updater); ok {
			v.OnUpdate(ctx)
		}
	}
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "matched", matched)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "variables", variables)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "Component", fmt.Sprintf("%T", c.LayoutComponent))

	if v, ok := c.LayoutComponent.(app.Navigator); ok {
		v.OnNav(ctx)
	}
}

func (c *LayoutWrapper) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnUpdate")

	if v, ok := c.LayoutComponent.(app.Updater); ok {
		v.OnUpdate(ctx)
	}
}

// TODO: It looks like we want to leverage OnMount (first time) and OnUpdate (subsequent times) to tell our components that something has changed.
// TODO: Or, consider adding a public Route property to the pages that need routes so that this info can automatically do what it needs to.

func (c *LayoutWrapper) Render() app.UI {
	slog.InfoContext(context.TODO(), "LayoutWrapper: Render")

	return c.LayoutComponent.Render()
}
