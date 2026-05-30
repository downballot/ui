package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	PersonFieldID  string `route:"person_field_id"`
	PersonField    *downballotapi.PersonField
}

func (c *OrganizationIDPersonFieldIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav", "PersonFieldID", c.PersonFieldID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
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
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true
		})
	})
}

func (c *OrganizationIDPersonFieldIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldIDPage: Render")

	if !c.Loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	if c.PersonField == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
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
				}),
		)
}
