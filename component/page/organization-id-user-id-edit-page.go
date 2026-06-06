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

type OrganizationIDUserIDEditPage struct {
	app.Compo
	myui.EmbeddedPage

	organizationID string
	userID         string

	user *downballotapi.User

	Owner bool
}

func (c *OrganizationIDUserIDEditPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDEditPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("user_id", &c.userID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserIDEditPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDEditPage: OnNav", "UserID", c.userID)

	c.Reload(ctx)
}

func (c *OrganizationIDUserIDEditPage) Reload(ctx app.Context) {
	if c.organizationID == "" {
		return
	}

	if c.userID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetUserResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.user = output.User

			c.Owner = output.User.Owner
		})
	})
}

func (c *OrganizationIDUserIDEditPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDEditPage: Render")

	if c.user == nil {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input[string]().
					Disabled(true).
					Label("E-mail Address").
					Value(c.user.Username),
				myui.Input[bool]().
					Label("Owner").
					Bind(&c.Owner),
				app.Div().Body(
					myui.Button().
						Label("Save").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.Async(func() {
								input := downballotapi.PatchOrganizationUserRequest{
									Owner: &c.Owner,
								}
								var output downballotapi.PatchOrganizationUserResponse
								err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID, input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not patch user", "err", err)
									return
								}

								ctx.Navigate(fmt.Sprintf("/organization/%s/user", c.organizationID))
							})
						}),
				),
			),
	)
}
