package component

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/htmlevent"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type htmlAddFieldDialog struct {
	app.Compo

	IOrganizationID string
	IVoterID        string
	person          *downballotapi.Person

	error string

	personFields []*downballotapi.PersonField

	ISubmitName string
	IOnSubmit   func(ctx app.Context)

	selectedFieldName   string
	selectedPersonField *downballotapi.PersonField
	value               string
}

var _ app.Mounter = (*htmlAddFieldDialog)(nil)

func AddFieldDialog() *htmlAddFieldDialog {
	return &htmlAddFieldDialog{}
}

const (
	addFieldDialogEventOpen = "add-field-dialog-open"
)

func (c *htmlAddFieldDialog) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "htmlAddFieldDialog: OnMount")

	ctx.Handle(addFieldDialogEventOpen, func(ctx app.Context, e app.Action) {
		slog.InfoContext(ctx.Context, "htmlAddFieldDialog: Open", "Action", e)

		c.selectedFieldName = e.Value.(addFieldDialogOpenValue).FieldName
		c.value = e.Value.(addFieldDialogOpenValue).FieldValue

		c.JSValue().Call("showModal")
	})
}

func (c *htmlAddFieldDialog) OnSubmit(onSubmit func(ctx app.Context)) *htmlAddFieldDialog {
	c.IOnSubmit = onSubmit
	return c
}

func (c *htmlAddFieldDialog) OrganizationID(organizationID string) *htmlAddFieldDialog {
	c.IOrganizationID = organizationID
	return c
}

func (c *htmlAddFieldDialog) VoterID(voterID string) *htmlAddFieldDialog {
	c.IVoterID = voterID
	return c
}

type addFieldDialogOpenValue struct {
	FieldName  string
	FieldValue string
}

func (c *htmlAddFieldDialog) Open(ctx app.Context, fieldName string, fieldValue string) {
	slog.InfoContext(ctx.Context, "htmlAddFieldDialog: Open", "fieldName", fieldName, "fieldValue", fieldValue)
	slog.InfoContext(ctx.Context, "htmlAddFieldDialog: Open", "JSValue", c.JSValue(), "JSValue", app.Window().Get("JSON").Call("stringify", c.JSValue()))

	ctx.NewActionWithValue(addFieldDialogEventOpen, addFieldDialogOpenValue{FieldName: fieldName, FieldValue: fieldValue})
}

func (c *htmlAddFieldDialog) Close(ctx app.Context) {
	c.JSValue().Call("close")
}

