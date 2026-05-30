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

type OrganizationIDPersonFieldPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	PersonFields   []*downballotapi.PersonField
}

func (c *OrganizationIDPersonFieldPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnNav", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListPersonFieldsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person-field", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting person fields", "person fields", output.PersonFields)
			c.PersonFields = output.PersonFields
		})
	})
}

func (c *OrganizationIDPersonFieldPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldPage: Render")
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldPage: Render", "OrganizationID", c.OrganizationID)

	return myui.Page().Body(
		myui.Table[*downballotapi.PersonField]().
			Rows(c.PersonFields).
			Columns([]myui.TableColumn[*downballotapi.PersonField]{
				{
					Name: "ID",
					Value: func(row *downballotapi.PersonField) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.PersonField) any {
						return row.Name
					},
					To: func(row *downballotapi.PersonField) string {
						return fmt.Sprintf("/organization/%s/person-field/%s", c.OrganizationID, row.ID)
					},
				},
				{
					Name: "Type",
					Value: func(row *downballotapi.PersonField) any {
						return row.Type
					},
				},
				{
					Name: "Allow Empty",
					Value: func(row *downballotapi.PersonField) any {
						return row.AllowEmpty
					},
				},
				{
					Name: "Allowed Regex",
					Value: func(row *downballotapi.PersonField) any {
						return row.AllowedRegex
					},
				},
				{
					Name: "Allowed Values",
					Value: func(row *downballotapi.PersonField) any {
						return row.AllowedValues
					},
				},
			}).
			Action(myui.TableAction{
				Name: "New Person Field",
				To: func() string {
					slog.InfoContext(context.TODO(), "TableAction: New Person Field", "OrganizationID", c.OrganizationID)
					return fmt.Sprintf("/organization/%s/person-field/new", c.OrganizationID)
				},
			}),
	)
}
