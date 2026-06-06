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

type AddUserToGroupDialog struct {
	app.Compo

	OrganizationID string
	UserID         string
	OnSuccess      func(ctx app.Context)

	DialogID string
	error    string

	groups  []*downballotapi.Group
	groupID string
	owner   bool
}

func (c *AddUserToGroupDialog) Open(ctx app.Context) {
	slog.InfoContext(ctx.Context, "AddUserToGroupDialog: Open", "DialogID", c.DialogID)
	slog.InfoContext(ctx.Context, "AddUserToGroupDialog: Open", "JSValue", c.JSValue(), "JSValue", app.Window().Get("JSON").Call("stringify", c.JSValue()))

	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("showModal")

	ctx.Async(func() {
		var output downballotapi.ListGroupsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting groups", "groups", output.Groups)
			c.groups = output.Groups

			// TODO: WE NEED A TABLE-LIKE BIND THING HERE SO THAT THE PARENT CAN BE UPDATED.

			ctx.Update()
		})
	})
}

func (c *AddUserToGroupDialog) Close(ctx app.Context) {
	dialogElement := app.Window().GetElementByID(c.DialogID)
	if dialogElement == nil || dialogElement.IsNull() {
		slog.ErrorContext(context.TODO(), "Could not get dialog element", "dialogID", c.DialogID)
		return
	}
	dialogElement.Call("close")
}

func (c *AddUserToGroupDialog) Render() app.UI {
	slog.InfoContext(context.TODO(), "AddUserToGroupDialog: Render", "OrganizationID", c.OrganizationID, "UserID", c.UserID, "groups", len(c.groups))

	return app.Dialog().
		ID(c.DialogID).
		Body(
			app.H2().Text("Add User To Group"),
			app.Div().
				Body(
					myui.Select().
						Name("group_id").
						Label("Group").
						AllowedValue(
							myui.SelectOption{Label: "Select a group", Value: "", Disabled: true},
						).
						AllowedValue(
							func() []myui.SelectOption {
								var allowedValues []myui.SelectOption
								for _, group := range c.groups {
									allowedValues = append(allowedValues, myui.SelectOption{Label: group.Name, Value: group.ID})
								}
								return allowedValues
							}()...).
						Bind(&c.groupID),
					myui.Input[bool]().
						Label("Owner").
						Bind(&c.owner),
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
					myui.Button().
						Label("Add").
						On("click", func(ctx app.Context, event app.Event) {
							ctx.Async(func() {
								input := downballotapi.AddUserToGroupRequest{
									GroupID: c.groupID,
									Owner:   c.owner,
								}
								var output downballotapi.AddUserToGroupResponse
								err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.OrganizationID+"/user/"+c.UserID+"/group", input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not add user to group", "err", err, "input", input, "output", output)
									return
								}
								c.Close(ctx)

								if c.OnSuccess != nil {
									c.OnSuccess(ctx)
								}
							})
						}),
				),
		)
}
