package layout

import (
	"context"
	"log/slog"

	"github.com/downballot/ui/routelayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type CenterLayout struct {
	app.Compo
	routelayout.RouterViewComponent
}

var _ routelayout.RouterViewInterface = (*CenterLayout)(nil)

func (c *CenterLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "CenterLayout: Render")

	return app.Div().
		Class("center-layout-content").
		Style("display", "flex").
		Style("justify-content", "center").
		Style("align-items", "center").
		Style("height", "100vh").
		Style("width", "100%").
		Body(
			c.RouterViewComponent.RouterView(),
		)
}
