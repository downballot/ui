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

type OrganizationIDGroupNewPage struct {
	app.Compo

	organizationID string
	groups         []*downballotapi.Group

	parentID string
	parent   *downballotapi.Group

	Name   string
	Filter string
}

func (c *OrganizationIDGroupNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	c.parentID = ctx.Page().URL().Query().Get("parent_id")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: OnNav", "ParentID", c.parentID)

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}
	})
	ctx.Async(func() {
		var output downballotapi.ListGroupsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting groups", "groups", output.Groups)
			c.groups = output.Groups
		})
	})
}

func (c *OrganizationIDGroupNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupNewPage: Render")

	return myui.Page().Body(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input[string]().
					Label("Name").
					Type("text").
					Value(c.Name).
					On("change", c.ValueTo(&c.Name)),
				myui.Input[string]().
					Label("Parent").
					Type("text").
					Value(c.parentID).
					On("change", c.ValueTo(&c.parentID)),
				myui.Input[string]().
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
								input.ParentID = c.parentID
								input.Filter = c.Filter
								var output downballotapi.CreateGroupResponse
								err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/group", input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not create group", "err", err)
									return
								}
								slog.InfoContext(ctx.Context, "OrganizationIDGroupNewPage: Create button clicked: Navigating to group page", "group_id", output.ID)
								ctx.Navigate(fmt.Sprintf("/organization/%s/group/%s", c.organizationID, output.ID))
							})
						}),
				),
			),
	)
}
