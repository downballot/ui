package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonFieldIDEditPage struct {
	app.Compo
	component.EmbeddedPage

	organizationID string
	personFieldID  string

	loaded bool

	Name               string
	Type               string
	AllowEmpty         bool
	AllowedRegex       string
	AllowedValues      []string
	ComputedExpression string

	Error string
}

func (c *OrganizationIDPersonFieldIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("person_field_id", &c.personFieldID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: OnNav", "PersonFieldID", c.personFieldID)

	c.Reload(ctx)
}

func (c *OrganizationIDPersonFieldIDEditPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: Reload", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: Reload", "PersonFieldID", c.personFieldID)

	if c.organizationID == "" {
		return
	}

	if c.personFieldID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetPersonFieldResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person-field/"+c.personFieldID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person field", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.DebugContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: Reload: Setting person field", "PersonField", output.PersonField)
				c.Name = output.PersonField.Name
				c.Type = string(output.PersonField.Type)
				c.AllowEmpty = output.PersonField.AllowEmpty
				c.AllowedRegex = output.PersonField.AllowedRegex
				c.AllowedValues = output.PersonField.AllowedValues
				c.ComputedExpression = output.PersonField.ComputedExpression
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.DebugContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: Reload: Dispatching update")
			ctx.Update()
		})
	})
}

func (c *OrganizationIDPersonFieldIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonFieldIDEditPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

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
				blazar.Input[string]().
					Label("Computed Expression").
					Type("text").
					Bind(&c.ComputedExpression),
			).
			CancelLabel("Reset").
			CancelFunction(c.Reload).
			SubmitLabel("Save").
			SubmitFunction(func(ctx app.Context) {
				ctx.PreventUpdate()

				ctx.Async(func() {
					input := downballotapi.PatchPersonFieldRequest{
						Name:               &c.Name,
						Type:               (*downballotapi.PersonFieldDefinitionType)(&c.Type),
						AllowEmpty:         &c.AllowEmpty,
						AllowedRegex:       &c.AllowedRegex,
						AllowedValues:      c.AllowedValues,
						ComputedExpression: &c.ComputedExpression,
					}
					var output downballotapi.PatchPersonFieldResponse
					err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/person-field/"+c.personFieldID, input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not create person field", "err", err)
						return
					}
					slog.InfoContext(ctx.Context, "OrganizationIDPersonFieldIDEditPage: Create button clicked: Navigating to person field page", "person_field_id", output.PersonField.ID)
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
