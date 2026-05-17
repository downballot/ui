package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func Input() *MyUIInput {
	return &MyUIInput{}
}

type MyUIInput struct {
	app.Compo
	UseEvents
	inputType   string
	label       string
	placeholder string
	value       string
}

var _ app.Composer = (*MyUIInput)(nil)

func (c *MyUIInput) Placeholder(placeholder string) *MyUIInput {
	c.placeholder = placeholder
	return c
}

func (c *MyUIInput) Type(inputType string) *MyUIInput {
	c.inputType = inputType
	return c
}

func (c *MyUIInput) Label(label string) *MyUIInput {
	c.label = label
	return c
}

func (c *MyUIInput) Value(value string) *MyUIInput {
	c.value = value
	return c
}

func (c *MyUIInput) On(event string, function func(ctx app.Context, e app.Event)) *MyUIInput {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIInput) Render() app.UI {
	return app.Span().
		Class("myui-input").
		Body(
			app.If(c.label != "", func() app.UI {
				return app.Span().
					Class("myui-input-label").
					Text(c.label)
			}),
			c.UseEvents.Wrap(
				app.Input().
					Class("myui-input-input").
					Type(c.inputType).
					Value(c.value).
					Placeholder(c.placeholder),
			),
		)
}
