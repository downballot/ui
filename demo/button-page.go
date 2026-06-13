package demo

import (
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type ButtonPage struct {
	app.Compo
}

func (c *ButtonPage) OnMount(ctx app.Context) {
}

func (c *ButtonPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "ButtonPage: OnNav")
}

func (c *ButtonPage) Render() app.UI {
	clickFunction := func(ctx app.Context, e app.Event) {
		slog.InfoContext(ctx.Context, "ButtonPage: Click", "event", e)
		app.Window().Call("alert", "Button clicked")
	}

	return myui.Page().
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Navigation"),
					app.Div().Body(
						myui.Button().
							Label("Default").
							To("/"),
						myui.Button().
							Label("Flat").
							Flat(true).
							To("/"),
						myui.Button().
							Label("Round").
							Round(true).
							To("/"),
						myui.Button().
							Label("Disabled").
							Disabled(true).
							To("/"),
					),
					app.Div().Body(
						myui.Button().
							Label("Default").
							Icon("home").
							To("/"),
						myui.Button().
							Label("Flat").
							Flat(true).
							Icon("home").
							To("/"),
						myui.Button().
							Label("Round").
							Round(true).
							Icon("home").
							To("/"),
						myui.Button().
							Label("Disabled").
							Disabled(true).
							Icon("home").
							To("/"),
					),
					app.Div().Body(
						myui.Button().
							Icon("home").
							To("/"),
						myui.Button().
							Flat(true).
							Icon("home").
							To("/"),
						myui.Button().
							Round(true).
							Icon("home").
							To("/"),
						myui.Button().
							Disabled(true).
							Icon("home").
							To("/"),
					),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Action"),
					app.Div().Body(
						myui.Button().
							Label("Default").
							On("click", clickFunction),
						myui.Button().
							Label("Flat").
							Flat(true).
							On("click", clickFunction),
						myui.Button().
							Label("Round").
							Round(true).
							On("click", clickFunction),
						myui.Button().
							Label("Disabled").
							Disabled(true).
							On("click", clickFunction),
					),
					app.Div().Body(
						myui.Button().
							Label("Default").
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Label("Flat").
							Flat(true).
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Label("Round").
							Round(true).
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Label("Disabled").
							Disabled(true).
							Icon("home").
							On("click", clickFunction),
					),
					app.Div().Body(
						myui.Button().
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Flat(true).
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Round(true).
							Icon("home").
							On("click", clickFunction),
						myui.Button().
							Disabled(true).
							Icon("home").
							On("click", clickFunction),
					),
				),
		)
}
