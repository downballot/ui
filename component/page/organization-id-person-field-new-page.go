package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldNewPage struct {
	app.Compo
	component.EmbeddedPage

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
}

func (c *OrganizationIDPersonFieldNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldNewPage: Render")

	return c.EmbeddedPage.Wrap(
		blazar.Form().
			Body(
				blazar.Input[string]().
					Label("Name").
					Type("text").
					Bind(&c.Name),
				blazar.Select().
					Label("Type").
					AllowedValue(
						blazar.SelectOption{Label: "", Value: "", Disabled: true},
						blazar.SelectOption{Label: "Boolean", Value: string(downballotapi.PersonFieldDefinitionTypeBoolean)},
						blazar.SelectOption{Label: "Date", Value: string(downballotapi.PersonFieldDefinitionTypeDate)},
						blazar.SelectOption{Label: "Enum", Value: string(downballotapi.PersonFieldDefinitionTypeEnum)},
						blazar.SelectOption{Label: "Integer", Value: string(downballotapi.PersonFieldDefinitionTypeInteger)},
						blazar.SelectOption{Label: "Set", Value: string(downballotapi.PersonFieldDefinitionTypeSet)},
						blazar.SelectOption{Label: "String", Value: string(downballotapi.PersonFieldDefinitionTypeString)},
					).
					Bind(&c.Type),
				blazar.Input[bool]().
					Label("Allow Empty").
					Type("checkbox").
					Bind(&c.AllowEmpty),
				blazar.Input[string]().
					Label("Allowed Regex").
					Type("text").
					Bind(&c.AllowedRegex),
				blazar.Input[string]().
					Label("Allowed Values").
					Type("text").
					Value(strings.Join(c.AllowedValues, ",")).
					On("change", func(ctx app.Context, e app.Event) {
						var stringValue string
						c.ValueTo(&stringValue)(ctx, e)
						c.AllowedValues = strings.Split(stringValue, ",")
						for i := range c.AllowedValues {
							c.AllowedValues[i] = strings.TrimSpace(c.AllowedValues[i])
						}

						ctx.Update()
					}),
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
