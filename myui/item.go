package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func Item() *MyUIItem {
	return &MyUIItem{}
}

type MyUIItem struct {
	app.Compo
	UseEvents
	icon string
	name string
	to   string
}

var _ app.Composer = (*MyUIItem)(nil)

func (c *MyUIItem) Icon(icon string) *MyUIItem {
	c.icon = icon
	return c
}

func (c *MyUIItem) Name(name string) *MyUIItem {
	c.name = name
	return c
}

func (c *MyUIItem) To(to string) *MyUIItem {
	c.to = to
	return c
}

func (c *MyUIItem) Render() app.UI {
	return c.UseEvents.Wrap(
		app.Div().
			Class("myui-item").
			Body(
				app.If(c.to != "", func() app.UI {
					return app.A().Href(c.to).
						Body(
							app.Span().
								Class("myui-item-icon").
								Body(
									app.If(c.icon != "", func() app.UI {
										return Icon().Icon(c.icon)
									}),
								),
							app.Span().
								Class("myui-item-name").
								Text(c.name),
						)
				}).ElseSlice(func() []app.UI {
					return []app.UI{
						app.Span().
							Class("myui-item-icon").
							Body(
								app.If(c.icon != "", func() app.UI {
									return Icon().Icon(c.icon)
								}),
							),
						app.Span().
							Class("myui-item-name").
							Text(c.name),
					}
				}),
			),
	)
}
