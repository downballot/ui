package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserIDGroupNewPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	userID         string

	user   *downballotapi.User
	groups []*downballotapi.Group

	groupID string
	owner   bool
}

func (c *OrganizationIDUserIDGroupNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("user_id", &c.userID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: OnNav", "UserID", c.userID)

	c.Reload(ctx)
}

func (c *OrganizationIDUserIDGroupNewPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: Reload", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: Reload", "UserID", c.userID)

	if c.organizationID == "" {
		return
	}

	if c.userID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetUserResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			c.user = output.User
		})
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			c.groups = output.Groups
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDUserIDGroupNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDGroupNewPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.user == nil {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("User not found"),
		)
	}

	return c.EmbeddedPage.Wrap(
		blazar.Form().
			Body(
				blazar.Input[string]().
					Disabled(true).
					Label("E-mail Address").
					Value(c.user.Username),
				blazar.Select().
					Name("group_id").
					Label("Group").
					AllowedValue(
						func() []blazar.SelectOption {
							var allowedValues []blazar.SelectOption
							allowedValues = append(allowedValues, blazar.SelectOption{Label: "Select a group", Value: "", Disabled: true})
							for _, group := range c.groups {
								allowedValues = append(allowedValues, blazar.SelectOption{Label: group.Name, Value: group.ID})
							}
							return allowedValues
						}()...).
					Bind(&c.groupID),
				blazar.Input[bool]().
					Label("Owner").
					Bind(&c.owner),
			).
			SubmitLabel("Add User To Group").
			SubmitIcon(component.IconSave).
			SubmitFunction(func(ctx app.Context) {
				ctx.PreventUpdate()

				ctx.Async(func() {
					input := downballotapi.AddUserToGroupRequest{
						GroupID: c.groupID,
						Owner:   c.owner,
					}
					var output downballotapi.PatchGroupUserResponse
					err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID+"/group", input, &output)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not add user to group", "err", err)
						return
					}

					ctx.Navigate("/organization/" + c.organizationID + "/user/" + c.userID)
				})
			}).
			Action(blazar.FormAction{
				Name: "Delete",
				Icon: component.IconDelete,
				Function: func(ctx app.Context) {
					ctx.PreventUpdate()

					result := app.Window().Call("confirm", "Are you sure you want to remove this user from this group?")
					slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: Delete button clicked", "result", result.Bool())
					if !result.Bool() {
						slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupNewPage: Delete button clicked: User cancelled", "result", result.Bool())
						return
					}

					ctx.Async(func() {
						err := api.Do(ctx, http.MethodDelete, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/user/"+c.userID, nil, nil)
						if err != nil {
							slog.ErrorContext(ctx.Context, "Could not delete user from group", "err", err)
							return
						}

						ctx.Navigate("/organization/" + c.organizationID + "/user/" + c.userID)
					})
				},
			}),
	)
}
