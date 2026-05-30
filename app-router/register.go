package router

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/downballot/ui/app-router/route"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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

type registeredRoute struct {
	Regexp    *regexp.Regexp
	Component func() app.Composer
	Meta      map[string]string
}

var registeredRoutes []registeredRoute

// Register the various routes.
func Register(ctx context.Context, routes ...Route) error {
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
		compiledRegexp, err := regexp.Compile(route.Regexp())
		if err != nil {
			return fmt.Errorf("could not compile route %q: %w", r.Path, err)
		}
		newRegisteredRoute := registeredRoute{
			Regexp: compiledRegexp,
			Component: func() app.Composer {
				slog.InfoContext(ctx, "Register: func(): creating component for route.", "route", route)
				routeComponent := composeRoute(ctx, r.ComponentFunctions...)
				slog.InfoContext(ctx, "Register: func()", "routeComponent", routeComponent, "type", fmt.Sprintf("%T", routeComponent))

				wrapper := LayoutWrapper{
					LayoutComponent:        routeComponent,
					Meta:                   r.Meta,
					Route:                  *route,
					PathVariablesFunctions: r.PathVariablesFunctions,
				}

				slog.InfoContext(ctx, "Register: func()", "wrapper", fmt.Sprintf("%p", &wrapper), "LayoutComponent", fmt.Sprintf("%T", wrapper.LayoutComponent), "LayoutComponentPointer", fmt.Sprintf("%p", wrapper.LayoutComponent))
				return &wrapper
			},
			Meta: r.Meta,
		}
		registeredRoutes = append(registeredRoutes, newRegisteredRoute)
		app.RouteWithRegexp(route.Regexp(), newRegisteredRoute.Component)
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
