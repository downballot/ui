package myui

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func StatusBar() *MyUIStatusBar {
	return &MyUIStatusBar{}
}

type MyUIStatusBar struct {
	app.Compo

	TextValue string
	GoodValue *bool
}

var _ app.Composer = (*MyUIStatusBar)(nil)

func (c *MyUIStatusBar) Text(text string) *MyUIStatusBar {
	c.TextValue = text
	return c
}

func (c *MyUIStatusBar) Good() *MyUIStatusBar {
	c.GoodValue = new(bool)
	*c.GoodValue = true
	return c
}

func (c *MyUIStatusBar) Bad() *MyUIStatusBar {
	c.GoodValue = new(bool)
	*c.GoodValue = false
	return c
}

func (c *MyUIStatusBar) Neutral() *MyUIStatusBar {
	c.GoodValue = nil
	return c
}

func (c *MyUIStatusBar) Render() app.UI {
	goodClass := ""
	goodIcon := "circle-info"
	if c.GoodValue != nil {
		if *c.GoodValue {
			goodClass = "good"
			goodIcon = "circle-check"
		} else {
			goodClass = "bad"
			goodIcon = "circle-exclamation"
		}
	}

	return app.Div().
		Class("myui-statusbar").
		Class(goodClass).
		Body(
			app.If(goodIcon != "", func() app.UI {
				return Icon().
					Icon(goodIcon)
			}),
			app.Span().
				Class("myui-button-label").
				Text(c.TextValue),
		)
}
