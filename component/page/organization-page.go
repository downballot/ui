package page

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationPage struct {
	app.Compo

	Organizations []*downballotapi.Organization
}

func (c *OrganizationPage) OnUpdate(ctx app.Context) {
	ctx.Async(func() {
		var output downballotapi.ListOrganizationsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.Organizations = output.Organizations
		})
	})
}

func (c *OrganizationPage) Render() app.UI {
	return myui.Page().
		Body(
			app.Div().Text(
				"These are the organizations that you are a part of.",
			),
			myui.NewTable[*downballotapi.Organization]().
				Rows(c.Organizations).
				Columns([]myui.TableColumn[*downballotapi.Organization]{
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
				}).Render(),
		)
}
