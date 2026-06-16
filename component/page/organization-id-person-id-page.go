package page

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/router"
	"github.com/google/uuid"
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

	addFieldDialog component.AddFieldDialog
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

	c.addFieldDialog = component.AddFieldDialog{
		OrganizationID: c.organizationID,
		VoterID:        c.voterID,
		DialogID:       "id-" + uuid.New().String(),
	}

	c.addFieldDialog.SubmitFunctionValue = func(ctx app.Context) error {
		slog.InfoContext(ctx.Context, "AddFieldDialog: SubmitFunctionValue", "SelectedFieldValue", c.addFieldDialog.SelectedFieldValue, "ValueValue", c.addFieldDialog.ValueValue)
		input := downballotapi.PatchPersonRequest{
			Fields: map[string]*string{},
		}
		input.Fields[c.addFieldDialog.SelectedFieldValue] = &c.addFieldDialog.ValueValue
		var output downballotapi.PatchPersonRequest
		err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/person/"+c.voterID, input, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not update person", "err", err)
			return err
		}

		c.Reload(ctx)
		return nil
	}

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

	return c.EmbeddedPage.Wrap(
		blazar.Table[Record]().
			Title("Fields").
			Rows(rows).
			Columns(columns).
			Action(blazar.TableAction{
				Name: "Add Field",
				Icon: "plus",
				Function: func(ctx app.Context) {
					c.addFieldDialog.Open(ctx)
				},
			}).
			RowAction(blazar.RowAction[Record]{
				Name: "Edit",
				Icon: "edit",
				Function: func(ctx app.Context, row Record) {
					c.addFieldDialog.SelectedFieldValue = row.Field
					c.addFieldDialog.ValueValue = row.Value
					c.addFieldDialog.Open(ctx)
				},
			}),
		&c.addFieldDialog,
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
				{
					Name: "User",
					Value: func(row *downballotapi.PersonAudit) any {
						return row.Username
					},
				},
			}),
	)
}
