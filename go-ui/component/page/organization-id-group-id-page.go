package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/go-ui/api"
	"github.com/downballot/ui/go-ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationIDGroupIDPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	GroupID        string `route:"group_id"`
	Group          *downballotapi.Group

	Filter  string
	Persons []*downballotapi.Person
}

func (c *OrganizationIDGroupIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "GroupID", c.GroupID)

	if c.OrganizationID == "" {
		return
	}

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting organization", "organization", output.Organization)
			c.Organization = &output.Organization
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Organization should be set", "organization", c.Organization)

			ctx.Update()
		})
	})
	ctx.Async(func() {
		var output downballotapi.GetGroupResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting group", "group", output.Group)
			c.Group = output.Group
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Group should be set", "group", c.Group)

			ctx.Update()
		})
	})
}

func (c *OrganizationIDGroupIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render")

	return app.Div().Body(
		app.If(c.Organization == nil || c.Group == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Body(
				app.Div().Text(fmt.Sprintf("%+v", *c.Organization)),
				app.Div().Text(fmt.Sprintf("%+v", *c.Group)),
				app.Hr(),
				app.Div().Body(
					app.Div().Text("Filter:"),
					app.Div().Body(
						app.Input().
							Placeholder("key = 'value' or ...").
							Value(c.Filter).
							OnChange(c.ValueTo(&c.Filter)),
					),
					app.Div().Body(
						app.Button().
							Text("Search").
							OnClick(func(ctx app.Context, e app.Event) {
								queryParameters := url.Values{}
								queryParameters.Set("filter", c.Filter)
								var output downballotapi.ListPersonsResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID+"/person?"+queryParameters.Encode(), nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get persons", "err", err)
									return
								}

								ctx.Dispatch(func(ctx app.Context) {
									slog.InfoContext(ctx.Context, "Dispatch: Setting persons", "persons", output.Persons)
									c.Persons = output.Persons
								})
								ctx.Defer(func(ctx app.Context) {
									slog.InfoContext(ctx.Context, "Defer: Persons should be set", "persons", c.Persons)

									ctx.Update()
								})
							}),
					),
				),
				myui.NewTable[*downballotapi.Person]().
					Rows(c.Persons).
					Columns([]myui.TableColumn[*downballotapi.Person]{
						{
							Name: "ID",
							Value: func(row *downballotapi.Person) any {
								return row.ID
							},
						},
						{
							Name: "Voter ID",
							Value: func(row *downballotapi.Person) any {
								return row.VoterID
							},
							To: func(row *downballotapi.Person) string {
								return fmt.Sprintf("/organization/%s/person/%s", c.OrganizationID, row.VoterID)
							},
						},
						{
							Name: "Fields",
							Value: func(row *downballotapi.Person) any {
								return fmt.Sprintf("%v", row.Fields)
							},
						},
					}).Render(),
			)
		}),
	)
}
