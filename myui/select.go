package myui

import (
	"context"
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Select() *MyUISelect {
	return &MyUISelect{}
}

type MyUISelect struct {
	app.Compo
	UseEvents
	IName             string
	ILabel            string
	IAllowedValues    []SelectOption
	ISelectedValue    string
	BindSelectedValue *string
}

var _ app.Composer = (*MyUISelect)(nil)
var _ app.Updater = (*MyUISelect)(nil)

func (c *MyUISelect) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUISelect: OnUpdate")
}

func (c *MyUISelect) Name(name string) *MyUISelect {
	c.IName = name
	return c
}

func (c *MyUISelect) Label(label string) *MyUISelect {
	c.ILabel = label
	return c
}

func (c *MyUISelect) AllowedValue(allowedValue ...SelectOption) *MyUISelect {
	c.IAllowedValues = allowedValue
	return c
}

func (c *MyUISelect) SelectedValue(selectedValue string) *MyUISelect {
	c.ISelectedValue = selectedValue
	if c.BindSelectedValue != nil {
		*c.BindSelectedValue = selectedValue
	}
	return c
}

func (c *MyUISelect) Bind(bindSelectedValue *string) *MyUISelect {
	c.BindSelectedValue = bindSelectedValue
	if c.BindSelectedValue != nil {
		c.ISelectedValue = *c.BindSelectedValue
	}
	return c
}

func (c *MyUISelect) On(event string, function func(ctx app.Context, e app.Event)) *MyUISelect {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUISelect) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUISelect: Render", "label", c.ILabel, "allowedValues", c.IAllowedValues, "selectedValue", c.ISelectedValue, "bindSelectedValue", c.BindSelectedValue)
	return InputWrapper().
		Class("myui-select").
		Label(c.ILabel).
		Body(
			c.UseEvents.Wrap(
				app.Select().
					Name(c.IName).
					Class("myui-select__select").
					Multiple(false).
					Body(
						app.Range(c.IAllowedValues).Slice(func(i int) app.UI {
							allowedValue := c.IAllowedValues[i]
							return app.Option().
								Value(allowedValue.Value).
								Text(allowedValue.Label).
								Disabled(allowedValue.Disabled).
								Selected(c.ISelectedValue == allowedValue.Value)
						}),
					),
				WithOn("change", func(ctx app.Context, e app.Event) {
					c.ValueTo(&c.ISelectedValue)(ctx, e)
					if c.BindSelectedValue != nil {
						*c.BindSelectedValue = c.ISelectedValue
					}
					ctx.PreventUpdate()
				}),
			),
		)
}
