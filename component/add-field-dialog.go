package component

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AddFieldDialog struct {
	app.Compo

	OrganizationID string
	VoterID        string
	Person         *downballotapi.Person

	DialogID string
	error    string

	PersonFields []*downballotapi.PersonField

	SubmitNameValue     string
	SubmitFunctionValue func(ctx app.Context) error

	SelectedFieldValue string
	ValueValue         string
}

func (c *AddFieldDialog) Open(ctx app.Context) {
	slog.InfoContext(ctx.Context, "AddFieldDialog: Open", "DialogID", c.DialogID)
	slog.InfoContext(ctx.Context, "AddFieldDialog: Open", "JSValue", c.JSValue(), "JSValue", app.Window().Get("JSON").Call("stringify", c.JSValue()))

	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("showModal")

	ctx.Async(func() {
		var output downballotapi.GetPersonResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person/"+c.VoterID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting person", "person", output.Person)
			c.Person = output.Person

			ctx.Update()
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

			ctx.Update()
		})
	})
}

func (c *AddFieldDialog) Close(ctx app.Context) {
	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("close")
}

func (c *AddFieldDialog) Render() app.UI {
	slog.InfoContext(context.TODO(), "AddFieldDialog: Render", "OrganizationID", c.OrganizationID, "VoterID", c.VoterID, "Person", c.Person)

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

	var valueElement app.UI
	if selectedPersonField != nil {
		switch selectedPersonField.Type {
		case downballotapi.PersonFieldDefinitionTypeDate:
			valueElement = myui.Input[string]().
				Label("Value").
				Type("date").
				Placeholder("Value").
				Bind(&c.ValueValue)
		case downballotapi.PersonFieldDefinitionTypeEnum:
			valueElement = myui.Select().
				AllowedValue(
					myui.SelectOption{Label: "", Value: "", Disabled: true},
				).
				AllowedValue(
					func() []myui.SelectOption {
						var allowedValues []myui.SelectOption
						for _, allowedValue := range selectedPersonField.AllowedValues {
							allowedValues = append(allowedValues, myui.SelectOption{Label: allowedValue, Value: allowedValue})
						}
						return allowedValues
					}()...).
				Bind(&c.ValueValue)
		case downballotapi.PersonFieldDefinitionTypeString:
			valueElement = myui.Input[string]().
				Label("Value").
				Type("text").
				Placeholder("Value").
				Bind(&c.ValueValue)
		case downballotapi.PersonFieldDefinitionTypeInteger:
			valueElement = myui.Input[string]().
				Label("Value").
				Type("number").
				Placeholder("Value").
				Bind(&c.ValueValue)
		case downballotapi.PersonFieldDefinitionTypeBoolean:
			valueElement = myui.Select().
				AllowedValue(
					myui.SelectOption{Label: "", Value: "", Disabled: true},
					myui.SelectOption{Label: "true", Value: "true"},
					myui.SelectOption{Label: "false", Value: "false"},
				).
				Bind(&c.ValueValue)
		default:
			valueElement = myui.Input[string]().
				Label("Value").
				Type("text").
				Placeholder("Value").
				Bind(&c.ValueValue)
		}
	}

	return app.Dialog().
		ID(c.DialogID).
		Body(
			app.H2().Text("Add Field"),
			app.Div().
				Body(
					myui.Select().
						Name("field").
						Label("Field").
						AllowedValue(
							myui.SelectOption{Label: "", Value: "", Disabled: true},
						).
						AllowedValue(
							func() []myui.SelectOption {
								var allowedValues []myui.SelectOption
								for _, personField := range c.PersonFields {
									allowedValues = append(allowedValues, myui.SelectOption{Label: personField.Name, Value: personField.Name})
								}
								return allowedValues
							}()...).
						Bind(&c.SelectedFieldValue).
						On("change", func(ctx app.Context, e app.Event) {
							slog.InfoContext(ctx.Context, "AddFieldDialog: OnChange", "SelectedFieldValue", c.SelectedFieldValue)

							if c.SelectedFieldValue == "" {
								c.ValueValue = ""
							} else {
								c.ValueValue = c.Person.Fields[c.SelectedFieldValue]
							}
							slog.InfoContext(ctx.Context, "AddFieldDialog: OnChange", "ValueValue", c.ValueValue)
							ctx.Update()
						}),
					valueElement,
				),
			app.If(c.error != "", func() app.UI {
				return myui.StatusBar().
					Text(c.error).
					Bad()
			}),
			app.Div().
				Class("myui-dialog-actions").
				Body(
					myui.Button().
						Label("Cancel").
						On("click", func(ctx app.Context, event app.Event) {
							c.Close(ctx)
						}),
					app.Span().Style("flex", "1"),
					app.If(c.SubmitFunctionValue != nil, func() app.UI {
						return myui.Button().
							Label("Save").
							On("click", func(ctx app.Context, event app.Event) {
								err := c.SubmitFunctionValue(ctx)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not submit", "err", err)
									c.error = err.Error()
									return
								}
								c.error = ""
								c.Close(ctx)
							})
					}),
				),
		)
}
