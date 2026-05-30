package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/downballot/ui/app-router/route"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// LayoutWrapper is a wrapper around the desired component that handles top-level events to ensure that the component is properly updated.
type LayoutWrapper struct {
	app.Compo

	// These need to be public so that the component is properly re-rendered.
	LayoutComponent        app.Composer
	Meta                   map[string]string
	Route                  route.Route
	RouteVariables         map[string]string
	PathVariablesFunctions []func(ctx app.Context, variables map[string]string)

	GlobalRouteCount uint64
}

func (c *LayoutWrapper) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnMount", "url", ctx.Page().URL())

	/*
		if v, ok := c.LayoutComponent.(app.Mounter); ok {
			v.OnMount(ctx)
		}
	*/
}

func (c *LayoutWrapper) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "url", ctx.Page().URL())
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "self", fmt.Sprintf("%p", c), "LayoutComponent", fmt.Sprintf("%T", c.LayoutComponent), "LayoutComponentPointer", fmt.Sprintf("%p", c.LayoutComponent))
	if currentRouteURL == "" {
		currentRouteURL = ctx.Page().URL().String()
	} else if currentRouteURL == ctx.Page().URL().String() {
		slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav: Preventing update because the URL has not changed.", "url", ctx.Page().URL())
		ctx.PreventUpdate()
		return
	}
	currentRouteURL = ctx.Page().URL().String()

	matched, variables := c.Route.Match(ctx.Page().URL().Path)

	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "matched", matched)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "variables", variables)
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav", "Component", fmt.Sprintf("%T", c.LayoutComponent))

	if !matched {
		slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not match route somehow.", "route", c.Route, "path", ctx.Page().URL().Path)
	} else {
		c.RouteVariables = variables

		for _, f := range c.PathVariablesFunctions {
			if f == nil {
				continue
			}
			f(ctx, c.RouteVariables)
		}

		activeRoute := ActiveRoute{
			Path:      ctx.Page().URL().Path,
			Meta:      c.Meta,
			Variables: c.RouteVariables,
		}
		slog.InfoContext(ctx.Context, "LayoutWrapper: OnNav: Setting active route.", "activeRoute", activeRoute)
		ctx.SetState(StateRoute, activeRoute)

		if v, ok := c.LayoutComponent.(RouterViewInterface); ok {
			slog.DebugContext(ctx, "LayoutWrapper: OnNav: Applying variables.", "LayoutComponent", fmt.Sprintf("%T", c.LayoutComponent))
			err := v.ApplyVariables(variables)
			if err != nil {
				slog.WarnContext(ctx.Context, "LayoutWrapper: OnNav: Could not apply variables.", "err", err)
			}
		}
	}
}

var currentRouteURL string

func (c *LayoutWrapper) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "LayoutWrapper: OnUpdate", "url", ctx.Page().URL())
	if currentRouteURL == "" {
		// TODO: ???
	} else if currentRouteURL == ctx.Page().URL().String() {
		slog.InfoContext(ctx.Context, "LayoutWrapper: OnUpdate: Preventing update because the URL has not changed.", "url", ctx.Page().URL())
		ctx.PreventUpdate()
		return
	}
}

// TODO: It looks like we want to leverage OnMount (first time) and OnUpdate (subsequent times) to tell our components that something has changed.
// TODO: Or, consider adding a public Route property to the pages that need routes so that this info can automatically do what it needs to.

func (c *LayoutWrapper) Render() app.UI {
	slog.InfoContext(context.TODO(), "LayoutWrapper: Render")

	//return c.LayoutComponent.Render()
	return c.LayoutComponent
}
