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
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldNewPage struct {
	app.Compo
	myui.EmbeddedPage

	organizationID string

	Name          string
	Type          string
	AllowEmpty    bool
	AllowedRegex  string
	AllowedValues []string

	Error string
}

func (c *OrganizationIDPersonFieldNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}
	})
}

func (c *OrganizationIDPersonFieldNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldNewPage: Render")

	return c.EmbeddedPage.Wrap(
		myui.Form().
			Body(
				myui.Input[string]().
					Label("Name").
					Type("text").
					Bind(&c.Name),
				myui.Select().
					Label("Type").
					AllowedValue(
						myui.SelectOption{Label: "", Value: "", Disabled: true},
						myui.SelectOption{Label: "Boolean", Value: string(downballotapi.PersonFieldDefinitionTypeBoolean)},
						myui.SelectOption{Label: "Date", Value: string(downballotapi.PersonFieldDefinitionTypeDate)},
						myui.SelectOption{Label: "Enum", Value: string(downballotapi.PersonFieldDefinitionTypeEnum)},
						myui.SelectOption{Label: "Integer", Value: string(downballotapi.PersonFieldDefinitionTypeInteger)},
						myui.SelectOption{Label: "Set", Value: string(downballotapi.PersonFieldDefinitionTypeSet)},
						myui.SelectOption{Label: "String", Value: string(downballotapi.PersonFieldDefinitionTypeString)},
					).
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
			).
			SubmitLabel("Create").
			SubmitFunction(func(ctx app.Context) {
				ctx.PreventUpdate()

				ctx.Async(func() {
					input := downballotapi.CreatePersonFieldRequest{
						Name:          c.Name,
						Type:          downballotapi.PersonFieldDefinitionType(c.Type),
						AllowEmpty:    c.AllowEmpty,
						AllowedRegex:  c.AllowedRegex,
						AllowedValues: c.AllowedValues,
					}
					var output downballotapi.CreatePersonFieldResponse
					err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/person-field", input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not create person field", "err", err)
						return
					}
					slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldNewPage: Create button clicked: Navigating to person field page", "person_field_id", output.PersonField.ID)
					ctx.Navigate(fmt.Sprintf("/organization/%s/person-field/%s", c.organizationID, output.PersonField.ID))
				})
			}),
		app.If(c.Error != "", func() app.UI {
			return app.Div().Body(
				app.Span().Text(c.Error),
			)
		}),
	)
}
