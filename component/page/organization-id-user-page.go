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

type OrganizationIDUserPage struct {
	app.Compo
	myui.EmbeddedPage

	OrganizationID string `route:"organization_id"`
	Users          []*downballotapi.User
}

var _ app.Mounter = (*OrganizationIDUserPage)(nil)

func (c *OrganizationIDUserPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnMount")
	c.EmbeddedPage.Setup(ctx)
}

func (c *OrganizationIDUserPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDUserPage: OnUpdate", "OrganizationID", c.OrganizationID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.ListUsersResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/user", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get users", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting users", "users", output.Users)
			c.Users = output.Users
		})
	})
}

func (c *OrganizationIDUserPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserPage: Render")

	return c.EmbeddedPage.Wrap(
		myui.Table[*downballotapi.User]().
			Rows(c.Users).
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
						return fmt.Sprintf("/organization/%s/user/%s", c.OrganizationID, row.ID)
					},
				},
			}).Render(),
	)
}
