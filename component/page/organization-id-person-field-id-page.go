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
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldIDPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	personFieldID  string
	personField    *downballotapi.PersonField
}

func (c *OrganizationIDPersonFieldIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("person_field_id", &c.personFieldID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDPage: OnNav", "PersonFieldID", c.personFieldID)

	if c.organizationID == "" {
		return
	}

	if c.personFieldID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
			var output downballotapi.GetPersonFieldResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person-field/"+c.personFieldID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person field", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting person field", "person field", output.PersonField)
				c.personField = output.PersonField
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDPersonFieldIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldIDPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.personField == nil {
		return c.EmbeddedPage.Wrap(
			myui.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return c.EmbeddedPage.Wrap(
		myui.Table[*downballotapi.PersonField]().
			Rows([]*downballotapi.PersonField{c.personField}).
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
			}),
	)
}
