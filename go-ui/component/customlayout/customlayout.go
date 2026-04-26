package customlayout

import (
	"log/slog"

	"github.com/downballot/ui/go-ui/component/layout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type DownballotLayout struct {
	app.Compo

	Content app.UI
}

func (c *DownballotLayout) OnNav(ctx app.Context) {
	var apiToken string
	ctx.GetState("api-token", &apiToken)
	slog.InfoContext(ctx.Context, "State", "api-token", apiToken)
	if apiToken == "" {
		ctx.Navigate("/login")
	}
}

func (c *DownballotLayout) Render() app.UI {
	return &layout.MainLayout{
		Content: c.Content,
		Header: app.Div().Body(
			app.H1().Text("Downballot"),
			app.A().Href("/organization/1").Text("Org1"),
			app.A().Href("/organization/2").Text("Org2"),
			app.A().Href("/organization/3").Text("Org3"),
		),
		Drawer: app.Div().Body(
			app.Ul().Body(
				app.Li().Body(
					app.A().Href("/login").Text("Login"),
				),
				app.Li().Body(
					app.A().Href("/organization").Text("Organization"),
				),
				app.Li().Body(
					app.A().Href("/profile").Text("Profile"),
				),
			),
		),
	}
}

func (c *DownballotLayout) WithComponent(component app.Composer) app.Composer {
	var output DownballotLayout
	output.Content = component
	return &output
}
