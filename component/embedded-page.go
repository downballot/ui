package component

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-app-blazar/blazar/blazar"
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
	To       string
	Function func(ctx app.Context)
	Disabled bool
}

func (c *EmbeddedPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "EmbeddedPage: OnMount")

	slog.InfoContext(ctx.Context, "EmbeddedPage: Setting up embedded page")
	c.Setup(ctx)
}

func (c *EmbeddedPage) Action(action ...PageAction) *EmbeddedPage {
	c.IActions = action
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

			if c.IError != "" {
				ctx.Async(func() {
					var wg sync.WaitGroup
					wg.Go(func() {
						slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState: Async: Sleep", "api-error", c.IError)
						time.Sleep(10 * time.Second)
						slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState: Async: Sleep done", "api-error", c.IError)
					})
					wg.Wait()

					slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState: Async: Done waiting.")

					slog.InfoContext(ctx.Context, "EmbeddedPage: ObserveState: Defer: Clear api-error")
					ctx.SetState("api-error", "").ExpiresIn(1 * time.Millisecond)
				})
			}
		})
}

func (c *EmbeddedPage) Wrap(content ...app.UI) app.UI {
	var allElements []app.UI
	if c.IError != "" {
		allElements = append(allElements, blazar.StatusBar().
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
			form := blazar.Form()
			var formActions []blazar.FormAction
			for _, action := range actions {
				formActions = append(formActions, blazar.FormAction{
					Name:     action.Name,
					Icon:     action.Icon,
					To:       action.To,
					Function: action.Function,
				})
			}
			form.Action(formActions...)
			allElements = append(allElements, form)
		}
	}
	for _, element := range content {
		allElements = append(allElements, element)
	}

	return blazar.Page().
		Body(
			allElements...,
		)
}
