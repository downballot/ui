package layout

import (
	"context"
	"log/slog"

	router "github.com/downballot/ui/app-router"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type MainLayout struct {
	app.Compo
	router.RouterViewComponent

	Header app.UI
	Drawer app.UI
}

var _ router.RouterViewInterface = (*MainLayout)(nil)

func (c *MainLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "MainLayout: Render")
	slog.InfoContext(context.TODO(), "MainLayout:", "c", c)

	return app.Div().
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
