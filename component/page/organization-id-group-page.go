package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupPage struct {
	app.Compo

	loaded bool

	organizationID string
	groups         []*downballotapi.Group
}

func (c *OrganizationIDGroupPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListGroupsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group?parent_id=null", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting groups", "groups", output.Groups)
			c.groups = output.Groups
		})
	})
}

func (c *OrganizationIDGroupPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupPage: Render")

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	return myui.Page().Body(
		myui.Table[*downballotapi.Group]().
			Rows(c.groups).
			Columns([]myui.TableColumn[*downballotapi.Group]{
				{
					Name: "ID",
					Value: func(row *downballotapi.Group) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.Group) any {
						return row.Name
					},
					To: func(row *downballotapi.Group) string {
						return fmt.Sprintf("/organization/%s/group/%s", c.organizationID, row.ID)
					},
				},
				{
					Name: "Filter",
					Value: func(row *downballotapi.Group) any {
						return row.Filter
					},
				},
			}),
	)
}
