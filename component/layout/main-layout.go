package layout

import (
	"context"
	"log/slog"

	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type MainLayout struct {
	app.Compo
	router.RouterViewComponent

	Header app.UI
	Drawer app.UI
}

var _ router.RouterViewInterface = (*MainLayout)(nil)

func (c *MainLayout) ToggleDrawer() {
	slog.InfoContext(context.TODO(), "MainLayout: ToggleDrawer")
}

func (c *MainLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "MainLayout: Render")
	slog.InfoContext(context.TODO(), "MainLayout:", "c", c)

	return app.Div().
		Class("main-layout").
		Body(
			app.If(c.Header != nil, func() app.UI {
				return app.Div().
					Class("main-layout-header").
					Body(c.Header)
			}),
			app.Div().
				Class("main-layout-body").
				Body(
					app.If(c.Drawer != nil, func() app.UI {
						return app.Div().
							Class("main-layout-drawer").
							Body(c.Drawer)
					}),
					app.Div().
						Class("main-layout-content").
						Body(
							app.If(c.RouterViewComponent.RouterView() != nil, func() app.UI {
								return c.RouterViewComponent.RouterView()
							}),
						),
				),
		)
}
