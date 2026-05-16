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
	icon  string
	label string
	to    string
}

var _ app.Composer = (*MyUIButton)(nil)

func (c *MyUIButton) Icon(icon string) *MyUIButton {
	c.icon = icon
	return c
}

func (c *MyUIButton) Label(label string) *MyUIButton {
	c.label = label
	return c
}

func (c *MyUIButton) To(to string) *MyUIButton {
	c.to = to
	return c
}

func (c *MyUIButton) On(event string, function func(ctx app.Context, e app.Event)) *MyUIButton {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIButton) Render() app.UI {
	return c.UseEvents.Wrap(
		app.Span().
			Class("myui-button").
			Body(
				app.If(c.icon != "", func() app.UI {
					return Icon().
						Icon(c.icon)
				}),
				app.If(c.to != "", func() app.UI {
					return app.A().
						Href(c.to).
						Body(
							app.Span().
								Class("myui-button-label").
								Text(c.label),
						)
				}).Else(func() app.UI {
					return app.Span().
						Class("myui-button-label").
						Text(c.label)
				}),
			),
	)
}
