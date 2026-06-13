package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDFilterPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	filters        []*downballotapi.Filter
	permissionSet  permissionset.PermissionSet
}

func (c *OrganizationIDFilterPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDFilterPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDFilterPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	ctx.Async(func() {
		var output downballotapi.ListFiltersResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/filter", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get filters", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting filters", "filters", output.Filters)
			c.filters = output.Filters
		})
	})
}

func (c *OrganizationIDFilterPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDFilterPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		myui.Table[*downballotapi.Filter]().
			Rows(c.filters).
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
						return fmt.Sprintf("/organization/%s/filter/%s", c.organizationID, row.ID)
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
				Name:     "New filter",
				Icon:     "plus",
				To:       fmt.Sprintf("/organization/%s/filter/new", c.organizationID),
				Disabled: !c.permissionSet.Match(iam.IAMFilterCreate),
			}).
			RowAction(myui.RowAction[*downballotapi.Filter]{
				Name: "Edit",
				Icon: "edit",
				To: func(row *downballotapi.Filter) string {
					return fmt.Sprintf("/organization/%s/filter/%s/edit", c.organizationID, row.ID)
				},
				Disabled: !c.permissionSet.Match(iam.IAMFilterUpdate),
			}),
	)
}
