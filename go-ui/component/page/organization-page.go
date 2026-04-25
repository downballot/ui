package page

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/downballot/ui/go-ui/component/customlayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationPage struct {
	app.Compo

	organizations []*downballotapi.Organization
}

func (c *OrganizationPage) OnMount(ctx app.Context) {
	var output downballotapi.ListOrganizationsResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization", nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
		return
	}
	c.organizations = output.Organizations
}

func (c *OrganizationPage) Render() app.UI {
	return &customlayout.DownballotLayout{
		Content: app.Range(c.organizations).Slice(func(i int) app.UI {
			organization := *c.organizations[i]

			return app.Div().Body(
				app.A().
					Href("/organization/" + organization.ID).
					Text(fmt.Sprintf("%+v", organization)),
			)
		}),
	}
}
