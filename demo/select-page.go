package demo

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-app-blazar/blazar/blazar"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type SelectPage struct {
	app.Compo

	single   string
	multiple []string
}

var _ app.Composer = (*SelectPage)(nil)
var _ app.Mounter = (*SelectPage)(nil)
var _ app.Navigator = (*SelectPage)(nil)

func (c *SelectPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SelectPage: OnMount")
	c.single = "option2"
	c.multiple = []string{"option2", "option3"}
}

func (c *SelectPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "SelectPage: OnNav")
}

func (c *SelectPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "SelectPage: Render", "multiple", c.multiple)
	return blazar.Page().
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Input"),
					blazar.Select().
						Label("Select").
						AllowedValue(
							blazar.SelectOption{Label: "Option 1", Value: "option1"},
							blazar.SelectOption{Label: "Option 2", Value: "option2"},
							blazar.SelectOption{Label: "Option 3", Value: "option3"},
							blazar.SelectOption{Label: "Option 4", Value: "option4", Disabled: true},
						).
						Bind(&c.single),
					blazar.Multiselect().
						Label("Multiselect").
						AllowedValue(
							blazar.SelectOption{Label: "Option 1", Value: "option1"},
							blazar.SelectOption{Label: "Option 2", Value: "option2"},
							blazar.SelectOption{Label: "Option 3", Value: "option3"},
							blazar.SelectOption{Label: "Option 4", Value: "option4", Disabled: true},
						).
						Bind(&c.multiple),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Output"),
					app.Div().Text("Select"),
					app.Pre().Text(c.single),
					app.Div().Text("Multiselect"),
					app.Pre().Text(strings.Join(c.multiple, ", ")),
				),
		)
}
