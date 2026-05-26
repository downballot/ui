package myui

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Icon() *MyUIIcon {
	return &MyUIIcon{}
}

type MyUIIcon struct {
	app.Compo
	UseEvents
	ClassValue string
	IconValue  string
}

var _ app.Composer = (*MyUIIcon)(nil)

func (c *MyUIIcon) Class(class string) *MyUIIcon {
	c.ClassValue = class
	return c
}

func (c *MyUIIcon) Icon(icon string) *MyUIIcon {
	c.IconValue = icon
	return c
}

func (c *MyUIIcon) On(event string, function func(ctx app.Context, e app.Event)) *MyUIIcon {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIIcon) Render() app.UI {
	return c.UseEvents.Wrap(
		app.Span().
			Class("myui-icon").
			Class(c.ClassValue).
			DataSet("icon", c.IconValue).
			Body(
				app.I().
					Class("myui-icon__icon").
					Class("fa-solid").
					Class("fa-" + c.IconValue),
			),
	)
}
