package myui

import (
	"github.com/downballot/ui/slot"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Page() *MyUIPage {
	return &MyUIPage{
		IStyles: map[string]string{},
	}
}

type MyUIPage struct {
	app.Compo
	slot.Slotted

	IClasses []string
	IStyles  map[string]string
}

var _ app.Composer = (*MyUIPage)(nil)

func (c *MyUIPage) Class(class ...string) *MyUIPage {
	c.IClasses = class
	return c
}

func (c *MyUIPage) Style(name, value string) *MyUIPage {
	c.IStyles[name] = value
	return c
}

func (c *MyUIPage) Render() app.UI {
	element := app.Div().
		Class(append([]string{"myui-page"}, c.IClasses...)...)
	for name, value := range c.IStyles {
		element.Style(name, value)
	}
	return element.
		Body(c.SlotContents()...)
}

func (c *MyUIPage) Body(components ...app.UI) *MyUIPage {
	c.Slotted.AddSlotContents(components...)
	return c
}
