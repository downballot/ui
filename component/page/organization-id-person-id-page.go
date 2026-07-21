package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonIDPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	voterID        string
	person         *downballotapi.Person
	audits         []*downballotapi.PersonAudit
	permissionSet  permissionset.PermissionSet
}

func (c *OrganizationIDPersonIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("voter_id", &c.voterID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnNav", "VoterID", c.voterID)

	if c.organizationID == "" {
		return
	}

	if c.voterID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	c.Reload(ctx)
}

func (c *OrganizationIDPersonIDPage) Reload(ctx app.Context) {
	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetPersonResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person/"+c.voterID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting person", "person", output.Person)
				c.person = output.Person
			})
		})
		wg.Go(func() {
			var output downballotapi.ListPersonAuditsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person/"+c.voterID+"/audit", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person audits", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting person audits", "Audits", output.Audits)
				c.audits = output.Audits
			})
		})

		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true
		})
	})
}

func (c *OrganizationIDPersonIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonIDPage: Render", "OrganizationID", c.organizationID, "VoterID", c.voterID, "Person", c.person)

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.person == nil {
		return c.EmbeddedPage.Wrap(
			blazar.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	type Record struct {
		Field string
		Value string
	}

	columns := []blazar.TableColumn[Record]{
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
	for name := range c.person.Fields {
		rows = append(rows, Record{Field: name, Value: c.person.Fields[name]})
	}
	slices.SortFunc(rows, func(left, right Record) int {
		return strings.Compare(left.Field, right.Field)
	})

	addFieldDialog := component.AddFieldDialog().
		OrganizationID(c.organizationID).
		VoterID(c.voterID).
		OnSubmit(func(ctx app.Context) {
			c.Reload(ctx)
		})

	icon := "circle-question"
	switch c.person.Fields["candidate.support"] {
	case "-2":
		icon = "thumbs-down"
	case "-1":
		icon = "thumbs-down"
	case "0":
		icon = "circle-question"
	case "+1":
		icon = "thumbs-up"
	case "+2":
		icon = "thumbs-up"
	}

	summaryItems := []app.UI{
		app.Div().
			Class("person-summary-header").
			Body(
				blazar.Icon().
					Icon(icon),
				app.Div().
					Text(c.person.Fields["name"]),
			),
		app.Div().
			Class("person-summary-registration").
			Text("Registered as " + c.person.Fields["political_party"] + ", living in " + c.person.Fields["district_representative"] + ", " + c.person.Fields["district_senate"]),
		app.Div().
			Class("person-summary-address").
			Body(
				app.A().
					Href(fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(c.person.Fields["residential_address"]))).
					Target("_blank").
					Text(c.person.Fields["residential_address"]),
			),
		app.Div().
			Class("person-summary-phone").
			Text(c.person.Fields["phone_number"]),
		app.Div().
			Class("person-summary-chips").
			Body(
				app.If(c.person.Fields["candidate.connected"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("handshake"),
							app.Text("Connected"),
						)
				}),
				app.If(c.person.Fields["candidate.support"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Support: " + c.person.Fields["candidate.support"])
				}),
				app.If(c.person.Fields["candidate.cat"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("cat"),
							app.Text("Cat"),
						)
				}),
				app.If(c.person.Fields["candidate.dog"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("dog"),
							app.Text("Dog"),
						)
				}),
				app.If(c.person.Fields["candidate.date_called"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Called on " + c.person.Fields["candidate.date_called"])
				}),
				app.If(c.person.Fields["candidate.date_canvassed"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Canvassed on " + c.person.Fields["candidate.date_canvassed"])
				}),
				app.If(c.person.Fields["candidate.date_texted"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Texted on " + c.person.Fields["candidate.date_texted"])
				}),
			),
	}

	return c.EmbeddedPage.Wrap(
		app.Div().
			Class("person-summary").
			Body(summaryItems...),
		blazar.Table[Record]().
			Title("Fields").
			Rows(rows).
			Columns(columns).
			Action(blazar.TableAction{
				Name: "Add Field",
				Icon: component.IconAdd,
				Function: func(ctx app.Context) {
					addFieldDialog.Open(ctx, "", "")
				},
			}).
			RowAction(
				blazar.RowAction[Record]{
					Name: "Edit",
					Icon: component.IconEdit,
					Function: func(ctx app.Context, row Record) {
						addFieldDialog.Open(ctx, row.Field, row.Value)
					},
					Disabled: !c.permissionSet.Match(iam.IAMPersonUpdate),
				},
				blazar.RowAction[Record]{
					Name: "Delete",
					Icon: component.IconDelete,
					Function: func(ctx app.Context, row Record) {
						result := app.Window().Call("confirm", "Are you sure you want to delete this field ("+row.Field+")?")
						slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: Delete button clicked", "result", result.Bool())
						if !result.Bool() {
							slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: Delete button clicked: User cancelled", "result", result.Bool())
							return
						}

						ctx.Async(func() {
							input := downballotapi.PatchPersonRequest{
								Fields: map[string]*string{
									row.Field: nil,
								},
							}
							err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/person/"+c.voterID, input, nil)
							if err != nil {
								slog.ErrorContext(ctx.Context, "Could not delete field", "err", err)
								return
							}

							c.Reload(ctx)
						})
					},
					Disabled: !c.permissionSet.Match(iam.IAMPersonUpdate),
				},
			),
		addFieldDialog,
		blazar.Table[*downballotapi.PersonAudit]().
			Title("Audit Log").
			Rows(c.audits).
			Columns([]blazar.TableColumn[*downballotapi.PersonAudit]{
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
							return "—"
						}
						return *row.OldValue
					},
				},
				{
					Name: "New Value",
					Value: func(row *downballotapi.PersonAudit) any {
						if row.NewValue == nil {
							return "—"
						}
						return *row.NewValue
					},
				},
				{
					Name: "User",
					Value: func(row *downballotapi.PersonAudit) any {
						return row.Username
					},
				},
			}),
	)
}
