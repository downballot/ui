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

type OrganizationIDPersonFieldIDPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	PersonFieldID  string `route:"person_field_id"`
	PersonField    *downballotapi.PersonField
}

func (c *OrganizationIDPersonFieldIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnUpdate", "PersonFieldID", c.PersonFieldID)

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
		var output downballotapi.GetPersonFieldResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person-field/"+c.PersonFieldID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person field", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting person field", "person field", output.PersonField)
			c.PersonField = output.PersonField
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Person field should be set", "person field", c.PersonField)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDPersonFieldIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldIDPage: Render")

	return app.Div().Body(
		app.If(c.Organization == nil || c.PersonField == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Div().Text(fmt.Sprintf("%+v", *c.Organization)),
				app.Div().Text(fmt.Sprintf("%+v", *c.PersonField)),
				app.Hr(),
				myui.Table[*downballotapi.PersonField]().
					Rows([]*downballotapi.PersonField{c.PersonField}).
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
