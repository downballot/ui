package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDFilterIDEditPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	FilterID       string `route:"filter_id"`
	Name           string
	Description    string
	Filter         string
}

func (c *OrganizationIDFilterIDEditPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload", "FilterID", c.FilterID)

	if c.OrganizationID == "" {
		return
	}

	if c.FilterID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetFilterResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/filter/"+c.FilterID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get filter", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.DebugContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload: Setting filter", "filter", output.Filter)
				c.Name = output.Filter.Name
				c.Description = output.Filter.Description
				c.Filter = output.Filter.Filter
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			slog.DebugContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload: Dispatching update")
			ctx.Update()
		})
	})
}

func (c *OrganizationIDFilterIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: OnNav", "OrganizationID", c.OrganizationID)

	c.Reload(ctx)
}

func (c *OrganizationIDFilterIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterIDEditPage: Render", "Name", c.Name, "Description", c.Description, "Filter", c.Filter)

	return myui.Page().
		Body(
			app.Div().
				Style("display", "flex").
				Style("flex-direction", "column").
				Body(
					myui.Input[string]().
						Label("Name").
						Type("text").
						Bind(&c.Name),
					myui.Input[string]().
						Label("Description").
						Type("text").
						Bind(&c.Description),
					myui.Input[string]().
						Label("Filter").
						Type("text").
						Bind(&c.Filter),
					app.Div().Body(
						myui.Button().
							Label("Save").
							On("click", func(ctx app.Context, e app.Event) {
								ctx.Async(func() {
									var input downballotapi.PatchFilterRequest
									input.Name = &c.Name
									input.Description = &c.Description
									input.Filter = &c.Filter
									var output downballotapi.PatchFilterResponse
									err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.OrganizationID+"/filter/"+c.FilterID, input, &output)
									if err != nil {
										slog.ErrorContext(ctx.Context, "Could not patch filter", "err", err)
										return
									}

									c.Reload(ctx)
								})
							}),
					),
				),
		)
}
