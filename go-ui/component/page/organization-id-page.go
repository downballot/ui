package page

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/downballot/ui/go-ui/component/customlayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func init() {
	app.RouteWithRegexp("^/organization/([^/]+)$", func() app.Composer { return &OrganizationIDPage{} })
}

type OrganizationIDPage struct {
	app.Compo

	organizationID string
	organization   *downballotapi.Organization
}

func (c *OrganizationIDPage) OnNav(ctx app.Context) {
	r := regexp.MustCompile("^/organization/([^/]+)$")
	matches := r.FindStringSubmatch(ctx.Page().URL().Path)
	slog.InfoContext(ctx.Context, "Matched regex", "matches", matches)
	if len(matches) < 2 {
		slog.ErrorContext(ctx.Context, "Could now match path properly.")
		return
	}
	c.organizationID = matches[1]
	slog.InfoContext(ctx.Context, "Organization ID", "id", c.organizationID)

	var output downballotapi.GetOrganizationResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
		return
	}
	c.organization = &output.Organization
}

func (c *OrganizationIDPage) OnMount(ctx app.Context) {
}

func (c *OrganizationIDPage) Render() app.UI {
	return &customlayout.DownballotLayout{
		Content: app.If(c.organization == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Text(fmt.Sprintf("%+v", *c.organization))
		}),
	}
}
