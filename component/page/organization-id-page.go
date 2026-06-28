package page

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	permissionSet  permissionset.PermissionSet
}

func (c *OrganizationIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Setting organization", "organization", output.Organization)
			c.organization = &output.Organization
		})
	})
}

func (c *OrganizationIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPage: Render")

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.organization == nil {
		return c.EmbeddedPage.Wrap(
			blazar.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return c.EmbeddedPage.Wrap(
		app.If(c.permissionSet.Match(iam.IAMGroupRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Text("Groups").
							Href("/organization/"+c.organizationID+"/group"),
					),
					app.Div().Body(
						app.Span().Text("View and manage groups, which are used to subdivide the persons.  They can also be used for user permissions."),
					),
				)
		}),
		app.If(c.permissionSet.Match(iam.IAMFilterRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Text("Filters").
							Href("/organization/"+c.organizationID+"/filter"),
					),
					app.Div().Body(
						app.Span().Text("View and manage filters, which can be used on top of groups to further limit the results."),
					),
				)
		}),
		app.If(c.permissionSet.Match(iam.IAMPersonFieldDefinitionRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Text("Person Fields").
							Href("/organization/"+c.organizationID+"/person-field"),
					),
					app.Div().Body(
						app.Span().Text("View and manage person fields, which are the fields available for each person."),
					),
				)
		}),
		app.If(c.permissionSet.Match(iam.IAMGroupUserRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Text("Users").
							Href("/organization/"+c.organizationID+"/user"),
					),
					app.Div().Body(
						app.Span().Text("View and manage users in this organization."),
					),
				)
		}),
	)
}
