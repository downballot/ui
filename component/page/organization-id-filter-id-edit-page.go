package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDFilterIDEditPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	filterID       string

	Name        string
	Description string
	Filter      string
}

func (c *OrganizationIDFilterIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("filter_id", &c.filterID)

	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: OnNav", "FilterID", c.filterID)

	c.Reload(ctx)
}

func (c *OrganizationIDFilterIDEditPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload", "FilterID", c.filterID)

	if c.organizationID == "" {
		return
	}

	if c.filterID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetFilterResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/filter/"+c.filterID, nil, &output)
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
			c.loaded = true

			slog.DebugContext(ctx.Context, "OrganizationIDFilterIDEditPage: Reload: Dispatching update")
			ctx.Update()
		})
	})
}

func (c *OrganizationIDFilterIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterIDEditPage: Render", "Name", c.Name, "Description", c.Description, "Filter", c.Filter)

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
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
								err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/filter/"+c.filterID, input, &output)
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
