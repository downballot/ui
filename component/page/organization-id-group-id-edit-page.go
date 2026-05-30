package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDEditPage struct {
	app.Compo

	organizationID string
	groupID        string

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

			c.ParentID = output.Group.ParentID
			c.Name = output.Group.Name
			c.Filter = output.Group.Filter
		})
		wg.Wait()

		ctx.Update()
	})
}

func (c *OrganizationIDGroupIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDEditPage: Render")

	return myui.Page().Body(
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
						Label("Save").
						On("click", func(ctx app.Context, e app.Event) {
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
