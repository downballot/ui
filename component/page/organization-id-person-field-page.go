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

type OrganizationIDPersonFieldPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	PersonFields   []*downballotapi.PersonField
}

func (c *OrganizationIDPersonFieldPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldPage: OnUpdate", "OrganizationID", c.OrganizationID)

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

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting organization", "organization", output.Organization)
			c.Organization = &output.Organization
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Organization should be set", "organization", c.Organization)

			//ctx.Update()
		})
	})
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
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Person fields should be set", "person fields", c.PersonFields)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDPersonFieldPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldPage: Render")

	return app.Div().Body(
		app.If(c.Organization == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Div().Text(fmt.Sprintf("%+v", *c.Organization)),
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
					}).Render(),
			)
		}),
	)
}
