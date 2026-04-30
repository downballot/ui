package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationIDPersonIDPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	VoterID        string `route:"voter_id"`
	Person         *downballotapi.Person
}

func (c *OrganizationIDPersonIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "VoterID", c.VoterID)

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

			//ctx.Update()
		})
	})
	ctx.Async(func() {
		var output downballotapi.GetPersonResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person/"+c.VoterID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting person", "person", output.Person)
			c.Person = output.Person
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Person should be set", "person", c.Person)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDPersonIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonIDPage: Render", "OrganizationID", c.OrganizationID, "Organization", c.Organization, "VoterID", c.VoterID, "Person", c.Person)

	return app.Div().Body(
		app.If(c.Organization == nil || c.Person == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			type Record struct {
				Field string
				Value string
			}

			columns := []myui.TableColumn[Record]{
				{
					Name: "Field",
					Value: func(row Record) any {
						return row.Field
					},
				},
				{
					Name: "Value",
					Value: func(row Record) any {
						return row.Value
					},
				},
			}

			rows := []Record{}
			for name := range c.Person.Fields {
				rows = append(rows, Record{Field: name, Value: c.Person.Fields[name]})
			}
			slices.SortFunc(rows, func(left, right Record) int {
				return strings.Compare(left.Field, right.Field)
			})

			return app.Div().Body(
				app.Div().Text(fmt.Sprintf("%+v", *c.Organization)),
				app.Hr(),
				myui.NewTable[Record]().
					Rows(rows).
					Columns(columns).
					Render(),
			)
		}),
	)
}
