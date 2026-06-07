package myui

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Button() *MyUIButton {
	return &MyUIButton{}
}

type MyUIButton struct {
	app.Compo
	UseEvents
	FlatValue     bool
	IconValue     string
	LabelValue    string
	ToValue       string
	RoundValue    bool
	DisabledValue bool
}

var _ app.Composer = (*MyUIButton)(nil)

func (c *MyUIButton) Disabled(disabled bool) *MyUIButton {
	c.DisabledValue = disabled
	return c
}

func (c *MyUIButton) Flat(flat bool) *MyUIButton {
	c.FlatValue = flat
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

func (c *MyUIButton) Round(round bool) *MyUIButton {
	c.RoundValue = round
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

	flatClass := ""
	if c.FlatValue {
		flatClass = "flat"
	}

	roundClass := ""
	if c.RoundValue {
		roundClass = "round"
	}

	element = app.A().
		Class("myui-button").
		Class(disabledClass).
		Class(roundClass).
		Class(flatClass).
		TabIndex(0).
		Role("button").
		Href(c.ToValue).
		Body(
			app.Div().
				Class("myui-button__content").
				Body(
					app.If(c.IconValue != "", func() app.UI {
						return Icon().
							Class("myui-button__icon").
							Icon(c.IconValue)
					}),
					app.If(c.LabelValue != "", func() app.UI {
						return app.Div().
							Class("myui-button__label").
							Body(
								app.Span().
									Text(c.LabelValue),
							)
					}),
				),
		).
		OnKeyPress(func(ctx app.Context, e app.Event) {
			if e.Get("key").String() == "Enter" || e.Get("key").String() == " " {
				e.Get("target").Call("click")
			}
		})
	if !c.DisabledValue {
		element = c.UseEvents.Wrap(element)
	}
	return element
}
