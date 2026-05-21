package component

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/google/uuid"
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

func (c *AddFieldDialog) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "AddFieldDialog: OnUpdate")
	slog.InfoContext(ctx.Context, "AddFieldDialog: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "AddFieldDialog: OnUpdate", "VoterID", c.VoterID)

	if c.OrganizationID == "" {
		return
	}

	c.DialogID = "id-" + uuid.New().String()

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
	})
}

func (c *AddFieldDialog) Open(ctx app.Context) {
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
			valueElement = myui.Input().
				Label("Value").
				Type("date").
				Placeholder("Value").
				Value(c.ValueValue).
				On("change", c.ValueTo(&c.ValueValue))
		case downballotapi.PersonFieldDefinitionTypeEnum:
			valueElement = app.Select().
				Body(
					app.Option().
						Value("").
						Disabled(true).
						Selected(c.ValueValue == ""),
					app.Range(selectedPersonField.AllowedValues).Slice(func(i int) app.UI {
						allowedValue := selectedPersonField.AllowedValues[i]
						return app.Option().
							Text(allowedValue).
							Value(allowedValue).
							Selected(c.ValueValue == allowedValue)
					}),
				).
				OnChange(c.ValueTo(&c.ValueValue))
		case downballotapi.PersonFieldDefinitionTypeString:
			valueElement = myui.Input().
				Label("Value").
				Type("text").
				Placeholder("Value").
				Value(c.ValueValue).
				On("change", c.ValueTo(&c.ValueValue))
		case downballotapi.PersonFieldDefinitionTypeInteger:
			valueElement = myui.Input().
				Label("Value").
				Type("number").
				Placeholder("Value").
				Value(c.ValueValue).
				On("change", c.ValueTo(&c.ValueValue))
		case downballotapi.PersonFieldDefinitionTypeBoolean:
			valueElement = app.Select().
				Body(
					app.Option().
						Text("").
						Value("").
						Disabled(true),
					app.Option().
						Text("true").
						Value("true").
						Selected(c.ValueValue == "true"),
					app.Option().
						Text("false").
						Value("false").
						Selected(c.ValueValue == "false"),
				).
				OnChange(c.ValueTo(&c.ValueValue))
		default:
			valueElement = myui.Input().
				Label("Value").
				Type("text").
				Placeholder("Value").
				Value(c.ValueValue).
				On("change", c.ValueTo(&c.ValueValue))
		}
	}

	return app.Dialog().
		ID(c.DialogID).
		Body(
			app.H2().Text("Add Field"),
			app.Div().
				Body(
					app.Select().
						Name("field").
						Body(
							app.Option().
								Text("").
								Value("").
								Disabled(true).
								Selected(c.SelectedFieldValue == ""),
							app.Range(c.PersonFields).Slice(func(i int) app.UI {
								personField := c.PersonFields[i]
								return app.Option().
									Text(personField.Name).
									Value(personField.Name).
									Selected(c.SelectedFieldValue == personField.Name)
							}),
						).
						OnChange(func(ctx app.Context, e app.Event) {
							c.ValueTo(&c.SelectedFieldValue)(ctx, e)
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
