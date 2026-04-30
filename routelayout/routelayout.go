package routelayout

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/downballot/ui/route"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Route struct {
	Path          string
	PathVariables func(ctx app.Context, variables map[string]string)
	Component     func() app.Composer
	Meta          map[string]string
	Children      []Route
}

type internalRoute struct {
	Path               string
	PathVariables      []func(ctx app.Context, variables map[string]string)
	ComponentFunctions []func() app.Composer
	Meta               map[string]string
}

// Apply the various routes.
func Apply(ctx context.Context, routes ...Route) error {
	flatRoutes, err := flattenRoutes(ctx, internalRoute{}, routes...)
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
			slog.InfoContext(ctx, "RouteLayout: func(): creating component for route.", "route", route)
			routeComponent := composeRoute(ctx, r.ComponentFunctions...)
			slog.InfoContext(ctx, "RouteLayout: func()", "routeComponent", routeComponent, "type", fmt.Sprintf("%T", routeComponent))

			wrapper := LayoutWrapper{
				LayoutComponent: routeComponent,
				Meta:            r.Meta,
				Route:           *route,
				PathVariables:   r.PathVariables,
			}
			return &wrapper
		})
	}
	return nil
}

func composeRoute(ctx context.Context, fs ...func() app.Composer) app.Composer {
	goodFunctions := []func() app.Composer{}
	for _, f := range fs {
		if f == nil {
			continue
		}
		goodFunctions = append(goodFunctions, f)
	}

	if len(goodFunctions) == 0 {
		return nil
	}

	slices.Reverse(goodFunctions)

	var output app.Composer
	for _, f := range goodFunctions {
		component := f()
		slog.DebugContext(ctx, "composeRoute: Created component", "type", fmt.Sprintf("%T", component))
		if component != nil {
			if hasRouterView, ok := component.(RouterViewInterface); ok {
				slog.InfoContext(ctx, "composeRoute: component is a RouterViewInterface.", "component", fmt.Sprintf("%T", component))
				hasRouterView.SetRouterView(output)
			}
		}
		output = component
	}
	return output
}

// flattenRoutes flattens a list of potentially nested routes.
func flattenRoutes(ctx context.Context, parentRoute internalRoute, routes ...Route) ([]internalRoute, error) {
	var output []internalRoute

	for _, route := range routes {
		newPath := strings.TrimLeft(strings.TrimRight(route.Path, "/"), "/")
		newRoute := internalRoute{
			Path:               strings.TrimRight(parentRoute.Path, "/"),
			ComponentFunctions: []func() app.Composer{},
			Meta:               map[string]string{},
		}
		if newPath != "" {
			newRoute.Path += "/" + newPath
		}
		newRoute.ComponentFunctions = append(newRoute.ComponentFunctions, parentRoute.ComponentFunctions...)
		newRoute.PathVariables = append(newRoute.PathVariables, parentRoute.PathVariables...)
		for key, value := range parentRoute.Meta {
			newRoute.Meta[key] = value
		}
		newRoute.ComponentFunctions = append(newRoute.ComponentFunctions, route.Component)
		newRoute.PathVariables = append(newRoute.PathVariables, route.PathVariables)
		for key, value := range route.Meta {
			newRoute.Meta[key] = value
		}
		//slog.DebugContext(ctx, "flattenRoutes", "parentRoute", parentRoute)
		//slog.DebugContext(ctx, "flattenRoutes", "parentRoute.Path", parentRoute.Path, "route.Path", route.Path)
		//slog.DebugContext(ctx, "flattenRoutes", "newRoute", newRoute)

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
	PathVariables   []func(ctx app.Context, variables map[string]string)
}

func (c *LayoutWrapper) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnMount", "url", ctx.Page().URL())

	if v, ok := c.LayoutComponent.(app.Mounter); ok {
		v.OnMount(ctx)
	}
}

func (c *LayoutWrapper) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "url", ctx.Page().URL())
	ctx.SetState("meta", c.Meta)

	matched, variables := c.Route.Match(ctx.Page().URL().Path)
	if !matched {
		slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not match route somehow.", "route", c.Route, "path", ctx.Page().URL().Path)
	} else {
		c.RouteVariables = variables

		//ctx.Dispatch(func(ctx app.Context) {
		for _, f := range c.PathVariables {
			if f == nil {
				continue
			}
			f(ctx, c.RouteVariables)
		}
		ctx.Update()

		if v, ok := c.LayoutComponent.(RouterViewInterface); ok {
			slog.DebugContext(ctx, "LayoutWrapper: OnNav: Applying variables.", "LayoutComponent", fmt.Sprintf("%T", c.LayoutComponent))
			err := v.ApplyVariables(variables)
			if err != nil {
				slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not apply variables.", "err", err)
			}
		}
		if v, ok := c.LayoutComponent.(app.Updater); ok {
			v.OnUpdate(ctx)
		}
		//})
	}
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "matched", matched)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "variables", variables)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "Component", fmt.Sprintf("%T", c.LayoutComponent))

	if v, ok := c.LayoutComponent.(app.Navigator); ok {
		v.OnNav(ctx)
	}
}

func (c *LayoutWrapper) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnUpdate", "url", ctx.Page().URL())

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
