package demo

import (
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type IndexPage struct {
	app.Compo
}

func (c *IndexPage) OnMount(ctx app.Context) {
}

func (c *IndexPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "IndexPage: OnNav")
}

func (c *IndexPage) Render() app.UI {
	return app.Div().
		Style("padding", "1em").
		Body(
			myui.Item().
				Label("Button").
				To("/button"),
			myui.Item().
				Label("Collapse").
				To("/collapse"),
			myui.Item().
				Label("Input").
				To("/input"),
			myui.Item().
				Label("Select").
				To("/select"),
			myui.Item().
				Label("Table").
				To("/table"),
		)
}
