package router

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// StateRoute is the state key for the active route.
const StateRoute = "route"

// ActiveRoute is the active route that is currently being displayed.
type ActiveRoute struct {
	Meta      map[string]string
	Variables map[string]string
}

// GetActiveRoute gets the active route from the context.
func GetActiveRoute(ctx app.Context) ActiveRoute {
	var activeRoute ActiveRoute
	ctx.GetState(StateRoute, &activeRoute)
	return activeRoute
}
