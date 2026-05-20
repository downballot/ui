package myui

import (
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
	LabelValue          string
	AllowedValuesValue  []SelectOption
	SelectedValuesValue []string
}

type SelectOption struct {
	Label string
	Value string
}

var _ app.Composer = (*MyUIMultiselect)(nil)

func (c *MyUIMultiselect) Label(label string) *MyUIMultiselect {
	c.LabelValue = label
	return c
}

func (c *MyUIMultiselect) AllowedValue(allowedValue ...SelectOption) *MyUIMultiselect {
	c.AllowedValuesValue = append(c.AllowedValuesValue, allowedValue...)
	return c
}

func (c *MyUIMultiselect) SelectedValue(selectedValue ...string) *MyUIMultiselect {
	c.SelectedValuesValue = append(c.SelectedValuesValue, selectedValue...)
	return c
}

func (c *MyUIMultiselect) On(event string, function func(ctx app.Context, e app.Event)) *MyUIMultiselect {
	c.UseEvents.On(event, function)
	return c
}

func (c *MyUIMultiselect) Render() app.UI {
	return app.Span().
		Class("myui-multiselect").
		Body(
			app.If(c.LabelValue != "", func() app.UI {
				return app.Span().
					Class("myui-input-label").
					Text(c.LabelValue)
			}),
			c.UseEvents.Wrap(
				app.Select().
					Class("myui-multiselect-select").
					Multiple(true).
					Body(
						app.Range(c.AllowedValuesValue).Slice(func(i int) app.UI {
							allowedValue := c.AllowedValuesValue[i]
							return app.Option().
								Value(allowedValue.Value).
								Text(allowedValue.Label).
								Selected(slices.Contains(c.SelectedValuesValue, allowedValue.Value))
						}),
					),
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
