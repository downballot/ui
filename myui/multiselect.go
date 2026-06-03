package myui

import (
	"context"
	"log/slog"
	"slices"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Multiselect() *MyUIMultiselect {
	return &MyUIMultiselect{}
}

type MyUIMultiselect struct {
	app.Compo
	UseEvents
	IName              string
	ILabel             string
	IAllowedValues     []SelectOption
	ISelectedValues    []string
	BindSelectedValues *[]string
}

var _ app.Composer = (*MyUIMultiselect)(nil)
var _ app.Updater = (*MyUIMultiselect)(nil)

func (c *MyUIMultiselect) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUIMultiselect: OnUpdate")
}

func (c *MyUIMultiselect) Name(name string) *MyUIMultiselect {
	c.IName = name
	return c
}

func (c *MyUIMultiselect) Label(label string) *MyUIMultiselect {
	c.ILabel = label
	return c
}

func (c *MyUIMultiselect) AllowedValue(allowedValue ...SelectOption) *MyUIMultiselect {
	c.IAllowedValues = append(c.IAllowedValues, allowedValue...)
	return c
}

func (c *MyUIMultiselect) SelectedValue(selectedValue ...string) *MyUIMultiselect {
	c.ISelectedValues = append(c.ISelectedValues, selectedValue...)
	if c.BindSelectedValues != nil {
		*c.BindSelectedValues = append(*c.BindSelectedValues, selectedValue...)
	}
	return c
}

func (c *MyUIMultiselect) Bind(bindSelectedValues *[]string) *MyUIMultiselect {
	c.BindSelectedValues = bindSelectedValues
	if c.BindSelectedValues != nil {
		c.ISelectedValues = make([]string, len(*c.BindSelectedValues))
		copy(c.ISelectedValues, *c.BindSelectedValues)
	}
	return c
}

func (c *MyUIMultiselect) On(event string, function func(ctx app.Context, e app.Event)) *MyUIMultiselect {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIMultiselect) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUIMultiselect: Render", "label", c.ILabel, "allowedValues", c.IAllowedValues, "selectedValues", c.ISelectedValues, "bindSelectedValues", c.BindSelectedValues)
	return InputWrapper().
		Class("myui-multiselect").
		Label(c.ILabel).
		Body(
			c.UseEvents.Wrap(
				app.Select().
					Name(c.IName).
					Class("myui-multiselect__select").
					Multiple(true).
					Body(
						app.Range(c.IAllowedValues).Slice(func(i int) app.UI {
							allowedValue := c.IAllowedValues[i]
							return app.Option().
								Value(allowedValue.Value).
								Text(allowedValue.Label).
								Disabled(allowedValue.Disabled).
								Selected(slices.Contains(c.ISelectedValues, allowedValue.Value))
						}),
					),
				WithOn("change", func(ctx app.Context, e app.Event) {
					SelectedValuesTo(&c.ISelectedValues)(ctx, e)
					if c.BindSelectedValues != nil {
						*c.BindSelectedValues = make([]string, len(c.ISelectedValues))
						copy(*c.BindSelectedValues, c.ISelectedValues)
					}
					ctx.PreventUpdate()
				}),
			),
		)
}

// SelectedValuesTo is an event handler that updates the variable based on the selected options in a multiselect.
func SelectedValuesTo(selectedValues *[]string) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		targetElement := e.Value.Get("target")
		if targetElement.IsNull() {
			return
		}
		selectedOptions := targetElement.Get("selectedOptions")
		if selectedOptions.IsNull() {
			return
		}

		*selectedValues = []string{}
		selectedOptionsLength := selectedOptions.Length()
		for i := range selectedOptionsLength {
			selectedOption := selectedOptions.Index(i)
			if selectedOption.IsNull() {
				continue
			}
			selectedOptionValue := selectedOption.Get("value").String()
			*selectedValues = append(*selectedValues, selectedOptionValue)
		}
		slog.InfoContext(ctx.Context, "SelectedValuesTo: Selected values", "selectedValues", *selectedValues)
	}
}
