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
	TypeValue        string
	LabelValue       string
	PlaceholderValue string
	ValueValue       string
}

var _ app.Composer = (*MyUIInput)(nil)

func (c *MyUIInput) Placeholder(placeholder string) *MyUIInput {
	c.PlaceholderValue = placeholder
	return c
}

func (c *MyUIInput) Type(inputType string) *MyUIInput {
	c.TypeValue = inputType
	return c
}

func (c *MyUIInput) Label(label string) *MyUIInput {
	c.LabelValue = label
	return c
}

func (c *MyUIInput) Value(value string) *MyUIInput {
	c.ValueValue = value
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
			app.If(c.LabelValue != "", func() app.UI {
				return app.Span().
					Class("myui-input-label").
					Text(c.LabelValue)
			}),
			c.UseEvents.Wrap(
				app.Input().
					Class("myui-input-input").
					Type(c.TypeValue).
					Value(c.ValueValue).
					Placeholder(c.PlaceholderValue),
			),
		)
}
