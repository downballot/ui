package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	router "github.com/downballot/ui/app-router"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	users          []*downballotapi.User
}

var _ app.Mounter = (*OrganizationIDUserPage)(nil)

func (c *OrganizationIDUserPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnMount")

	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: Setting up embedded page")
	c.EmbeddedPage.Setup(ctx)
}

func (c *OrganizationIDUserPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListUsersResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get users", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
			c.users = output.Users
		})
	})
}

func (c *OrganizationIDUserPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserPage: Render")

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		myui.Table[*downballotapi.User]().
			Rows(c.users).
			Columns([]myui.TableColumn[*downballotapi.User]{
				{
					Name: "ID",
					Value: func(row *downballotapi.User) any {
						return row.ID
					},
				},
				{
					Name: "Name",
					Value: func(row *downballotapi.User) any {
						return row.Name
					},
					To: func(row *downballotapi.User) string {
						return fmt.Sprintf("/organization/%s/user/%s", c.organizationID, row.ID)
					},
				},
			}),
	)
}
