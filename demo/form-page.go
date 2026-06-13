package demo

import (
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type FormPage struct {
	app.Compo

	name string
}

func (c *FormPage) OnMount(ctx app.Context) {
	c.name = "Monkey D. Luffy"
}

func (c *FormPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "FormPage: OnNav")
}

func (c *FormPage) Render() app.UI {
	cancelFunction := func(ctx app.Context) {
		app.Window().Call("alert", "Cancel called")
	}
	submitFunction := func(ctx app.Context) {
		app.Window().Call("alert", "Submit called")
	}

	action1Function := func(ctx app.Context) {
		app.Window().Call("alert", "Action 1 called")
	}
	action2Function := func(ctx app.Context) {
		app.Window().Call("alert", "Action 2 called")
	}

	return app.Div().
		Style("padding", "1em").
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Default"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With cancel"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						CancelFunction(cancelFunction),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With submit"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						SubmitFunction(submitFunction),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With both submit and cancel"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						CancelFunction(cancelFunction).
						SubmitFunction(submitFunction),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With custom labels"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						CancelLabel("Please cancel this").
						CancelIcon("trash").
						CancelFunction(cancelFunction).
						SubmitLabel("Please submit this").
						SubmitIcon("save").
						SubmitFunction(submitFunction),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With custom actions only"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						Action(
							myui.FormAction{Name: "Action 1", Icon: "person", Function: action1Function},
							myui.FormAction{Name: "Action 2", Icon: "gear", Function: action2Function},
						),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With custom actions and default actions"),
					myui.Form().
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						CancelLabel("Please cancel this").
						CancelIcon("trash").
						CancelFunction(cancelFunction).
						SubmitLabel("Please submit this").
						SubmitIcon("save").
						SubmitFunction(submitFunction).
						Action(
							myui.FormAction{Name: "Action 1", Icon: "person", Function: action1Function},
							myui.FormAction{Name: "Action 2", Icon: "gear", Function: action2Function},
						),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("With actions to the left"),
					myui.Form().
						Spacer(false).
						Body(
							myui.Input[string]().
								Label("Name").
								Bind(&c.name),
						).
						SubmitLabel("Please submit this").
						SubmitIcon("save").
						SubmitFunction(submitFunction).
						Action(
							myui.FormAction{Name: "Action 1", Icon: "person", Function: action1Function},
							myui.FormAction{Name: "Action 2", Icon: "gear", Function: action2Function},
						),
				),
		)
}
