package layout

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type MainLayout struct {
	app.Compo

	Header  app.UI
	Drawer  app.UI
	Content app.UI
}

func (c *MainLayout) Render() app.UI {
	return app.Div().
		Style("height", "100vh").
		Style("width", "100%").
		Body(
			app.Div().
				Class("main-layout-header").
				Body(c.Header),
			app.Div().
				Class("main-layout-drawer").
				Body(c.Drawer),
			app.Div().
				Class("main-layout-content").
				Body(c.Content),
		)
}
