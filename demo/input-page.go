package demo

import (
	"fmt"
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type InputPage struct {
	app.Compo

	stringValue string
	intValue    int
	floatValue  float64
	uintValue   uint
	boolValue   bool
}

func (c *InputPage) OnMount(ctx app.Context) {
	c.stringValue = "Hello, World!"
	c.intValue = 123
	c.floatValue = 123.456
	c.uintValue = 123
	c.boolValue = true
}

func (c *InputPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "InputPage: OnNav")
}

func (c *InputPage) Render() app.UI {
	return app.Div().
		Style("padding", "1em").
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Input"),
					myui.Input[string]().
						Label("string").
						Bind(&c.stringValue),
					myui.Input[int]().
						Label("int").
						Bind(&c.intValue),
					myui.Input[float64]().
						Label("float").
						Bind(&c.floatValue),
					myui.Input[uint]().
						Label("uint").
						Bind(&c.uintValue),
					myui.Input[bool]().
						Label("bool").
						Bind(&c.boolValue),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Output"),
					app.Div().Text("string"),
					app.Pre().Text(c.stringValue),
					app.Div().Text("int"),
					app.Pre().Text(fmt.Sprintf("%d", c.intValue)),
					app.Div().Text("float"),
					app.Pre().Text(fmt.Sprintf("%f", c.floatValue)),
					app.Div().Text("uint"),
					app.Pre().Text(fmt.Sprintf("%d", c.uintValue)),
					app.Div().Text("bool"),
					app.Pre().Text(fmt.Sprintf("%t", c.boolValue)),
				),
		)
}
