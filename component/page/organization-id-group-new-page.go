package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupNewPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Groups         []*downballotapi.Group
	ParentID       string `query:"parent_id"`
	Parent         *downballotapi.Group
	Name           string
	Filter         string
}

func (c *OrganizationIDGroupNewPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}
	})
	ctx.Async(func() {
		var output downballotapi.ListGroupsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting groups", "groups", output.Groups)
			c.Groups = output.Groups
		})
	})
}

func (c *OrganizationIDGroupNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav", "OrganizationID", c.OrganizationID)

	c.ParentID = ctx.Page().URL().Query().Get("parent_id")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav", "ParentID", c.ParentID)
}

func (c *OrganizationIDGroupNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupNewPage: Render")

	return app.Div().
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
					Label("Create").
					On("click", func(ctx app.Context, e app.Event) {
						ctx.Async(func() {
							var input downballotapi.CreateGroupRequest
							input.Name = c.Name
							input.ParentID = c.ParentID
							input.Filter = c.Filter
							var output downballotapi.CreateGroupResponse
							err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.OrganizationID+"/group", input, &output)
							if err != nil {
								slog.ErrorContext(ctx.Context, "Could not create group", "err", err)
								return
							}
							ctx.Navigate(fmt.Sprintf("/organization/%s/group/%s", c.OrganizationID, output.ID))
						})
					}),
			),
		)
}
