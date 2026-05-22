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
	IconValue string
}

var _ app.Composer = (*MyUIIcon)(nil)

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
			DataSet("icon", c.IconValue).
			Body(
				app.I().
					Class("fa-solid").
					Class("fa-" + c.IconValue).
					Class("myui-icon").
					Text(c.IconValue),
			),
	)
}
