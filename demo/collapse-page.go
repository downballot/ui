package demo

import (
	"fmt"
	"log/slog"

	"github.com/go-app-blazar/blazar/blazar"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type CollapsePage struct {
	app.Compo

	defaultOpen   bool
	defaultClosed bool
}

func (c *CollapsePage) OnMount(ctx app.Context) {
	c.defaultOpen = true
	c.defaultClosed = false
}

func (c *CollapsePage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "CollapsePage: OnNav")
}

func (c *CollapsePage) Render() app.UI {
	return blazar.Page().
		Body(
			blazar.Collapse().
				Bind(&c.defaultOpen).
				Label("Default Open").
				SummaryText("Current state: "+fmt.Sprintf("%t", c.defaultOpen)).
				Body(
					app.Div().Text("This is the body of the collapse."),
				),
			blazar.Collapse().
				Bind(&c.defaultClosed).
				Label("Default Closed").
				SummaryText("Current state: "+fmt.Sprintf("%t", c.defaultClosed)).
				Body(
					app.Div().Text("This is the body of the collapse."),
				),
		)
}
