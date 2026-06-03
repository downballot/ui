package myui

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// InputWrapper is a wrapper around an input element.
//
// Use this component to create custom input elements with a label and a body.
func InputWrapper() *MyUIInputWrapper {
	return &MyUIInputWrapper{}
}

type MyUIInputWrapper struct {
	app.Compo
	UseEvents
	IClass []string
	ILabel string
	IBody  []app.UI
}

var _ app.Composer = (*MyUIInputWrapper)(nil)

func (c *MyUIInputWrapper) Class(class ...string) *MyUIInputWrapper {
	c.IClass = append(c.IClass, class...)
	return c
}

func (c *MyUIInputWrapper) Label(label string) *MyUIInputWrapper {
	c.ILabel = label
	return c
}

func (c *MyUIInputWrapper) Body(input ...app.UI) *MyUIInputWrapper {
	c.IBody = append(c.IBody, input...)
	return c
}

func (c *MyUIInputWrapper) Render() app.UI {
	var body []app.UI
	if c.ILabel != "" {
		body = append(body, app.Span().
			Class("myui-input-wrapper__label").
			Text(c.ILabel))
	}
	body = append(body, c.IBody...)

	return app.Span().
		Class("myui-input-wrapper").
		Class(c.IClass...).
		Body(
			body...,
		)
}
