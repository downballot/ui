package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDFilterNewPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Name           string
	Description    string
	Filter         string
}

func (c *OrganizationIDFilterNewPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}
}

func (c *OrganizationIDFilterNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDFilterNewPage: OnNav", "OrganizationID", c.OrganizationID)
}

func (c *OrganizationIDFilterNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterNewPage: Render")

	return myui.Page().Body(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input().
					Label("Name").
					Type("text").
					Value(c.Name).
					On("change", c.ValueTo(&c.Name)),
				myui.Input().
					Label("Description").
					Type("text").
					Value(c.Description).
					On("change", c.ValueTo(&c.Description)),
				myui.Input().
					Label("Filter").
					Type("text").
					Value(c.Filter).
					On("change", c.ValueTo(&c.Filter)),
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
								err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.OrganizationID+"/filter", input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not create filter", "err", err)
									return
								}
								ctx.Navigate(fmt.Sprintf("/organization/%s/filter/%s", c.OrganizationID, output.ID))
							})
						}),
				),
			),
	)
}
