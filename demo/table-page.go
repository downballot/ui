package demo

import (
	"fmt"
	"log/slog"

	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type TablePage struct {
	app.Compo

	columns       []myui.TableColumn[characterRow]
	rows          []characterRow
	newCharacters uint
}

type characterRow struct {
	Name string
	Role string
	Crew string
}

func (c *TablePage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "TablePage: OnMount")

	c.columns = []myui.TableColumn[characterRow]{
		{
			Name: "Name",
			Value: func(row characterRow) any {
				return row.Name
			},
		},
		{
			Name: "Role",
			Value: func(row characterRow) any {
				return row.Role
			},
		},
		{
			Name: "Crew",
			Value: func(row characterRow) any {
				return row.Crew
			},
		},
	}
	c.rows = []characterRow{
		{
			Name: "Monkey D. Luffy",
			Role: "Captain",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Roronoa Zoro",
			Role: "Swordsman",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Nami",
			Role: "Navigator",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Usopp",
			Role: "Sniper",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Sanji",
			Role: "Cook",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Chopper",
			Role: "Doctor",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Nico Robin",
			Role: "Archaeologist",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Franky",
			Role: "Shipwright",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Brook",
			Role: "Musician",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Jinbe",
			Role: "Helmsman",
			Crew: "Straw Hat Pirates",
		},
		{
			Name: "Trafalgar D. Water Law",
			Role: "Captain",
			Crew: "Heart Pirates",
		},
		{
			Name: "Monkey D. Garp",
			Role: "Vice Admiral",
			Crew: "Navy",
		},
	}
}

func (c *TablePage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "TablePage: OnNav")
}

func (c *TablePage) Render() app.UI {
	clickFunction := func(ctx app.Context, row characterRow) {
		app.Window().Call("alert", "Clicked on "+row.Name)
	}

	addCharacterFunction := func(ctx app.Context) {
		c.newCharacters++
		c.rows = append(c.rows, characterRow{
			Name: fmt.Sprintf("New character %d", c.newCharacters),
			Role: fmt.Sprintf("New role %d", c.newCharacters),
			Crew: fmt.Sprintf("New crew %d", c.newCharacters),
		})

		ctx.Update()
	}

	return app.Div().
		Style("padding", "1em").
		Body(
			app.FieldSet().
				Body(
					app.Legend().Text("Simple"),
					myui.Table[characterRow]().
						Rows(c.rows).
						Columns(c.columns),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Actions"),
					app.Div().Body(
						myui.Table[characterRow]().
							Rows(c.rows).
							Columns(c.columns).
							Action(myui.TableAction{
								Name:     "Add character",
								Function: addCharacterFunction,
							}).
							RowAction(myui.RowAction[characterRow]{
								Name:     "Click",
								Function: clickFunction,
							}),
					),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Paginated"),
					myui.Table[characterRow]().
						PageSize(10).
						Rows(c.rows).
						Columns(c.columns),
				),
			app.FieldSet().
				Body(
					app.Legend().Text("Column Visibility"),
					myui.Table[characterRow]().
						PageSize(10).
						VisibleColumns([]string{"Name"}).
						Rows(c.rows).
						Columns(c.columns),
				),
		)
}
