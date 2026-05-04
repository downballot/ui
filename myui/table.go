package myui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type MyUITable[T any] struct {
	app.Compo

	columns []TableColumn[T]
	rows    []T
}

type TableColumn[T any] struct {
	Name  string
	To    func(row T) string
	Value func(row T) any
}

func Table[T any]() *MyUITable[T] {
	table := MyUITable[T]{}

	return &table
}

func (t *MyUITable[T]) Rows(rows []T) *MyUITable[T] {
	t.rows = rows

	return t
}

func (t *MyUITable[T]) Columns(columns []TableColumn[T]) *MyUITable[T] {
	t.columns = columns

	return t
}

func (t *MyUITable[T]) Render() app.UI {
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
