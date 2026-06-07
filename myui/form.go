package myui

import (
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Form() *MyUIForm {
	return &MyUIForm{}
}

type MyUIForm struct {
	app.Compo
	UseEvents

	IBody           []app.UI
	ICancelFunction func(ctx app.Context)
	ICancelLabel    string
	ICancelIcon     string
	ISubmitFunction func(ctx app.Context)
	ISubmitLabel    string
	ISubmitIcon     string
	IActions        []FormAction
}

type FormAction struct {
	Name     string
	Icon     string
	To       func() string
	Function func(ctx app.Context)
}

var _ app.Composer = (*MyUIForm)(nil)

func (c *MyUIForm) Action(actions ...FormAction) *MyUIForm {
	c.IActions = append(c.IActions, actions...)
	return c
}

func (c *MyUIForm) CancelFunction(function func(ctx app.Context)) *MyUIForm {
	c.ICancelFunction = function
	return c
}

func (c *MyUIForm) CancelLabel(label string) *MyUIForm {
	c.ICancelLabel = label
	return c
}

func (c *MyUIForm) CancelIcon(icon string) *MyUIForm {
	c.ICancelIcon = icon
	return c
}

func (c *MyUIForm) SubmitFunction(function func(ctx app.Context)) *MyUIForm {
	c.ISubmitFunction = function
	return c
}

func (c *MyUIForm) SubmitIcon(icon string) *MyUIForm {
	c.ISubmitIcon = icon
	return c
}

func (c *MyUIForm) SubmitLabel(label string) *MyUIForm {
	c.ISubmitLabel = label
	return c
}

func (c *MyUIForm) Body(body ...app.UI) *MyUIForm {
	c.IBody = append(c.IBody, body...)
	return c
}

func (c *MyUIForm) On(event string, function func(ctx app.Context, e app.Event)) *MyUIForm {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIForm) Render() app.UI {
	return app.Div().
		Class("myui-form").
		Body(
			c.UseEvents.Wrap(
				app.Div().
					Class("myui-form__form").
					On("keypress", func(ctx app.Context, e app.Event) {
						ctx.PreventUpdate()

						if e.Get("key").String() == "Enter" {
							slog.InfoContext(ctx.Context, "MyUIForm: Keypress", "key", e.Get("key").String())
							if c.ISubmitFunction != nil {
								c.ISubmitFunction(ctx)
							}
						}
					}).
					Body(
						c.IBody...,
					),
			),
			app.Div().
				Class("myui-form__actions").
				Body(
					app.If(c.ICancelFunction != nil, func() app.UI {
						return Button().
							Flat(true).
							Label(func() string {
								if c.ICancelLabel != "" {
									return c.ICancelLabel
								}
								return "Cancel"
							}()).
							Icon(c.ICancelIcon).
							On("click", func(ctx app.Context, e app.Event) {
								c.ICancelFunction(ctx)
							})
					}),
					app.Span().Style("flex", "1"),
					app.Range(c.IActions).Slice(func(i int) app.UI {
						action := c.IActions[i]
						return Button().
							Flat(false).
							Label(action.Name).
							Icon(action.Icon).
							To(func() string {
								if action.To != nil {
									return action.To()
								}
								return ""
							}()).
							On("click", func(ctx app.Context, e app.Event) {
								if action.Function == nil {
									ctx.PreventUpdate()
									return
								}
								action.Function(ctx)
							})
					}),
					app.If(c.ISubmitFunction != nil, func() app.UI {
						return Button().
							Flat(false).
							Label(func() string {
								if c.ISubmitLabel != "" {
									return c.ISubmitLabel
								}
								return "Submit"
							}()).
							Icon(c.ISubmitIcon).
							On("click", func(ctx app.Context, e app.Event) {
								c.ISubmitFunction(ctx)
							})
					}),
				),
		)
}
