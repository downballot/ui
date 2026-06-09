package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDPage struct {
	app.Compo
	myui.EmbeddedPage

	IChildrenVisibleColumns []string

	organizationID string
	groupID        string

	loaded        bool
	group         *downballotapi.Group
	children      []*downballotapi.Group
	permissionSet permissionset.PermissionSet
}

var _ app.Navigator = (*OrganizationIDGroupIDPage)(nil)

func (c *OrganizationIDGroupIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "url", ctx.Page().URL())
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "GroupID", c.groupID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "ActiveRoute", router.GetActiveRoute(ctx))

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.groupID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	ctx.Async(func() {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav: Async: Getting organization and group")

		var group *downballotapi.Group
		var children []*downballotapi.Group

		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			group = output.Group
		})
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group?parent_id="+c.groupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			children = output.Groups
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.group = group
			c.children = children
			c.loaded = true

			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav: Async: Dispatch", "self", fmt.Sprintf("%p", c), "Loaded", c.loaded)
			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav: Async: Dispatch", "self", fmt.Sprintf("%p", c), "c", *c)
		})
	})
}

func (c *OrganizationIDGroupIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render", "self", fmt.Sprintf("%p", c), "OrganizationID", c.organizationID, "GroupID", c.groupID, "Loaded", c.loaded)
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render", "self", fmt.Sprintf("%p", c), "c", *c)

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.group == nil {
		return c.EmbeddedPage.Wrap(
			myui.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return c.EmbeddedPage.
		Action(
			myui.PageAction{
				Name:     "Persons",
				Icon:     "people-group",
				To:       fmt.Sprintf("/organization/%s/group/%s/person", c.organizationID, c.groupID),
				Disabled: !c.permissionSet.Match(iam.IAMPersonRead),
			},
			myui.PageAction{
				Name:     "Edit",
				Icon:     "edit",
				To:       fmt.Sprintf("/organization/%s/group/%s/edit", c.organizationID, c.groupID),
				Disabled: c.group.ParentID == "" || !c.permissionSet.Match(iam.IAMGroupUpdate),
			},
		).
		Wrap(
			myui.Table[*downballotapi.Group]().
				Title("Sub-groups").
				Rows(c.children).
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
							return fmt.Sprintf("/organization/%s/group/%s", c.organizationID, row.ID)
						},
					},
				}).
				VisibleColumns(c.IChildrenVisibleColumns).
				BindVisibleColumns(&c.IChildrenVisibleColumns).
				Action(myui.TableAction{
					Name:     "New group",
					Icon:     "plus",
					To:       fmt.Sprintf("/organization/%s/group/new?parent_id=%s", c.organizationID, c.groupID),
					Disabled: !c.permissionSet.Match(iam.IAMGroupCreate),
				}).
				RowAction(
					myui.RowAction[*downballotapi.Group]{
						Name: "Persons",
						Icon: "people-group",
						To: func(row *downballotapi.Group) string {
							return fmt.Sprintf("/organization/%s/group/%s/person", c.organizationID, row.ID)
						},
						Disabled: !c.permissionSet.Match(iam.IAMPersonRead),
					},
					myui.RowAction[*downballotapi.Group]{
						Name: "Edit",
						Icon: "edit",
						To: func(row *downballotapi.Group) string {
							return fmt.Sprintf("/organization/%s/group/%s/edit", c.organizationID, row.ID)
						},
						Disabled: c.group.ParentID == "" || !c.permissionSet.Match(iam.IAMGroupUpdate),
					},
				),
		)
}
