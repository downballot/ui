package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func Icon() *MyUIIcon {
	return &MyUIIcon{}
}

type MyUIIcon struct {
	app.Compo
	UseEvents
	icon string
}

var _ app.Composer = (*MyUIIcon)(nil)

func (c *MyUIIcon) Icon(icon string) *MyUIIcon {
	c.icon = icon
	return c
}

func (c *MyUIIcon) Render() app.UI {
	return c.UseEvents.Wrap(
		app.Span().
			Class("myui-icon").
			Body(
				app.I().
					Class("fa-solid").
					Class("fa-" + c.icon).
					Class("myui-icon"),
			),
	)
}
