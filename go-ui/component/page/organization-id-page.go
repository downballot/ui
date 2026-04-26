package page

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationIDPage struct {
	app.Compo

	organizationID string
	organization   *downballotapi.Organization
}

func (c *OrganizationIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav")

	var variables map[string]string
	ctx.GetState("route", &variables)
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav", "variables", variables)

	c.organizationID = variables["organization_id"]
	slog.InfoContext(ctx.Context, "Organization ID", "id", c.organizationID)
	if c.organizationID == "" {
		return
	}

	var output downballotapi.GetOrganizationResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
		return
	}
	c.organization = &output.Organization
}

func (c *OrganizationIDPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnMount")

	var variables map[string]string
	ctx.GetState("route", &variables)
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnMount", "variables", variables)
}

func (c *OrganizationIDPage) Render() app.UI {
	return app.Div().Body(
		app.If(c.organization == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Text(fmt.Sprintf("%+v", *c.organization))
		}),
	)
}
