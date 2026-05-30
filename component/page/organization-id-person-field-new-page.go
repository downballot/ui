package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldNewPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`

	Name          string
	Type          string
	AllowEmpty    bool
	AllowedRegex  string
	AllowedValues []string

	Error string
}

func (c *OrganizationIDPersonFieldNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: OnNav", "OrganizationID", c.OrganizationID)

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
}

func (c *OrganizationIDPersonFieldNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldNewPage: Render")

	return myui.Page().Body(
		myui.Input[string]().
			Label("Name").
			Type("text").
			Bind(&c.Name),
		myui.Input[string]().
			Label("Type").
			Type("text").
			Bind(&c.Type),
		myui.Input[bool]().
			Label("Allow Empty").
			Type("checkbox").
			Bind(&c.AllowEmpty),
		myui.Input[string]().
			Label("Allowed Regex").
			Type("text").
			Bind(&c.AllowedRegex),
		myui.Input[string]().
			Label("Allowed Values").
			Type("text").
			Value(strings.Join(c.AllowedValues, ",")).
			On("change", c.ValueTo(&c.AllowedValues)),
		app.If(c.Error != "", func() app.UI {
			return app.Div().Body(
				app.Span().Text(c.Error),
			)
		}),
		myui.Button().
			Label("Create").
			On("click", func(ctx app.Context, e app.Event) {
				var input downballotapi.CreatePersonFieldRequest
				input.Name = c.Name
				input.Type = downballotapi.PersonFieldDefinitionType(c.Type)
				input.AllowEmpty = c.AllowEmpty
				input.AllowedRegex = c.AllowedRegex
				input.AllowedValues = c.AllowedValues
				var output downballotapi.CreatePersonFieldResponse
				err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.OrganizationID+"/person-field", input, &output)
				if err != nil {
					slog.ErrorContext(ctx.Context, "Could not create person field", "err", err)
					return
				}
				slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: Create button clicked: Navigating to person field page", "person_field_id", output.PersonField.ID)
				ctx.Navigate(fmt.Sprintf("/organization/%s/person-field/%s", c.OrganizationID, output.PersonField.ID))
			}),
	)
}
