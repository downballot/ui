package page

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDUserIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	UserID         string `route:"user_id"`
	User           *downballotapi.User

	Groups []*downballotapi.Group

	GroupsTable *myui.MyUITable[*downballotapi.Group]
}

var _ app.Navigator = (*OrganizationIDUserIDPage)(nil)

func (c *OrganizationIDUserIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnNav", "OrganizationID", c.OrganizationID)

	ctx.Update()
}

func (c *OrganizationIDUserIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDUserIDPage: OnUpdate", "UserID", c.UserID)

	if c.OrganizationID == "" {
		return
	}

	c.GroupsTable = myui.Table[*downballotapi.Group]().
		PageSize(10)

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
			var output downballotapi.GetUserResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/user/"+c.UserID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting user", "user", output.User)
				c.User = output.User
			})
		})
		// TODO: GET THE GROUPS FOR THE USER
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true
		})
	})
}

func (c *OrganizationIDUserIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserIDPage: Render")

	if !c.Loaded {
		return nil
	}

	if c.User == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
			app.Div().
				Body(
					app.Div().Text("Name: " + c.User.Name),
				),
			//c.GroupsTable.Render(),
		)
}
