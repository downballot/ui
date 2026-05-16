package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
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
}

func (c *OrganizationIDGroupNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupNewPage: Render")

	return app.Div().
		Body(
			app.Div().Body(
				app.Span().Text("Name"),
				app.Input().
					Value(c.Name).
					OnChange(c.ValueTo(&c.Name)),
			),
			app.Div().Body(
				app.Span().Text("Parent"),
				app.Input().
					Value(c.ParentID).
					OnChange(c.ValueTo(&c.ParentID)),
			),
			app.Div().Body(
				app.Span().Text("Filter"),
				app.Input().
					Value(c.Filter).
					OnChange(c.ValueTo(&c.Filter)),
			),
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
