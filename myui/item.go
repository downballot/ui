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

func (c *MyUIItem) Render() app.UI {
	return c.UseEvents.Wrap(
		app.Div().
			Class("myui-item").
			Body(
				app.If(c.ToValue != "", func() app.UI {
					return app.A().Href(c.ToValue).
						Body(
							app.Span().
								Class("myui-item-icon").
								Body(
									app.If(c.IconValue != "", func() app.UI {
										return Icon().Icon(c.IconValue)
									}),
								),
							app.Span().
								Class("myui-item-name").
								Text(c.NameValue),
						)
				}).ElseSlice(func() []app.UI {
					return []app.UI{
						app.Span().
							Class("myui-item-icon").
							Body(
								app.If(c.IconValue != "", func() app.UI {
									return Icon().Icon(c.IconValue)
								}),
							),
						app.Span().
							Class("myui-item-name").
							Text(c.NameValue),
					}
				}),
			),
	)
}
