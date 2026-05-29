package myui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Input[T any]() *MyUIInput[T] {
	return &MyUIInput[T]{}
}

type MyUIInput[T any] struct {
	app.Compo
	UseEvents
	IAutoFocus   bool
	IType        string
	IName        string
	IDisabled    bool
	ILabel       string
	IPlaceholder string
	BindValue    *T
}

var _ app.Composer = (*MyUIInput[any])(nil)

func (c *MyUIInput[T]) AutoFocus(autoFocus bool) *MyUIInput[T] {
	c.IAutoFocus = autoFocus
	return c
}

func (c *MyUIInput[T]) Name(name string) *MyUIInput[T] {
	c.IName = name
	return c
}

func (c *MyUIInput[T]) Placeholder(placeholder string) *MyUIInput[T] {
	c.IPlaceholder = placeholder
	return c
}

func (c *MyUIInput[T]) Disabled(disabled bool) *MyUIInput[T] {
	c.IDisabled = disabled
	return c
}

func (c *MyUIInput[T]) Type(inputType string) *MyUIInput[T] {
	c.IType = inputType
	return c
}

func (c *MyUIInput[T]) Label(label string) *MyUIInput[T] {
	c.ILabel = label
	return c
}

func (c *MyUIInput[T]) Value(value T) *MyUIInput[T] {
	if c.BindValue == nil {
		c.BindValue = new(T)
	}
	*c.BindValue = value
	return c
}

func (c *MyUIInput[T]) Bind(valuePointer *T) *MyUIInput[T] {
	c.BindValue = valuePointer
	return c
}

func (c *MyUIInput[T]) On(event string, function func(ctx app.Context, e app.Event)) *MyUIInput[T] {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIInput[T]) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUIInput: Render", "label", c.ILabel, "type", c.IType, "value", c.BindValue, "placeholder", c.IPlaceholder, "disabled", c.IDisabled)

	value := ""
	if c.BindValue != nil {
		value = fmt.Sprintf("%v", *c.BindValue)
	}

	return app.Span().
		Class("myui-input").
		Body(
			app.If(c.ILabel != "", func() app.UI {
				return app.Span().
					Class("myui-input__label").
					Text(c.ILabel)
			}),
			c.UseEvents.Wrap(
				app.Input().
					Class("myui-input__input").
					Disabled(c.IDisabled).
					AutoFocus(c.IAutoFocus).
					Name(c.IName).
					Type(c.IType).
					Value(value).
					Placeholder(c.IPlaceholder),
				//WithOn("change", c.ValueTo(c.BindValue)),
				WithOn("change", func(ctx app.Context, e app.Event) {
					slog.InfoContext(ctx.Context, "MyUIInput: Change", "value", value)
					c.ValueTo(c.BindValue)(ctx, e)
				}),
			),
		)
}
