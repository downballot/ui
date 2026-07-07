package component

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/htmlevent"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type htmlAddFieldMultipleDialog struct {
	app.Compo

	IOrganizationID string
	voterIDs        []string

	error string

	personFields []*downballotapi.PersonField

	ISubmitName string
	IOnSubmit   func(ctx app.Context)

	selectedFieldName string
	selectedValue     string
}

func AddFieldMultipleDialog() *htmlAddFieldMultipleDialog {
	return &htmlAddFieldMultipleDialog{}
}

const (
	addFieldMultipleDialogEventOpen = "add-field-multiple-dialog-open"
)

func (c *htmlAddFieldMultipleDialog) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: OnMount")

	ctx.Handle(addFieldMultipleDialogEventOpen, func(ctx app.Context, e app.Action) {
		slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: Open", "Action", e)

		c.voterIDs = e.Value.([]string)

		c.JSValue().Call("showModal")
	})
}

func (c *htmlAddFieldMultipleDialog) OrganizationID(organizationID string) *htmlAddFieldMultipleDialog {
	c.IOrganizationID = organizationID
	return c
}

func (c *htmlAddFieldMultipleDialog) Open(ctx app.Context, voterIDs []string) {
	slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: Open", "voterIDs", voterIDs)
	slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: Open", "JSValue", c.JSValue(), "JSValue", app.Window().Get("JSON").Call("stringify", c.JSValue()))

	ctx.NewActionWithValue(addFieldMultipleDialogEventOpen, voterIDs)
}

func (c *htmlAddFieldMultipleDialog) Close(ctx app.Context) {
	c.JSValue().Call("close")
}

func (c *htmlAddFieldMultipleDialog) Render() app.UI {
	slog.InfoContext(context.TODO(), "htmlAddFieldMultipleDialog: Render", "OrganizationID", c.IOrganizationID, "voterIDs", c.voterIDs)

	submitName := c.ISubmitName
	if submitName == "" {
		submitName = "Submit"
	}

	var selectedPersonField *downballotapi.PersonField
	for _, personField := range c.personFields {
		if personField.Name == c.selectedFieldName {
			selectedPersonField = personField
			break
		}
	}
	slog.InfoContext(context.TODO(), "htmlAddFieldMultipleDialog: Render", "selectedPersonField", selectedPersonField)

	var valueElements []app.UI
	if selectedPersonField != nil {
		switch selectedPersonField.Type {
		case downballotapi.PersonFieldDefinitionTypeDate:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("date").
					Placeholder("Value").
					Bind(&c.selectedValue),
				blazar.Button().
					Flat(true).
					Label("Yesterday").
					On("click", func(ctx app.Context, e app.Event) {
						c.selectedValue = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
						ctx.Update()
					}),
				blazar.Button().
					Flat(true).
					Label("Today").
					On("click", func(ctx app.Context, e app.Event) {
						c.selectedValue = time.Now().Format("2006-01-02")
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
					Bind(&c.selectedValue),
			)
		case downballotapi.PersonFieldDefinitionTypeString:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.selectedValue),
			)
		case downballotapi.PersonFieldDefinitionTypeInteger:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("number").
					Placeholder("Value").
					Bind(&c.selectedValue),
			)
		case downballotapi.PersonFieldDefinitionTypeBoolean:
			valueElements = append(valueElements,
				blazar.Select().
					AllowedValue(
						blazar.SelectOption{Label: "", Value: "", Disabled: true},
						blazar.SelectOption{Label: "true", Value: "true"},
						blazar.SelectOption{Label: "false", Value: "false"},
					).
					Bind(&c.selectedValue),
			)
		default:
			valueElements = append(valueElements,
				blazar.Input[string]().
					Label("Value").
					Type("text").
					Placeholder("Value").
					Bind(&c.selectedValue),
			)
		}
	}

	formBody := []app.UI{
		app.Div().
			Text(fmt.Sprintf("This change will apply to %d person(s).", len(c.voterIDs))),
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
				slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: OnChange", "SelectedFieldValue", c.selectedFieldName)

				c.selectedValue = ""

				// TODO: Can we pick the first option if it's a dropdown?

				slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: OnChange", "ValueValue", c.selectedValue)
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
					slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: SubmitFunction", "SelectedFieldValue", c.selectedFieldName, "ValueValue", c.selectedValue)

					input := downballotapi.PostPersonUpdateRequest{
						VoterIDs: c.voterIDs,
						Fields:   map[string]*string{},
					}
					input.Fields[c.selectedFieldName] = &c.selectedValue
					var output downballotapi.PostPersonUpdateResponse
					err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.IOrganizationID+"/person/update", input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not update persons", "err", err)
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
			slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: OnToggle", "toggleEvent", fmt.Sprintf("%+v", toggleEvent))

			// Don't do anything if we're closing the dialog.
			if toggleEvent.NewState != "open" {
				return
			}

			ctx.Async(func() {
				var output downballotapi.ListPersonFieldsResponse
				err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.IOrganizationID+"/person-field", nil, &output)
				if err != nil {
					slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
					return
				}

				ctx.Dispatch(func(ctx app.Context) {
					slog.InfoContext(ctx.Context, "Dispatch: Setting person fields", "person fields", output.PersonFields)
					c.personFields = output.PersonFields

					slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: Open", "Src", ctx.Src(), "Src", fmt.Sprintf("%T", ctx.Src()))
					slog.InfoContext(ctx.Context, "htmlAddFieldMultipleDialog: Open", "JSSrc", ctx.JSSrc(), "JSSrc", app.Window().Get("JSON").Call("stringify", ctx.JSSrc()))

					ctx.Update()
				})
			})

		})
}
