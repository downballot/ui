package demo

import (
	"context"
	"log/slog"
	"strings"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type SelectPage struct {
	app.Compo

	selectedValues []string
}

var _ app.Composer = (*SelectPage)(nil)
var _ app.Mounter = (*SelectPage)(nil)
var _ app.Navigator = (*SelectPage)(nil)

func (c *SelectPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SelectPage: OnMount")
	c.selectedValues = []string{"option2", "option3"}
}

func (c *SelectPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SelectPage: OnNav")
}

func (c *SelectPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "SelectPage: Render", "selectedValues", c.selectedValues)
	return app.Div().
		Style("padding", "1em").
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Input"),
					myui.Multiselect().
						Label("Multiselect").
						AllowedValue(
							myui.SelectOption{Label: "Option 1", Value: "option1"},
							myui.SelectOption{Label: "Option 2", Value: "option2"},
							myui.SelectOption{Label: "Option 3", Value: "option3"},
						).
						Bind(&c.selectedValues),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Output"),
					app.Div().Text("Multiselect"),
					app.Pre().Text(strings.Join(c.selectedValues, ", ")),
				),
		)
}
