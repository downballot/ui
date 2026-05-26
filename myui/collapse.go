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
	ILabel    string
	IDisabled bool
	ISummary  []app.UI
	IBody     []app.UI
	BindOpen  *bool
}

var _ app.Composer = (*MyUICollapse)(nil)
var _ app.Updater = (*MyUICollapse)(nil)

func (c *MyUICollapse) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate")
	if c.BindOpen != nil {
		slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate", "*BindValue", *c.BindOpen)
	} else {
		slog.InfoContext(ctx.Context, "MyUICollapse: OnUpdate: BindOpen is nil.")
	}
}
func (c *MyUICollapse) Disabled(disabled bool) *MyUICollapse {
	c.IDisabled = disabled
	return c
}

func (c *MyUICollapse) Open(open bool) *MyUICollapse {
	if c.BindOpen == nil {
		c.BindOpen = new(bool)
	}
	*c.BindOpen = open
	return c
}

func (c *MyUICollapse) Label(label string) *MyUICollapse {
	c.ILabel = label
	return c
}

func (c *MyUICollapse) Summary(summary ...app.UI) *MyUICollapse {
	c.ISummary = summary
	return c
}

func (c *MyUICollapse) Body(body ...app.UI) *MyUICollapse {
	c.IBody = body
	return c
}

func (c *MyUICollapse) On(event string, function func(ctx app.Context, e app.Event)) *MyUICollapse {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUICollapse) Bind(variable *bool) *MyUICollapse {
	c.BindOpen = variable
	return c
}

func (c *MyUICollapse) Render() app.UI {
	open := false
	if c.BindOpen != nil {
		open = *c.BindOpen
	}
	slog.InfoContext(context.TODO(), "MyUICollapse: Render", "BindOpen", c.BindOpen, "open", open)

	var element app.UI

	disabledClass := ""
	if c.IDisabled {
		disabledClass = "disabled"
	}

	closedIcon := "chevron-down"
	closedClass := ""
	if !open {
		closedIcon = "chevron-right"
		closedClass = "closed"
	}

	element = app.Div().
		Class("myui-collapse").
		Class(disabledClass).
		Class(closedClass).
		Body(
			app.Div().
				Class("myui-collapse__top").
				Style("cursor", "pointer").
				Body(
					app.Span().
						Class("myui-collapse__label").
						Text(c.ILabel),
					app.If(len(c.ISummary) > 0, func() app.UI {
						return app.Span().
							Class("myui-collapse__summary").
							Body(c.ISummary...)
					}),
					app.Span().Style("flex", "1"),
					app.Span().
						Class("myui-collapse-icon").
						Body(
							Icon().
								Icon(closedIcon),
						),
				).
				On("click", func(ctx app.Context, e app.Event) {
					open = !open
					if c.BindOpen != nil {
						*c.BindOpen = open
					}
					slog.InfoContext(ctx.Context, "Collapse: OnClick", "BindOpen", c.BindOpen, "open", open)
					ctx.Update()
				}),
			app.If(open, func() app.UI {
				return app.Div().
					Class("myui-collapse-content").
					Body(c.IBody...)
			}),
		)
	if !c.IDisabled {
		element = c.UseEvents.Wrap(element)
	}
	return element
}
