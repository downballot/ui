package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDEditPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	GroupID        string `route:"group_id"`
	Groups         []*downballotapi.Group
	ParentID       string
	Name           string
	Filter         string
}

func (c *OrganizationIDGroupIDEditPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	c.Reload(ctx)
}

func (c *OrganizationIDGroupIDEditPage) Reload(ctx app.Context) {
	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			c.Groups = output.Groups
		})
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID, nil, &output)
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

func (c *OrganizationIDGroupIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDEditPage: OnNav", "OrganizationID", c.OrganizationID)
}

func (c *OrganizationIDGroupIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDEditPage: Render")

	return myui.Page().Body(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input().
					Label("Name").
					Type("text").
					Value(c.Name).
					On("change", c.ValueTo(&c.Name)),
				myui.Input().
					Label("Parent").
					Type("text").
					Value(c.ParentID).
					On("change", c.ValueTo(&c.ParentID)),
				myui.Input().
					Label("Filter").
					Type("text").
					Value(c.Filter).
					On("change", c.ValueTo(&c.Filter)),
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
								err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID, input, &output)
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
