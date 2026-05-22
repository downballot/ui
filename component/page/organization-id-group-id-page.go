package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	GroupID        string `route:"group_id"`
	Group          *downballotapi.Group
	Children       []*downballotapi.Group
}

var _ app.Navigator = (*OrganizationIDGroupIDPage)(nil)

func (c *OrganizationIDGroupIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "OrganizationID", c.OrganizationID)
}

func (c *OrganizationIDGroupIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "GroupID", c.GroupID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting group", "group", output.Group)
				c.Group = output.Group
			})
		})
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group?parent_id="+c.GroupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting children", "groups", output.Groups)
				c.Children = output.Groups
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true
		})
	})
}

func (c *OrganizationIDGroupIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render")

	if !c.Loaded {
		return nil
	}

	if c.Group == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
			myui.Table[*downballotapi.Group]().
				Title("Sub-groups").
				Rows(c.Children).
				Columns([]myui.TableColumn[*downballotapi.Group]{
					{
						Name: "ID",
						Value: func(row *downballotapi.Group) any {
							return row.ID
						},
					},
					{
						Name: "Name",
						Value: func(row *downballotapi.Group) any {
							return row.Name
						},
						To: func(row *downballotapi.Group) string {
							return fmt.Sprintf("/organization/%s/group/%s", c.OrganizationID, row.ID)
						},
					},
				}).
				Action(myui.TableAction{
					Name: "New group",
					Icon: "plus",
					To: func() string {
						return fmt.Sprintf("/organization/%s/group/new?parent_id=%s", c.OrganizationID, c.GroupID)
					},
				}).
				RowAction(myui.RowAction[*downballotapi.Group]{
					Name: "Persons",
					Icon: "people-group",
					To: func(row *downballotapi.Group) string {
						return fmt.Sprintf("/organization/%s/group/%s/person", c.OrganizationID, row.ID)
					},
				}).
				Render(),
		)
}
