package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDFilterNewPage struct {
	app.Compo
	myui.EmbeddedPage

	organizationID string

	Name        string
	Description string
	Filter      string
}

func (c *OrganizationIDFilterNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnNav", "OrganizationID", c.organizationID)
}

func (c *OrganizationIDFilterNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterNewPage: Render")

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
						Label("Create").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.Async(func() {
								var input downballotapi.CreateFilterRequest
								input.Name = c.Name
								input.Description = c.Description
								input.Filter = c.Filter
								var output downballotapi.CreateFilterResponse
								err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/filter", input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not create filter", "err", err)
									return
								}
								slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: Create button clicked: Navigating to filter page", "filter_id", output.ID)
								ctx.Navigate(fmt.Sprintf("/organization/%s/filter/%s", c.organizationID, output.ID))
							})
						}),
				),
			),
	)
}
