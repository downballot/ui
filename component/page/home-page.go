package page

import (
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type HomePage struct {
	app.Compo
	myui.EmbeddedPage
}

func (c *HomePage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "HomePage: OnNav: Navigating to organization page")
	ctx.Navigate("/organization")
}

func (c *HomePage) Render() app.UI {
	return c.EmbeddedPage.Wrap(
		app.Span().
			Text("Home page (this should not be visible)"),
	)
}
