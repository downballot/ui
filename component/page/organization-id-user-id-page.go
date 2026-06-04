package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserIDPage struct {
	app.Compo

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	userID         string
	user           *downballotapi.User

	groups []*downballotapi.Group
}

var _ app.Navigator = (*OrganizationIDUserIDPage)(nil)

func (c *OrganizationIDUserIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("user_id", &c.userID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav", "UserID", c.userID)

	if c.organizationID == "" {
		return
	}

	if c.userID == "" {
		return
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
			var output downballotapi.GetUserResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/user/"+c.userID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting user", "user", output.User)
				c.user = output.User
			})
		})
		// TODO: GET THE GROUPS FOR THE USER
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDUserIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDPage: Render")

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	if c.user == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
			app.Div().
				Body(
					app.Div().Text("Name: "+c.user.Name),
					app.Div().Text("E-mail address: "+c.user.Username),
				),
		)
}
