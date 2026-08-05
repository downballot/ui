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
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	personFields   []*downballotapi.PersonField
	permissionSet  permissionset.PermissionSet
}

func (c *OrganizationIDPersonFieldPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	ctx.Async(func() {
		var output downballotapi.ListPersonFieldsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person-field", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting person fields", "person fields", output.PersonFields)
			c.personFields = output.PersonFields
		})
	})
}

func (c *OrganizationIDPersonFieldPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldPage: Render")

	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldPage: Render", "OrganizationID", c.organizationID)

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		blazar.Table[*downballotapi.PersonField]().
			Rows(c.personFields).
			Columns([]blazar.TableColumn[*downballotapi.PersonField]{
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
						return fmt.Sprintf("/organization/%s/person-field/%s", c.organizationID, row.ID)
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
				{
					Name: "Computed Expression",
					Value: func(row *downballotapi.PersonField) any {
						return row.ComputedExpression
					},
				},
			}).
			Action(blazar.TableAction{
				Name:     "New Person Field",
				Icon:     component.IconAdd,
				To:       fmt.Sprintf("/organization/%s/person-field/new", c.organizationID),
				Disabled: !c.permissionSet.Match(iam.IAMPersonFieldDefinitionCreate),
			}).
			RowAction(blazar.RowAction[*downballotapi.PersonField]{
				Name: "Edit",
				Icon: component.IconEdit,
				To: func(row *downballotapi.PersonField) string {
					return fmt.Sprintf("/organization/%s/person-field/%s/edit", c.organizationID, row.ID)
				},
				Disabled: !c.permissionSet.Match(iam.IAMPersonFieldDefinitionUpdate),
			}),
	)
}
