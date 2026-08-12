package customlayout

import (
	"context"
	"log/slog"
	"strings"

	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationMenu struct {
	app.Compo

	OrganizationID string

	menuItems []OrganizationMenuItem
}

var _ app.Updater = (*OrganizationMenu)(nil)

type OrganizationMenuItem struct {
	Icon string
	Name string
	To   string
}

var allMenuItems = []OrganizationMenuItem{
	{
		Icon: component.IconHome,
		Name: "Home",
		To:   "/organization/:organization_id",
	},
	{
		Icon: component.IconSearch,
		Name: "Search Persons",
		To:   "/organization/:organization_id/person/search",
	},
	{
		Icon: component.IconGroup,
		Name: "Groups",
		To:   "/organization/:organization_id/group",
	},
	{
		Icon: component.IconFilter,
		Name: "Filters",
		To:   "/organization/:organization_id/filter",
	},
	{
		Icon: component.IconPersonField,
		Name: "Person Fields",
		To:   "/organization/:organization_id/person-field",
	},
	{
		Icon: component.IconUser,
		Name: "Users",
		To:   "/organization/:organization_id/user",
	},
}

func (c *OrganizationMenu) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationMenu: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationMenu: OnUpdate", "OrganizationID", c.OrganizationID)

	var permissionSet permissionset.PermissionSet
	ctx.GetState("organization/"+c.OrganizationID+"/permission-set", &permissionSet)

	for _, item := range allMenuItems {
		newItem := item
		newItem.To = strings.ReplaceAll(newItem.To, ":organization_id", c.OrganizationID)

		route := router.GetRoute(ctx, newItem.To)
		if route != nil {
			permissionRequired := permissionset.Permission(route.Meta[MetaPermission])
			if len(permissionRequired) > 0 {
				if !permissionSet.Match(permissionRequired) {
					continue
				}
			}
		}

		c.menuItems = append(c.menuItems, newItem)
	}
	slog.InfoContext(ctx.Context, "OrganizationMenu: OnUpdate", "menuItems", c.menuItems)
}

func (c *OrganizationMenu) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationMenu: Render", "menuItems", c.menuItems)

	return app.Div().
		Class("organizationlayout-menu").
		Body(
			app.Range(c.menuItems).Slice(func(i int) app.UI {
				item := c.menuItems[i]

				return blazar.Item().
					Icon(item.Icon).
					Label(item.Name).
					To(item.To)
			}),
		)
}
