package layout

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type CenterLayout struct {
	app.Compo

	Content app.UI
}

func (c *CenterLayout) Render() app.UI {
	return app.Div().
		Class("center-layout-content").
		Style("display", "flex").
		Style("justify-content", "center").
		Style("align-items", "center").
		Style("height", "100vh").
		Style("width", "100%").
		Body(
			c.Content,
		)
}
