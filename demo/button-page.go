package demo

import (
	"log/slog"

	"github.com/go-app-blazar/blazar/blazar"
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

	return blazar.Page().
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Navigation"),
					app.Div().Body(
						blazar.Button().
							Label("Default").
							To("/"),
						blazar.Button().
							Label("Flat").
							Flat(true).
							To("/"),
						blazar.Button().
							Label("Round").
							Round(true).
							To("/"),
						blazar.Button().
							Label("Disabled").
							Disabled(true).
							To("/"),
					),
					app.Div().Body(
						blazar.Button().
							Label("Default").
							Icon("home").
							To("/"),
						blazar.Button().
							Label("Flat").
							Flat(true).
							Icon("home").
							To("/"),
						blazar.Button().
							Label("Round").
							Round(true).
							Icon("home").
							To("/"),
						blazar.Button().
							Label("Disabled").
							Disabled(true).
							Icon("home").
							To("/"),
					),
					app.Div().Body(
						blazar.Button().
							Icon("home").
							To("/"),
						blazar.Button().
							Flat(true).
							Icon("home").
							To("/"),
						blazar.Button().
							Round(true).
							Icon("home").
							To("/"),
						blazar.Button().
							Disabled(true).
							Icon("home").
							To("/"),
					),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Action"),
					app.Div().Body(
						blazar.Button().
							Label("Default").
							On("click", clickFunction),
						blazar.Button().
							Label("Flat").
							Flat(true).
							On("click", clickFunction),
						blazar.Button().
							Label("Round").
							Round(true).
							On("click", clickFunction),
						blazar.Button().
							Label("Disabled").
							Disabled(true).
							On("click", clickFunction),
					),
					app.Div().Body(
						blazar.Button().
							Label("Default").
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Label("Flat").
							Flat(true).
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Label("Round").
							Round(true).
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Label("Disabled").
							Disabled(true).
							Icon("home").
							On("click", clickFunction),
					),
					app.Div().Body(
						blazar.Button().
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Flat(true).
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Round(true).
							Icon("home").
							On("click", clickFunction),
						blazar.Button().
							Disabled(true).
							Icon("home").
							On("click", clickFunction),
					),
				),
		)
}
