package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Table[T any] struct {
	app.Compo

	columns []TableColumn[T]
	rows    []T
}

type TableColumn[T any] struct {
	Name  string
	To    func(row T) string
	Value func(row T) any
}

func NewTable[T any]() *Table[T] {
	table := Table[T]{}

	return &table
}

func (t *Table[T]) Rows(rows []T) *Table[T] {
	t.rows = rows

	return t
}

func (t *Table[T]) Columns(columns []TableColumn[T]) *Table[T] {
	t.columns = columns

	return t
}

func (t *Table[T]) Render() app.UI {
	return app.Table().
		Class("myui-table").
		Body(
			app.THead().
				Body(
					app.Tr().
						Body(
							app.Range(t.columns).Slice(func(i int) app.UI {
								return app.Th().
									Text(t.columns[i].Name)
							}),
						),
				),
			app.TBody().
				Body(
					app.Range(t.rows).Slice(func(i int) app.UI {
						row := t.rows[i]
						return app.Tr().
							Body(
								app.Range(t.columns).Slice(func(i int) app.UI {
									column := t.columns[i]
									return app.Td().
										Body(
											app.If(column.To != nil, func() app.UI {
												return app.A().
													Href(column.To(row)).
													Text(column.Value(row))
											}).Else(func() app.UI {
												return app.Span().Text(column.Value(row))
											}),
										)
								}),
							)
					}),
				),
		)
}
