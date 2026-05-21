package page

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
}

func (c *OrganizationIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting organization", "organization", output.Organization)
			c.Organization = &output.Organization
		})
	})
}

func (c *OrganizationIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPage: Render")

	if !c.Loaded {
		return nil
	}

	if c.Organization == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
			app.Div().Text("ID: "+c.Organization.ID),
			app.Div().Text("Name: "+c.Organization.Name),
		)
}
