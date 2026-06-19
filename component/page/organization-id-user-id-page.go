package page

import (
	"context"
	"fmt"
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

type OrganizationIDUserIDPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	userID         string
	user           *downballotapi.User

	userGroups []*downballotapi.UserGroup
}

var _ app.Navigator = (*OrganizationIDUserIDPage)(nil)

func (c *OrganizationIDUserIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("user_id", &c.userID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav", "UserID", c.userID)

	c.Reload(ctx)
}

func (c *OrganizationIDUserIDPage) Reload(ctx app.Context) {
	if c.organizationID == "" {
		return
	}

	if c.userID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
			var output downballotapi.GetUserResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting user", "user", output.User)
				c.user = output.User
			})
		})
		wg.Go(func() {
			var output downballotapi.ListUserGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID+"/group", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting user groups", "user groups", output.UserGroups)
				c.userGroups = output.UserGroups
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDUserIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.user == nil {
		return c.EmbeddedPage.Wrap(
			blazar.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().
			Body(
				app.Div().Text("Name: "+c.user.Name),
				app.Div().Text("E-mail address: "+c.user.Username),
			),
		blazar.Table[*downballotapi.UserGroup]().
			Rows(c.userGroups).
			Columns([]blazar.TableColumn[*downballotapi.UserGroup]{
				{
					Name: "ID",
					Value: func(row *downballotapi.UserGroup) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.UserGroup) any {
						return row.Name
					},
					To: func(row *downballotapi.UserGroup) string {
						return fmt.Sprintf("/organization/%s/group/%s", c.organizationID, row.ID)
					},
				},
				{
					Name: "Owner",
					Value: func(row *downballotapi.UserGroup) any {
						return row.Owner
					},
				},
			}).
			Action(blazar.TableAction{
				Name: "Add to group",
				Icon: component.IconAdd,
				To:   fmt.Sprintf("/organization/%s/user/%s/group/new", c.organizationID, c.userID),
			}).
			RowAction(
				blazar.RowAction[*downballotapi.UserGroup]{
					Name: "Edit",
					Icon: component.IconEdit,
					To: func(row *downballotapi.UserGroup) string {
						return fmt.Sprintf("/organization/%s/user/%s/group/%s/edit", c.organizationID, c.userID, row.ID)
					},
				},
				blazar.RowAction[*downballotapi.UserGroup]{
					Name: "Remove",
					Icon: component.IconDelete,
					Function: func(ctx app.Context, row *downballotapi.UserGroup) {
						ctx.PreventUpdate()

						result := app.Window().Call("confirm", "Are you sure you want to remove this user from this group?")
						slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: Delete button clicked", "result", result.Bool())
						if !result.Bool() {
							slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: Delete button clicked: User cancelled", "result", result.Bool())
							return
						}

						ctx.Async(func() {
							err := api.Do(ctx, http.MethodDelete, "/api/v1/organization/"+c.organizationID+"/group/"+row.ID+"/user/"+c.userID, nil, nil)
							if err != nil {
								slog.ErrorContext(ctx.Context, "Could not remove user from group", "err", err)
								return
							}

							c.Reload(ctx)
						})
					},
				},
			),
	)
}
