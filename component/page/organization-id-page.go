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
		app.If(c.permissionSet.Match(iam.IAMPersonRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Href("/organization/"+c.organizationID+"/person/search").
							Body(
								blazar.Icon().
									Icon(component.IconSearch),
								app.Text(" Search"),
							),
					),
					app.Div().Body(
						app.Span().Text("Search for persons by name, address, phone number, etc."),
					),
				)
		}),
		app.If(c.permissionSet.Match(iam.IAMGroupRead), func() app.UI {
			return app.Div().
				Body(
					app.Div().Body(
						app.A().
							Href("/organization/"+c.organizationID+"/group").
							Body(
								blazar.Icon().
									Icon(component.IconGroup),
								app.Text(" Groups"),
							),
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
							Href("/organization/"+c.organizationID+"/filter").
							Body(
								blazar.Icon().
									Icon(component.IconFilter),
								app.Text(" Filters"),
							),
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
							Href("/organization/"+c.organizationID+"/person-field").
							Body(
								blazar.Icon().
									Icon(component.IconPerson),
								app.Text(" Person Fields"),
							),
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
							Href("/organization/"+c.organizationID+"/user").
							Body(
								blazar.Icon().
									Icon(component.IconUser),
								app.Text(" Users"),
							),
					),
					app.Div().Body(
						app.Span().Text("View and manage users in this organization."),
					),
				)
		}),
	)
}
