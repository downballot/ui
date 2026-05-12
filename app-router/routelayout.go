package router

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/downballot/ui/app-router/route"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// Route is a route that can be applied to the application.
//
// This is analogous to a Vue router route.
type Route struct {
	Path          string                                             // The path of the route.  This may optionally start with "/", and variables are of the form ":variable_name".
	PathVariables func(ctx app.Context, variables map[string]string) // A function that will be called with the current path variables; additional work can be done to set others.
	Component     func() app.Composer                                // A function that will be called to create the component for the route.
	Meta          map[string]string                                  // Metadata for the route.
	Children      []Route                                            // Children routes (if any).
}

// internalRoute is a route that is used internally to flatten the routes.
type internalRoute struct {
	Path                   string
	PathVariablesFunctions []func(ctx app.Context, variables map[string]string)
	ComponentFunctions     []func() app.Composer
	Meta                   map[string]string
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
				layoutComponent:        routeComponent,
				meta:                   r.Meta,
				route:                  *route,
				pathVariablesFunctions: r.PathVariablesFunctions,
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
		newPath := "/" + strings.TrimLeft(strings.TrimRight(route.Path, "/"), "/")
		newRoute := internalRoute{
			Path:               strings.TrimRight(parentRoute.Path, "/"),
			ComponentFunctions: []func() app.Composer{},
			Meta:               map[string]string{},
		}
		if newPath != "" {
			newRoute.Path += "/" + strings.TrimLeft(newPath, "/")
		}
		if newRoute.Path != "/" {
			newRoute.Path = strings.TrimRight(newRoute.Path, "/")
		}
		newRoute.ComponentFunctions = append(newRoute.ComponentFunctions, parentRoute.ComponentFunctions...)
		newRoute.PathVariablesFunctions = append(newRoute.PathVariablesFunctions, parentRoute.PathVariablesFunctions...)
		for key, value := range parentRoute.Meta {
			newRoute.Meta[key] = value
		}
		newRoute.ComponentFunctions = append(newRoute.ComponentFunctions, route.Component)
		newRoute.PathVariablesFunctions = append(newRoute.PathVariablesFunctions, route.PathVariables)
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

// LayoutWrapper is a wrapper around the desired component that handles top-level events to ensure that the component is properly updated.
type LayoutWrapper struct {
	app.Compo

	layoutComponent        app.Composer
	meta                   map[string]string
	route                  route.Route
	routeVariables         map[string]string
	pathVariablesFunctions []func(ctx app.Context, variables map[string]string)
}

func (c *LayoutWrapper) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnMount", "url", ctx.Page().URL())

	if v, ok := c.layoutComponent.(app.Mounter); ok {
		v.OnMount(ctx)
	}
}

func (c *LayoutWrapper) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "url", ctx.Page().URL())

	matched, variables := c.route.Match(ctx.Page().URL().Path)
	if !matched {
		slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not match route somehow.", "route", c.route, "path", ctx.Page().URL().Path)
	} else {
		c.routeVariables = variables

		//ctx.Dispatch(func(ctx app.Context) {
		for _, f := range c.pathVariablesFunctions {
			if f == nil {
				continue
			}
			f(ctx, c.routeVariables)
		}

		activeRoute := ActiveRoute{
			Meta:      c.meta,
			Variables: c.routeVariables,
		}
		slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav: Setting active route.", "activeRoute", activeRoute)
		ctx.SetState(StateRoute, activeRoute)

		ctx.Update()

		if v, ok := c.layoutComponent.(RouterViewInterface); ok {
			slog.DebugContext(ctx, "LayoutWrapper: OnNav: Applying variables.", "LayoutComponent", fmt.Sprintf("%T", c.layoutComponent))
			err := v.ApplyVariables(variables)
			if err != nil {
				slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not apply variables.", "err", err)
			}
		}
		if v, ok := c.layoutComponent.(app.Updater); ok {
			v.OnUpdate(ctx)
		}
		//})
	}
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "matched", matched)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "variables", variables)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "Component", fmt.Sprintf("%T", c.layoutComponent))

	if v, ok := c.layoutComponent.(app.Navigator); ok {
		v.OnNav(ctx)
	}
}

func (c *LayoutWrapper) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnUpdate", "url", ctx.Page().URL())

	if v, ok := c.layoutComponent.(app.Updater); ok {
		v.OnUpdate(ctx)
	}
}

// TODO: It looks like we want to leverage OnMount (first time) and OnUpdate (subsequent times) to tell our components that something has changed.
// TODO: Or, consider adding a public Route property to the pages that need routes so that this info can automatically do what it needs to.

func (c *LayoutWrapper) Render() app.UI {
	slog.InfoContext(context.TODO(), "LayoutWrapper: Render")

	return c.layoutComponent.Render()
}
