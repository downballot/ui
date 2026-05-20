package router

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// StateRoute is the state key for the active route.
const StateRoute = "route"

// ActiveRoute is the active route that is currently being displayed.
type ActiveRoute struct {
	Path      string
	Meta      map[string]string
	Variables map[string]string
}

// GetActiveRoute gets the active route from the context.
func GetActiveRoute(ctx app.Context) ActiveRoute {
	var activeRoute ActiveRoute
	ctx.GetState(StateRoute, &activeRoute)
	return activeRoute
}

// GetRoute gets a route from the registered routes.
func GetRoute(ctx app.Context, path string) *ActiveRoute {
	for _, route := range registeredRoutes {
		if route.Regexp.MatchString(path) {
			return &ActiveRoute{
				Path: path,
				Meta: route.Meta,
				// TODO: Variables
			}
		}
	}
	return nil
}
