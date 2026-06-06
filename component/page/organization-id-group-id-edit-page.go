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

type OrganizationIDGroupIDEditPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	groupID        string
	group          *downballotapi.Group

	groups []*downballotapi.Group

	ParentID string
	Name     string
	Filter   string
}

func (c *OrganizationIDGroupIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnNav", "GroupID", c.groupID)

	c.Reload(ctx)
}

func (c *OrganizationIDGroupIDEditPage) Reload(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: Reload", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: Reload", "GroupID", c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.groupID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			c.groups = output.Groups
		})
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			c.group = output.Group

			c.ParentID = output.Group.ParentID
			c.Name = output.Group.Name
			c.Filter = output.Group.Filter
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDGroupIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDEditPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.group == nil {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Group not found"),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input[string]().
					Label("Name").
					Type("text").
					Bind(&c.Name),
				myui.Input[string]().
					Label("Parent").
					Type("text").
					Bind(&c.ParentID),
				myui.Input[string]().
					Label("Filter").
					Type("text").
					Bind(&c.Filter),
				app.Div().Body(
					myui.Button().
						Label("Delete").
						Icon("trash").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.PreventUpdate()

							result := app.Window().Call("confirm", "Are you sure you want to delete this group?")
							slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: Delete button clicked", "result", result.Bool())
							if !result.Bool() {
								slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: Delete button clicked: User cancelled", "result", result.Bool())
								return
							}

							ctx.Async(func() {
								err := api.Do(ctx, http.MethodDelete, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, nil, nil)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not delete group", "err", err)
									return
								}

								ctx.Navigate("/organization/" + c.organizationID + "/group")
							})
						}),
					myui.Button().
						Label("Save").
						Icon("save").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.PreventUpdate()

							ctx.Async(func() {
								var input downballotapi.PatchGroupRequest
								input.Name = &c.Name
								input.ParentID = &c.ParentID
								input.Filter = &c.Filter
								var output downballotapi.PatchGroupResponse
								err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not patch group", "err", err)
									return
								}

								c.Reload(ctx)
							})
						}),
				),
			),
	)
}
