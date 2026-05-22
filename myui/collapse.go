package myui

import (
	"context"
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Collapse() *MyUICollapse {
	return &MyUICollapse{}
}

type MyUICollapse struct {
	app.Compo
	UseEvents
	LabelValue    string
	DisabledValue bool
	OpenValue     bool
	BindValue     *bool
	BodyValue     []app.UI
}

var _ app.Composer = (*MyUICollapse)(nil)
var _ app.Updater = (*MyUICollapse)(nil)

func (c *MyUICollapse) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate")
	slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate", "OpenValue", c.OpenValue)
	slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate", "BindValue", c.BindValue)
	if c.BindValue != nil {
		slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate", "*BindValue", *c.BindValue)
	}

	if c.BindValue != nil {
		c.OpenValue = *c.BindValue
	}
}
func (c *MyUICollapse) Disabled(disabled bool) *MyUICollapse {
	c.DisabledValue = disabled
	return c
}

func (c *MyUICollapse) Open(open bool) *MyUICollapse {
	c.OpenValue = open
	return c
}

func (c *MyUICollapse) Label(label string) *MyUICollapse {
	c.LabelValue = label
	return c
}

func (c *MyUICollapse) Body(body ...app.UI) *MyUICollapse {
	c.BodyValue = body
	return c
}

func (c *MyUICollapse) On(event string, function func(ctx app.Context, e app.Event)) *MyUICollapse {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUICollapse) Bind(variable *bool) *MyUICollapse {
	c.BindValue = variable
	return c
}

func (c *MyUICollapse) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUICollapse: Render", "OpenValue", c.OpenValue)
	slog.InfoContext(context.TODO(), "MyUICollapse: Render", "BindValue", c.BindValue)
	if c.BindValue != nil {
		slog.InfoContext(context.TODO(), "MyUICollapse: Render", "*BindValue", *c.BindValue)
	}

	var element app.UI

	disabledClass := ""
	if c.DisabledValue {
		disabledClass = "disabled"
	}

	closedIcon := "chevron-down"
	closedClass := ""
	if !c.OpenValue {
		closedIcon = "chevron-right"
		closedClass = "closed"
	}

	element = app.Div().
		Class("myui-collapse").
		Class(disabledClass).
		Class(closedClass).
		Body(
			app.Div().
				Class("myui-collapse-label").
				Style("cursor", "pointer").
				Body(
					app.Span().
						Text(c.LabelValue),
					app.Span().Style("flex", "1"),
					app.Span().
						Class("myui-collapse-icon").
						Body(
							Icon().
								Icon(closedIcon),
						),
				).
				On("click", func(ctx app.Context, e app.Event) {
					c.OpenValue = !c.OpenValue
					if c.BindValue != nil {
						*c.BindValue = c.OpenValue
					}
					slog.InfoContext(ctx.Context, "Collapse: OnClick", "OpenValue", c.OpenValue)
					ctx.Update()
				}),
			app.If(c.OpenValue, func() app.UI {
				return app.Div().
					Class("myui-collapse-content").
					Body(c.BodyValue...)
			}),
		)
	if !c.DisabledValue {
		element = c.UseEvents.Wrap(element)
	}
	return element
}
