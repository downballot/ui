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
	IIcon string
	IName string
	ITo   string
}

var _ app.Composer = (*MyUIItem)(nil)

func (c *MyUIItem) Icon(icon string) *MyUIItem {
	c.IIcon = icon
	return c
}

func (c *MyUIItem) Name(name string) *MyUIItem {
	c.IName = name
	return c
}

func (c *MyUIItem) To(to string) *MyUIItem {
	c.ITo = to
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
			Href(c.ITo).
			Body(
				app.Span().
					Class("myui-item__icon").
					Body(
						app.If(c.IIcon != "", func() app.UI {
							return Icon().Icon(c.IIcon)
						}),
					),
				app.Span().
					Class("myui-item__name").
					Text(c.IName),
			),
	)
}
