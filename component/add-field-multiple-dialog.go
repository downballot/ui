package component

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AddFieldMultipleDialog struct {
	app.Compo

	OrganizationID string
	voterIDs       []string

	DialogID string
	error    string

	PersonFields []*downballotapi.PersonField

	SubmitNameValue string
	OnSubmit        func(ctx app.Context)

	SelectedFieldValue string
	ValueValue         string
}

func (c *AddFieldMultipleDialog) Open(ctx app.Context, voterIDs []string) {
	slog.InfoContext(ctx.Context, "AddFieldMultipleDialog: Open", "DialogID", c.DialogID)
	slog.InfoContext(ctx.Context, "AddFieldMultipleDialog: Open", "JSValue", c.JSValue(), "JSValue", app.Window().Get("JSON").Call("stringify", c.JSValue()))

	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("showModal")

	c.voterIDs = voterIDs

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

			ctx.Update()
		})
	})
}

func (c *AddFieldMultipleDialog) Close(ctx app.Context) {
	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("close")
}

func (c *AddFieldMultipleDialog) Render() app.UI {
	slog.InfoContext(context.TODO(), "AddFieldMultipleDialog: Render", "OrganizationID", c.OrganizationID, "voterIDs", c.voterIDs)

	submitName := c.SubmitNameValue
	if submitName == "" {
		submitName = "Submit"
	}

	var selectedPersonField *downballotapi.PersonField
	for _, personField := range c.PersonFields {
		if personField.Name == c.SelectedFieldValue {
			selectedPersonField = personField
			break
		}
	}

	var valueElements []app.UI
	if selectedPersonField != nil {
		switch selectedPersonField.Type {
		case downballotapi.PersonFieldDefinitionTypeDate:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("date").
					Placeholder("Value").
					Bind(&c.ValueValue),
				blazar.Button().
					Flat(true).
					Label("Yesterday").
					On("click", func(ctx app.Context, e app.Event) {
						c.ValueValue = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
						ctx.Update()
					}),
				blazar.Button().
					Flat(true).
					Label("Today").
					On("click", func(ctx app.Context, e app.Event) {
						c.ValueValue = time.Now().Format("2006-01-02")
						ctx.Update()
					}),
			)
		case downballotapi.PersonFieldDefinitionTypeEnum:
			valueElements = append(valueElements,
				blazar.Select().
					AllowedValue(
						func() []blazar.SelectOption {
							var allowedValues []blazar.SelectOption
							allowedValues = append(allowedValues, blazar.SelectOption{Label: "", Value: "", Disabled: true})
							for _, allowedValue := range selectedPersonField.AllowedValues {
								allowedValues = append(allowedValues, blazar.SelectOption{Label: allowedValue, Value: allowedValue})
							}
							return allowedValues
						}()...).
					Bind(&c.ValueValue),
			)
		case downballotapi.PersonFieldDefinitionTypeString:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.ValueValue),
			)
		case downballotapi.PersonFieldDefinitionTypeInteger:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("number").
					Placeholder("Value").
					Bind(&c.ValueValue),
			)
		case downballotapi.PersonFieldDefinitionTypeBoolean:
			valueElements = append(valueElements,
				blazar.Select().
					AllowedValue(
						blazar.SelectOption{Label: "", Value: "", Disabled: true},
						blazar.SelectOption{Label: "true", Value: "true"},
						blazar.SelectOption{Label: "false", Value: "false"},
					).
					Bind(&c.ValueValue),
			)
		default:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.ValueValue),
			)
		}
	}

	formBody := []app.UI{
		blazar.Select().
			Name("field").
			Label("Field").
			AllowedValue(
				func() []blazar.SelectOption {
					var allowedValues []blazar.SelectOption
					allowedValues = append(allowedValues, blazar.SelectOption{Label: "", Value: "", Disabled: true})
					for _, personField := range c.PersonFields {
						allowedValues = append(allowedValues, blazar.SelectOption{Label: personField.Name, Value: personField.Name})
					}
					return allowedValues
				}()...).
			Bind(&c.SelectedFieldValue).
			On("change", func(ctx app.Context, e app.Event) {
				slog.InfoContext(ctx.Context, "AddFieldMultipleDialog: OnChange", "SelectedFieldValue", c.SelectedFieldValue)

				c.ValueValue = ""

				// TODO: Can we pick the first option if it's a dropdown?

				slog.InfoContext(ctx.Context, "AddFieldMultipleDialog: OnChange", "ValueValue", c.ValueValue)
				ctx.Update()
			}),
	}
	formBody = append(formBody, valueElements...)

	return app.Dialog().
		ID(c.DialogID).
		Body(
			app.H2().Text("Add Field"),
			blazar.Form().
				Body(formBody...).
				CancelFunction(c.Close).
				SubmitLabel("Save").
				SubmitFunction(func(ctx app.Context) {
					slog.InfoContext(ctx.Context, "AddFieldMultipleDialog: SubmitFunction", "SelectedFieldValue", c.SelectedFieldValue, "ValueValue", c.ValueValue)

					input := downballotapi.PostPersonUpdateRequest{
						VoterIDs: c.voterIDs,
						Fields:   map[string]*string{},
					}
					input.Fields[c.SelectedFieldValue] = &c.ValueValue
					var output downballotapi.PostPersonUpdateResponse
					err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.OrganizationID+"/person/update", input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not update persons", "err", err)
						return
					}

					if c.OnSubmit != nil {
						c.OnSubmit(ctx)
					}

					c.error = ""
					c.Close(ctx)
				}),
			app.If(c.error != "", func() app.UI {
				return blazar.StatusBar().
					Text(c.error).
					Bad()
			}),
		)
}
