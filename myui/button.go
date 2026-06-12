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
	IFlat     bool
	IIcon     string
	ILabel    string
	ITo       string
	IRound    bool
	IDisabled bool
}

var _ app.Composer = (*MyUIButton)(nil)

func (c *MyUIButton) Disabled(disabled bool) *MyUIButton {
	c.IDisabled = disabled
	return c
}

func (c *MyUIButton) Flat(flat bool) *MyUIButton {
	c.IFlat = flat
	return c
}

func (c *MyUIButton) Icon(icon string) *MyUIButton {
	c.IIcon = icon
	return c
}

func (c *MyUIButton) Label(label string) *MyUIButton {
	c.ILabel = label
	return c
}

func (c *MyUIButton) Round(round bool) *MyUIButton {
	c.IRound = round
	return c
}

func (c *MyUIButton) To(to string) *MyUIButton {
	c.ITo = to
	return c
}

func (c *MyUIButton) On(event string, function func(ctx app.Context, e app.Event)) *MyUIButton {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIButton) Render() app.UI {
	disabledClass := ""
	if c.IDisabled {
		disabledClass = "disabled"
	}

	flatClass := ""
	if c.IFlat {
		flatClass = "flat"
	}

	roundClass := ""
	if c.IRound {
		roundClass = "round"
	}

	var body []app.UI
	if c.IIcon != "" {
		body = append(body, Icon().
			Class("myui-button__icon").
			Icon(c.IIcon),
		)
	}
	if c.ILabel != "" {
		body = append(body, app.Span().
			Class("myui-button__label").
			Body(
				app.Span().
					Text(c.ILabel),
			),
		)
	}

	var innerElement app.UI
	if c.ITo == "" || c.IDisabled {
		innerElement = app.Span().
			Class("myui-button__content").
			Body(body...)
	} else {
		innerElement = app.A().
			Class("myui-button__content").
			Href(c.ITo).
			Body(body...)
	}

	var element app.UI
	element = app.Span().
		Class("myui-button", disabledClass, roundClass, flatClass).
		TabIndex(0).
		Role("button").
		Body(innerElement).
		OnKeyPress(func(ctx app.Context, e app.Event) {
			if e.Get("key").String() == "Enter" || e.Get("key").String() == " " {
				e.Get("target").Call("click")
			}
		})
	if !c.IDisabled {
		element = c.UseEvents.Wrap(element)
	}
	return element
}
