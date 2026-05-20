package layout

import (
	"context"
	"log/slog"

	router "github.com/downballot/ui/app-router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type MainLayout struct {
	app.Compo
	router.RouterViewComponent

	Header        app.UI
	Drawer        app.UI
	DrawerVisible bool
}

var _ router.RouterViewInterface = (*MainLayout)(nil)

func (c *MainLayout) ToggleDrawer() {
	c.DrawerVisible = !c.DrawerVisible
	slog.InfoContext(context.TODO(), "MainLayout: ToggleDrawer", "DrawerVisible", c.DrawerVisible)
}

func (c *MainLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "MainLayout: Render")
	slog.InfoContext(context.TODO(), "MainLayout:", "c", c, "DrawerVisible", c.DrawerVisible)

	drawerDisplay := "none"
	if c.DrawerVisible {
		drawerDisplay = "block"
	}

	return app.Div().
		Class("main-layout").
		Style("height", "100vh").
		Style("width", "100%").
		Body(
			app.Div().
				Class("main-layout-header").
				Body(c.Header),
			app.Div().
				Class("main-layout-body").
				Body(
					app.Div().
						Class("main-layout-drawer").
						Style("display", drawerDisplay).
						Body(c.Drawer),
					app.Div().
						Class("main-layout-content").
						Body(
							app.If(c.RouterViewComponent.RouterView() != nil, func() app.UI {
								return c.RouterViewComponent.RouterView().Render()
							}),
						),
				),
		)
}
