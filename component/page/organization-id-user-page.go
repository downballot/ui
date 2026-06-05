package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	users          []*downballotapi.User
}

func (c *OrganizationIDUserPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnNav", "OrganizationID", c.organizationID)

	c.Reload(ctx)
}

func (c *OrganizationIDUserPage) Reload(ctx app.Context) {
	if c.organizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListUsersResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get users", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
			c.users = output.Users
		})
	})
}

func (c *OrganizationIDUserPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserPage: Render")

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		myui.Table[*downballotapi.User]().
			Rows(c.users).
			Columns([]myui.TableColumn[*downballotapi.User]{
				{
					Name: "ID",
					Value: func(row *downballotapi.User) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.User) any {
						return row.Name
					},
					To: func(row *downballotapi.User) string {
						return fmt.Sprintf("/organization/%s/user/%s", c.organizationID, row.ID)
					},
				},
				{
					Name: "Owner",
					Value: func(row *downballotapi.User) any {
						return row.Owner
					},
				},
			}).
			Action(myui.TableAction{
				Name: "Add user",
				Icon: "plus",
				To: func() string {
					return fmt.Sprintf("/organization/%s/user/new", c.organizationID)
				},
			}).
			RowAction(myui.RowAction[*downballotapi.User]{
				Name: "Remove",
				Icon: "trash",
				Function: func(ctx app.Context, row *downballotapi.User) {
					ctx.PreventUpdate()

					result := app.Window().Call("confirm", "Are you sure you want to remove this user from the organization?")
					slog.InfoContext(ctx.Context, "OrganizationIDUserPage: Delete button clicked", "result", result.Bool())
					if !result.Bool() {
						slog.InfoContext(ctx.Context, "OrganizationIDUserPage: Delete button clicked: User cancelled", "result", result.Bool())
						return
					}

					slog.InfoContext(ctx.Context, "OrganizationIDUserPage: Deleting user", "user", row)
					err := api.Do(ctx, http.MethodDelete, "/api/v1/organization/"+c.organizationID+"/user/"+row.ID, nil, nil)
					if err != nil {
						slog.ErrorContext(ctx.Context, "Could not delete user", "err", err)
						return
					}
					c.Reload(ctx)
				},
			}),
	)
}
