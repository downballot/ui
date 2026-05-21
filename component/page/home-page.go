package page

import (
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type HomePage struct {
	app.Compo
}

func (c *HomePage) OnNav(ctx app.Context) {
	ctx.Navigate("/organization")
}

func (c *HomePage) Render() app.UI {
	return myui.Page().
		Body(
			app.Span().
				Text("Home page (this should not be visible)"),
		)
}
