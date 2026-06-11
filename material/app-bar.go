package material

import (
	"context"
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AppBar struct {
	app.Compo

	Leading    app.UI
	Headline   string
	HeadlineUI app.UI
	Subtitle   string
	SubtitleUI app.UI
	Trailer    app.UI
}

func (c *AppBar) Render() app.UI {
	slog.InfoContext(context.TODO(), "AppBar: Render")

	return app.Div().
		Class("material-app-bar").
		Style("width", "100%").
		Body(
			app.If(c.Leading != nil, func() app.UI {
				return app.Div().
					Class("material-app-bar-leading").
					Body(c.Leading)
			}),
			app.Div().
				Class("material-app-bar-content").
				Body(
					app.Div().
						Class("material-app-bar-headline").
						Body(
							app.If(c.HeadlineUI != nil, func() app.UI {
								return c.HeadlineUI
							}).Else(func() app.UI {
								return app.Div().
									Class("material-app-bar-headline-text").
									Text(c.Headline)
							}),
						),
					app.If(c.Subtitle != "" || c.SubtitleUI != nil, func() app.UI {
						return app.Div().
							Class("material-app-bar-subtitle").
							Body(
								app.If(c.SubtitleUI != nil, func() app.UI {
									return c.SubtitleUI
								}).Else(func() app.UI {
									return app.Div().
										Class("material-app-bar-subtitle-text").
										Text(c.Subtitle)
								}),
							)
					}),
				),
			app.If(c.Trailer != nil, func() app.UI {
				return app.Div().
					Class("material-app-bar-trailer").
					Body(c.Trailer)
			}),
		)
}
