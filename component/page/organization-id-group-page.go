package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationIDGroupPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Groups         []*downballotapi.Group
}

func (c *OrganizationIDGroupPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupPage: OnUpdate", "OrganizationID", c.OrganizationID)

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
	})
	ctx.Async(func() {
		var output downballotapi.ListGroupsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting groups", "groups", output.Groups)
			c.Groups = output.Groups
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Groups should be set", "groups", c.Groups)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDGroupPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupPage: Render")

	return app.Div().Body(
		myui.Table[*downballotapi.Group]().
			Rows(c.Groups).
			Columns([]myui.TableColumn[*downballotapi.Group]{
				{
					Name: "ID",
					Value: func(row *downballotapi.Group) any {
						return row.ID
					},
				},
				{
					Name: "Parent ID",
					Value: func(row *downballotapi.Group) any {
						return row.ParentID
					},
					To: func(row *downballotapi.Group) string {
						if row.ParentID == "" {
							return ""
						}
						return fmt.Sprintf("/organization/%s/group/%s", c.OrganizationID, row.ParentID)
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.Group) any {
						return row.Name
					},
					To: func(row *downballotapi.Group) string {
						return fmt.Sprintf("/organization/%s/group/%s", c.OrganizationID, row.ID)
					},
				},
				{
					Name: "Filter",
					Value: func(row *downballotapi.Group) any {
						return row.Filter
					},
				},
			}).Render(),
	)
}
