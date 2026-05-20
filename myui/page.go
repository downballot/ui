package myui

import (
	"github.com/downballot/ui/slot"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Page() *MyUIPage {
	return &MyUIPage{}
}

type MyUIPage struct {
	app.Compo
	slot.Slotted
}

var _ app.Composer = (*MyUIPage)(nil)

func (c *MyUIPage) Render() app.UI {
	return app.Div().
		Class("myui-page").
		Body(app.FilterUIElems(c.SlotContents()...)...)
}

func (c *MyUIPage) Body(components ...app.UI) *MyUIPage {
	c.Slotted.AddSlotContents(components...)
	return c
}
