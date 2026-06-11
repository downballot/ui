package layout

import (
	"context"
	"log/slog"

	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type CenterLayout struct {
	app.Compo
	router.RouterViewComponent
}

var _ router.RouterViewInterface = (*CenterLayout)(nil)

func (c *CenterLayout) Render() app.UI {
	slog.InfoContext(context.TODO(), "CenterLayout: Render")

	return app.Div().
		Class("center-layout").
		Style("display", "flex").
		Style("justify-content", "center").
		Style("align-items", "center").
		Style("height", "100vh").
		Style("width", "100%").
		Body(
			c.RouterViewComponent.RouterView(),
		)
}
