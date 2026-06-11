package page

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
}

func (c *OrganizationIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting organization", "organization", output.Organization)
			c.organization = &output.Organization
		})
	})
}

func (c *OrganizationIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.organization == nil {
		return c.EmbeddedPage.Wrap(
			myui.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().Text("ID: "+c.organization.ID),
		app.Div().Text("Name: "+c.organization.Name),
	)
}
