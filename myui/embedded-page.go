package myui

import (
	"log/slog"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type EmbeddedPage struct {
	IError string

	IActions []PageAction
}

var _ app.Mounter = (*EmbeddedPage)(nil)

type PageAction struct {
	Name     string
	Icon     string
	To       func() string
	Function func(ctx app.Context)
	Disabled bool
}

func (c *EmbeddedPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "EmbeddedPage: OnMount")

	slog.InfoContext(ctx.Context, "EmbeddedPage: Setting up embedded page")
	c.Setup(ctx)
}

func (c *EmbeddedPage) Action(action ...PageAction) *EmbeddedPage {
	c.IActions = append(c.IActions, action...)
	return c
}

func (c *EmbeddedPage) Setup(ctx app.Context) {
	slog.InfoContext(ctx.Context, "EmbeddedPage: Setup")

	var apiError string
	ctx.GetState("api-error", &apiError)
	c.IError = apiError
	slog.InfoContext(ctx.Context, "EmbeddedPage: Setup", "api-error", c.IError)

	ctx.ObserveState("api-error", &c.IError).OnChange(
		func() {
			slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState", "api-error", c.IError)
		})
}

func (c *EmbeddedPage) Wrap(content ...app.UI) app.UI {
	var allElements []app.UI
	if c.IError != "" {
		allElements = append(allElements, StatusBar().
			Text(c.IError).
			Bad(),
		)
	}
	{
		var actions []PageAction
		for _, action := range c.IActions {
			if !action.Disabled {
				actions = append(actions, action)
			}
		}
		if len(actions) > 0 {
			form := Form()
			for _, action := range actions {
				form.Action(FormAction{
					Name:     action.Name,
					Icon:     action.Icon,
					To:       action.To,
					Function: action.Function,
				})
			}
			allElements = append(allElements, form)
		}
	}
	for _, element := range content {
		allElements = append(allElements, element)
	}

	return Page().
		Body(
			allElements...,
		)
}
