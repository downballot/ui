package page

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OrganizationIDPersonIDPage struct {
	app.Compo

	OrganizationID string `route:"organization_id"`
	VoterID        string `route:"voter_id"`
	Person         *downballotapi.Person
	Audits         []*downballotapi.PersonAudit
}

func (c *OrganizationIDPersonIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "VoterID", c.VoterID)

	if c.OrganizationID == "" {
		return
	}

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
	ctx.Async(func() {
		var output downballotapi.ListPersonAuditsResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person/"+c.VoterID+"/audit", nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get person audits", "err", err)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Dispatch: Setting person audits", "Audits", output.Audits)
			c.Audits = output.Audits
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Person audits should be set", "Audits", c.Audits)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDPersonIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonIDPage: Render", "OrganizationID", c.OrganizationID, "VoterID", c.VoterID, "Person", c.Person)

	return app.Div().Body(
		app.If(c.Person == nil, func() app.UI {
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
				myui.Table[Record]().
					Rows(rows).
					Columns(columns).
					Render(),
				myui.Table[*downballotapi.PersonAudit]().
					Rows(c.Audits).
					Columns([]myui.TableColumn[*downballotapi.PersonAudit]{
						{
							Name: "ID",
							Value: func(row *downballotapi.PersonAudit) any {
								return row.ID
							},
						},
						{
							Name: "Timestamp",
							Value: func(row *downballotapi.PersonAudit) any {
								return time.Time(row.Timestamp).Format(time.RFC3339)
							},
						},
						{
							Name: "Field",
							Value: func(row *downballotapi.PersonAudit) any {
								return row.Field
							},
						},
						{
							Name: "Old Value",
							Value: func(row *downballotapi.PersonAudit) any {
								if row.OldValue == nil {
									return ""
								}
								return *row.OldValue
							},
						},
						{
							Name: "New Value",
							Value: func(row *downballotapi.PersonAudit) any {
								if row.NewValue == nil {
									return ""
								}
								return *row.NewValue
							},
						},
					}).Render(),
			)
		}),
	)
}