func (c *htmlAddFieldDialog) Render() app.UI {
	slog.InfoContext(context.TODO(), "htmlAddFieldDialog: Render", "IOrganizationID", c.IOrganizationID, "IVoterID", c.IVoterID, "person", c.person)

	submitName := c.ISubmitName
	if submitName == "" {
		submitName = "Submit"
	}

	var valueElements []app.UI
	if c.selectedPersonField != nil {
		switch c.selectedPersonField.Type {
		case downballotapi.PersonFieldDefinitionTypeDate:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("date").
					Placeholder("Value").
					Bind(&c.value),
				blazar.Button().
					Flat(true).
					Label("Yesterday").
					On("click", func(ctx app.Context, e app.Event) {
						c.value = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
						ctx.Update()
					}),
				blazar.Button().
					Flat(true).
					Label("Today").
					On("click", func(ctx app.Context, e app.Event) {
						c.value = time.Now().Format("2006-01-02")
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
							for _, allowedValue := range c.selectedPersonField.AllowedValues {
								allowedValues = append(allowedValues, blazar.SelectOption{Label: allowedValue, Value: allowedValue})
							}
							return allowedValues
						}()...).
					Bind(&c.value),
			)
		case downballotapi.PersonFieldDefinitionTypeString:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.value),
			)
		case downballotapi.PersonFieldDefinitionTypeInteger:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("number").
					Placeholder("Value").
					Bind(&c.value),
			)
		case downballotapi.PersonFieldDefinitionTypeBoolean:
			valueElements = append(valueElements,
				blazar.Select().
					AllowedValue(
						blazar.SelectOption{Label: "", Value: "", Disabled: true},
						blazar.SelectOption{Label: "true", Value: "true"},
						blazar.SelectOption{Label: "false", Value: "false"},
					).
					Bind(&c.value),
			)
		default:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.value),
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
					for _, personField := range c.personFields {
						allowedValues = append(allowedValues, blazar.SelectOption{Label: personField.Name, Value: personField.Name})
					}
					return allowedValues
				}()...).
			Bind(&c.selectedFieldName).
			On("change", func(ctx app.Context, e app.Event) {
				slog.InfoContext(ctx.Context, "htmlAddFieldDialog: OnChange", "selectedFieldName", c.selectedFieldName)

				if c.selectedFieldName == "" {
					c.value = ""
				} else {
					c.value = c.person.Fields[c.selectedFieldName]
				}

				c.selectedPersonField = nil
				for _, personField := range c.personFields {
					if personField.Name == c.selectedFieldName {
						c.selectedPersonField = personField
						break
					}
				}
				slog.InfoContext(context.TODO(), "htmlAddFieldDialog: OnChange", "selectedPersonField", c.selectedPersonField)

				c.value = ""
				if c.selectedPersonField != nil {
					if existingValue, ok := c.person.Fields[c.selectedFieldName]; ok {
						c.value = existingValue
					} else {
						// Pick the first option if it's a dropdown.
						switch c.selectedPersonField.Type {
						case downballotapi.PersonFieldDefinitionTypeEnum:
							if len(c.selectedPersonField.AllowedValues) > 0 {
								c.value = c.selectedPersonField.AllowedValues[0]
							}
						case downballotapi.PersonFieldDefinitionTypeBoolean:
							c.value = "true"
						default:
							c.value = ""
						}
					}
				}

				slog.InfoContext(ctx.Context, "htmlAddFieldDialog: OnChange", "value", c.value)
				ctx.Update()
			}),
	}
	formBody = append(formBody, valueElements...)

	return app.Dialog().
		Body(
			app.H2().Text("Add Field"),
			blazar.Form().
				Body(formBody...).
				CancelFunction(c.Close).
				SubmitLabel("Save").
				SubmitFunction(func(ctx app.Context) {
					slog.InfoContext(ctx.Context, "htmlAddFieldDialog: SubmitFunction", "selectedFieldName", c.selectedFieldName, "value", c.value)
					input := downballotapi.PatchPersonRequest{
						Fields: map[string]*string{},
					}
					input.Fields[c.selectedFieldName] = &c.value
					var output downballotapi.PatchPersonRequest
					err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.IOrganizationID+"/person/"+c.IVoterID, input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not update person", "err", err)
						return
					}

					if c.IOnSubmit != nil {
						c.IOnSubmit(ctx)
					}

					c.error = ""
					c.Close(ctx)
				}),
			app.If(c.error != "", func() app.UI {
				return blazar.StatusBar().
					Text(c.error).
					Bad()
			}),
		).
		On("toggle", func(ctx app.Context, e app.Event) {
			var toggleEvent htmlevent.Toggle
			htmlevent.MustParse(e, &toggleEvent)
			slog.InfoContext(ctx.Context, "htmlAddFieldDialog: OnToggle", "toggleEvent", fmt.Sprintf("%+v", toggleEvent))

			// Don't do anything if we're closing the dialog.
			if toggleEvent.NewState != "open" {
				return
			}

			ctx.Async(func() {
				var person *downballotapi.Person
				var personFields []*downballotapi.PersonField

				var wg sync.WaitGroup
				wg.Go(func() {
					var output downballotapi.GetPersonResponse
					err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.IOrganizationID+"/person/"+c.IVoterID, nil, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not get person", "err", err)
						return
					}

					person = output.Person
				})
				wg.Go(func() {
					var output downballotapi.ListPersonFieldsResponse
					err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.IOrganizationID+"/person-field", nil, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
						return
					}

					personFields = output.PersonFields
				})
				wg.Wait()

				ctx.Dispatch(func(ctx app.Context) {
					slog.InfoContext(ctx.Context, "Dispatch: Setting person and fields.")
					c.person = person
					c.personFields = personFields

					ctx.Update()
				})
			})
		})
}
