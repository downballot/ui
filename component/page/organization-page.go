package page

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizations []*downballotapi.Organization
}

func (c *OrganizationPage) OnNav(ctx app.Context) {
	ctx.Async(func() {
		var output downballotapi.ListOrganizationsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
			c.organizations = output.Organizations
		})
	})
}

func (c *OrganizationPage) Render() app.UI {
	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().Text(
			"These are the organizations that you are a part of.",
		),
		blazar.Table[*downballotapi.Organization]().
			Rows(c.organizations).
			Columns([]blazar.TableColumn[*downballotapi.Organization]{
				{
					Name: "ID",
					Value: func(row *downballotapi.Organization) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.Organization) any {
						return row.Name
					},
					To: func(row *downballotapi.Organization) string {
						return fmt.Sprintf("/organization/%s", row.ID)
					},
				},
			}),
	)
}
