package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserIDGroupIDEditPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	userID         string
	groupID        string

	user  *downballotapi.User
	group *downballotapi.Group

	owner bool
}

func (c *OrganizationIDUserIDGroupIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("user_id", &c.userID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: OnNav", "UserID", c.userID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: OnNav", "GroupID", c.groupID)

	c.Reload(ctx)
}

func (c *OrganizationIDUserIDGroupIDEditPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: Reload", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: Reload", "GroupID", c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.userID == "" {
		return
	}

	if c.groupID == "" {
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
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			c.group = output.Group
		})
		wg.Go(func() {
			var output downballotapi.GetGroupUserResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/user/"+c.userID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get user group", "err", err)
				return
			}

			c.owner = output.GroupUser.Owner
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDUserIDGroupIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDGroupIDEditPage: Render")

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	if c.user == nil {
		return myui.Page().Body(
			app.Div().Text("User not found"),
		)
	}

	if c.group == nil {
		return myui.Page().Body(
			app.Div().Text("Group not found"),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input[string]().
					Disabled(true).
					Label("E-mail Address").
					Value(c.user.Username),
				myui.Input[string]().
					Disabled(true).
					Label("Group").
					Value(c.group.Name),
				myui.Input[bool]().
					Name("owner").
					Label("Owner").
					Bind(&c.owner),
				app.Div().Body(
					myui.Button().
						Label("Delete").
						Icon("trash").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.PreventUpdate()

							result := app.Window().Call("confirm", "Are you sure you want to remove this user from this group?")
							slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: Delete button clicked", "result", result.Bool())
							if !result.Bool() {
								slog.InfoContext(ctx.Context, "OrganizationIDUserIDGroupIDEditPage: Delete button clicked: User cancelled", "result", result.Bool())
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
						}),
					myui.Button().
						Label("Save").
						Icon("save").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.PreventUpdate()

							ctx.Async(func() {
								input := downballotapi.PatchGroupUserRequest{
									Owner: &c.owner,
								}
								var output downballotapi.PatchGroupUserResponse
								err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/user/"+c.userID, input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not patch group user", "err", err)
									return
								}

								ctx.Navigate("/organization/" + c.organizationID + "/user/" + c.userID)
							})
						}),
				),
			),
	)
}
