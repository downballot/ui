package demo

import (
	"log/slog"

	"github.com/go-app-blazar/blazar/blazar"
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
	return blazar.Page().
		Body(
			blazar.Item().
				Label("Button").
				To("/button"),
			blazar.Item().
				Label("Collapse").
				To("/collapse"),
			blazar.Item().
				Label("Form").
				To("/form"),
			blazar.Item().
				Label("Input").
				To("/input"),
			blazar.Item().
				Label("Select").
				To("/select"),
			blazar.Item().
				Label("Table").
				To("/table"),
		)
}
