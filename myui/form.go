package myui

import (
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Form() *MyUIForm {
	return &MyUIForm{
		ISpacer: true,
	}
}

type MyUIForm struct {
	app.Compo
	UseEvents

	IClasses []string
	IStyles  map[string]string

	ISpacer         bool
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
	To       string
	Function func(ctx app.Context)
}

var _ app.Composer = (*MyUIForm)(nil)

func (c *MyUIForm) Class(class ...string) *MyUIForm {
	c.IClasses = append(c.IClasses, class...)
	return c
}

func (c *MyUIForm) Spacer(spacer bool) *MyUIForm {
	c.ISpacer = spacer
	return c
}

func (c *MyUIForm) Style(name, value string) *MyUIForm {
	if c.IStyles == nil {
		c.IStyles = make(map[string]string)
	}
	c.IStyles[name] = value
	return c
}

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
	element := app.Div().
		Class("myui-form").
		Class(c.IClasses...).
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
					app.If(c.ISpacer, func() app.UI {
						return app.Span().Style("flex", "1")
					}),
					app.Range(c.IActions).Slice(func(i int) app.UI {
						action := c.IActions[i]
						return Button().
							Flat(false).
							Label(action.Name).
							Icon(action.Icon).
							To(action.To).
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
	for name, value := range c.IStyles {
		element = element.Style(name, value)
	}
	return element
}
