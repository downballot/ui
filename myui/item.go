package myui

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Item() *MyUIItem {
	return &MyUIItem{}
}

type MyUIItem struct {
	app.Compo
	UseEvents
	IconValue string
	NameValue string
	ToValue   string
}

var _ app.Composer = (*MyUIItem)(nil)

func (c *MyUIItem) Icon(icon string) *MyUIItem {
	c.IconValue = icon
	return c
}

func (c *MyUIItem) Name(name string) *MyUIItem {
	c.NameValue = name
	return c
}

func (c *MyUIItem) To(to string) *MyUIItem {
	c.ToValue = to
	return c
}

func (c *MyUIItem) On(event string, function func(ctx app.Context, e app.Event)) *MyUIItem {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIItem) Render() app.UI {
	return c.UseEvents.Wrap(
		app.A().
			Class("myui-item").
			Href(c.ToValue).
			Body(
				app.Span().
					Class("myui-item__icon").
					Body(
						app.If(c.IconValue != "", func() app.UI {
							return Icon().Icon(c.IconValue)
						}),
					),
				app.Span().
					Class("myui-item__name").
					Text(c.NameValue),
			),
	)
}
