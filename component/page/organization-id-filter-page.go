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

type OrganizationIDFilterPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Filters        []*downballotapi.Filter
}

func (c *OrganizationIDFilterPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDFilterPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListFiltersResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/filter", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get filters", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting filters", "filters", output.Filters)
			c.Filters = output.Filters
		})
	})
}

func (c *OrganizationIDFilterPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterPage: Render")

	return myui.Page().Body(
		myui.Table[*downballotapi.Filter]().
			Rows(c.Filters).
			Columns([]myui.TableColumn[*downballotapi.Filter]{
				{
					Name: "ID",
					Value: func(row *downballotapi.Filter) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.Filter) any {
						return row.Name
					},
					To: func(row *downballotapi.Filter) string {
						return fmt.Sprintf("/organization/%s/filter/%s", c.OrganizationID, row.ID)
					},
				},
				{
					Name: "Description",
					Value: func(row *downballotapi.Filter) any {
						return row.Description
					},
				},
				{
					Name: "Filter",
					Value: func(row *downballotapi.Filter) any {
						return row.Filter
					},
				},
			}).
			Action(myui.TableAction{
				Name: "New filter",
				Icon: "plus",
				To: func() string {
					return fmt.Sprintf("/organization/%s/filter/new", c.OrganizationID)
				},
			}).
			RowAction(myui.RowAction[*downballotapi.Filter]{
				Name: "Edit",
				Icon: "edit",
				To: func(row *downballotapi.Filter) string {
					return fmt.Sprintf("/organization/%s/filter/%s/edit", c.OrganizationID, row.ID)
				},
			}).
			Render(),
	)
}
