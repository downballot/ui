package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func Button() *MyUIButton {
	return &MyUIButton{}
}

type MyUIButton struct {
	app.Compo
	UseEvents
	IconValue     string
	LabelValue    string
	ToValue       string
	DisabledValue bool
}

var _ app.Composer = (*MyUIButton)(nil)

func (c *MyUIButton) Disabled(disabled bool) *MyUIButton {
	c.DisabledValue = disabled
	return c
}

func (c *MyUIButton) Icon(icon string) *MyUIButton {
	c.IconValue = icon
	return c
}

func (c *MyUIButton) Label(label string) *MyUIButton {
	c.LabelValue = label
	return c
}

func (c *MyUIButton) To(to string) *MyUIButton {
	c.ToValue = to
	return c
}

func (c *MyUIButton) On(event string, function func(ctx app.Context, e app.Event)) *MyUIButton {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIButton) Render() app.UI {
	var element app.UI

	disabledClass := ""
	if c.DisabledValue {
		disabledClass = "disabled"
	}

	element = app.Span().
		Class("myui-button").
		Class(disabledClass).
		Body(
			app.If(c.IconValue != "", func() app.UI {
				return Icon().
					Icon(c.IconValue)
			}),
			app.If(c.ToValue != "", func() app.UI {
				return app.A().
					Href(c.ToValue).
					Body(
						app.Span().
							Class("myui-button-label").
							Text(c.LabelValue),
					)
			}).Else(func() app.UI {
				return app.Span().
					Class("myui-button-label").
					Text(c.LabelValue)
			}),
		)
	if !c.DisabledValue {
		element = c.UseEvents.Wrap(element)
	}
	return element
}
