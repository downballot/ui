package myui

import (
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type EmbeddedPage struct {
	IError string
}

func (c *EmbeddedPage) Setup(ctx app.Context) {
	slog.InfoContext(ctx.Context, "EmbeddedPage: Setup")

	var apiError string
	ctx.GetState("api-error", &apiError)
	c.IError = apiError
	slog.InfoContext(ctx.Context, "EmbeddedPage: Setup", "api-error", c.IError)

	ctx.ObserveState("api-error", &c.IError).OnChange(
		func() {
			slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState", "api-error", c.IError)
		})
}

func (c *EmbeddedPage) Wrap(content app.UI) app.UI {
	return Page().
		Body(
			app.If(c.IError != "", func() app.UI {
				return StatusBar().
					Text(c.IError).
					Bad()
			}),
			content,
		)
}
