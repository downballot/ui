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
	"github.com/downballot/ui/myui"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDPersonIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	VoterID        string `route:"voter_id"`
	Person         *downballotapi.Person
	Audits         []*downballotapi.PersonAudit

	AddFieldDialog component.AddFieldDialog
}

func (c *OrganizationIDPersonIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDPersonIDPage: OnUpdate", "VoterID", c.VoterID)

	if c.OrganizationID == "" {
		return
	}

	c.AddFieldDialog = component.AddFieldDialog{
		OrganizationID: c.OrganizationID,
		VoterID:        c.VoterID,
		DialogID:       "id-" + uuid.New().String(),
	}

	c.AddFieldDialog.SubmitFunctionValue = func(ctx app.Context) error {
		slog.InfoContext(ctx.Context, "AddFieldDialog: SubmitFunctionValue", "SelectedFieldValue", c.AddFieldDialog.SelectedFieldValue, "ValueValue", c.AddFieldDialog.ValueValue)
		input := downballotapi.PatchPersonRequest{
			Fields: map[string]*string{},
		}
		input.Fields[c.AddFieldDialog.SelectedFieldValue] = &c.AddFieldDialog.ValueValue
		var output downballotapi.PatchPersonRequest
		err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.OrganizationID+"/person/"+c.VoterID, input, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not update person", "err", err)
			return err
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.OnUpdate(ctx)
		})
		return nil
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
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
		})
		wg.Go(func() {
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
		})

		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true
		})
	})
}

func (c *OrganizationIDPersonIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonIDPage: Render", "OrganizationID", c.OrganizationID, "VoterID", c.VoterID, "Person", c.Person)

	if !c.Loaded {
		return nil
	}

	if c.Person == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

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

	return myui.Page().Body(
		myui.Table[Record]().
			Title("Fields").
			Rows(rows).
			Columns(columns).
			Action(myui.TableAction{
				Name: "Add Field",
				Icon: "plus",
				Function: func(ctx app.Context) {
					c.AddFieldDialog.Open(ctx)
				},
			}).
			RowAction(myui.RowAction[Record]{
				Name: "Edit",
				Icon: "edit",
				Function: func(ctx app.Context, row Record) {
					c.AddFieldDialog.SelectedFieldValue = row.Field
					c.AddFieldDialog.ValueValue = row.Value
					c.AddFieldDialog.Open(ctx)
				},
			}).
			Render(),
		c.AddFieldDialog.Render(),
		myui.Table[*downballotapi.PersonAudit]().
			Title("Audit Log").
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
				{
					Name: "User",
					Value: func(row *downballotapi.PersonAudit) any {
						return row.Username
					},
				},
			}).Render(),
	)
}
