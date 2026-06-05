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

type OrganizationIDUserNewPage struct {
	app.Compo
	myui.EmbeddedPage

	organizationID string

	EmailAddress string
	Owner        bool
}

func (c *OrganizationIDUserNewPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDUserNewPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDUserNewPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}
}

func (c *OrganizationIDUserNewPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDUserNewPage: Render")

	return c.EmbeddedPage.Wrap(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				myui.Input[string]().
					Label("E-mail Address").
					Bind(&c.EmailAddress),
				myui.Input[bool]().
					Label("Owner").
					Bind(&c.Owner),
				app.Div().Body(
					myui.Button().
						Label("Add User To Organization").
						On("click", func(ctx app.Context, e app.Event) {
							ctx.Async(func() {
								var input downballotapi.AddUserToOrganizationRequest
								input.Username = c.EmailAddress
								input.Owner = c.Owner
								var output downballotapi.AddUserToOrganizationResponse
								err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/user", input, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not add user to organization", "err", err)
									return
								}
								slog.InfoContext(ctx.Context, "OrganizationIDUserNewPage: Create button clicked: Navigating to user page")
								ctx.Navigate(fmt.Sprintf("/organization/%s/user/%s", c.organizationID, output.UserID))
							})
						}),
				),
			),
	)
}
