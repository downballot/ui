package myui

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type MyUITable[T any] struct {
	app.Compo

	title                   string
	columns                 []TableColumn[T]
	IVisibleColumnNames     []string
	IBindVisibleColumnNames *[]string
	rows                    []T
	pageSize                uint
	pageIndex               uint
	actions                 []TableAction
	rowActions              []RowAction[T]
}

type TableAction struct {
	Name     string
	Icon     string
	To       func() string
	Function func(ctx app.Context)
}

type RowAction[T any] struct {
	Name     string
	Icon     string
	To       func(row T) string
	Function func(ctx app.Context, row T)
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

func (t *MyUITable[T]) Title(title string) *MyUITable[T] {
	t.title = title
	return t
}

func (t *MyUITable[T]) Rows(rows []T) *MyUITable[T] {
	t.rows = rows
	return t
}

func (t *MyUITable[T]) Columns(columns []TableColumn[T]) *MyUITable[T] {
	t.columns = columns
	return t
}

func (t *MyUITable[T]) VisibleColumns(visibleColumnNames []string) *MyUITable[T] {
	t.IVisibleColumnNames = visibleColumnNames
	return t
}

func (t *MyUITable[T]) BindVisibleColumns(visibleColumnNames *[]string) *MyUITable[T] {
	t.IBindVisibleColumnNames = visibleColumnNames
	t.IVisibleColumnNames = *visibleColumnNames
	return t
}

func (t *MyUITable[T]) PageIndex(pageIndex uint) *MyUITable[T] {
	t.pageIndex = pageIndex
	return t
}

func (t *MyUITable[T]) PageSize(pageSize uint) *MyUITable[T] {
	t.pageSize = pageSize
	return t
}

func (t *MyUITable[T]) Action(actions ...TableAction) *MyUITable[T] {
	t.actions = append(t.actions, actions...)
	return t
}

func (t *MyUITable[T]) RowAction(rowActions ...RowAction[T]) *MyUITable[T] {
	t.rowActions = append(t.rowActions, rowActions...)
	return t
}

func (t *MyUITable[T]) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUITable: Render", "pageIndex", t.pageIndex, "pageSize", t.pageSize, "rows", len(t.rows))

	visibleColumns := []TableColumn[T]{}
	for _, column := range t.columns {
		if len(t.IVisibleColumnNames) == 0 || slices.Contains(t.IVisibleColumnNames, column.Name) {
			visibleColumns = append(visibleColumns, column)
		}
	}
	slog.InfoContext(context.TODO(), "MyUITable: Render", "visibleColumns", visibleColumns)

	rows := t.rows
	paginated := t.pageSize > 0
	totalPages := uint(1)
	if t.pageSize > 0 && uint(len(t.rows)) > t.pageSize {
		totalPages = uint(uint(len(t.rows)) / t.pageSize)
		if uint(len(t.rows))%t.pageSize > 0 {
			totalPages++
		}

		pages := slices.Collect(slices.Chunk(t.rows, int(t.pageSize)))
		rows = pages[t.pageIndex]
	}
	pageIndexes := []uint{}
	for i := uint(0); i < totalPages; i++ {
		pageIndexes = append(pageIndexes, i)
	}

	pageSizes := []uint{1, 10, 50, 100, 500, 10000, 100000, 1000000}

	popoverSelectedColumns := []string{}
	for _, column := range visibleColumns {
		popoverSelectedColumns = append(popoverSelectedColumns, column.Name)
	}

	return app.Div().
		Class("myui-table").
		Body(
			app.Div().
				Class("myui-table-header").
				Body(
					app.Div().
						Class("myui-table-title").
						Text(t.title),
					app.Span().Style("flex", "1"),
					Button().
						Round(true).
						Flat(true).
						Icon("ellipsis-vertical").
						On("click", func(ctx app.Context, e app.Event) {
							thisElement := ctx.JSSrc()
							slog.InfoContext(ctx.Context, "MyUITable: Render", "thisElement", thisElement.Get("className").String())
							parentElement := ctx.JSSrc().Call("closest", ".myui-table-header")
							slog.InfoContext(ctx.Context, "MyUITable: Render", "parentElement", parentElement.Get("className").String())
							if parentElement.IsNull() {
								return
							}
							popoverElement := parentElement.Call("querySelector", "[popover][data-popover-name='table-menu']")
							if popoverElement.IsNull() {
								return
							}
							slog.InfoContext(ctx.Context, "MyUITable: Render", "popoverElement", popoverElement)
							options := app.ValueOf(map[string]any{})
							options.Set("source", thisElement)

							popoverElement.Call("togglePopover", options)
						}),
					app.Div().
						Attr("popover", "auto").
						DataSet("popover-name", "table-menu").
						Body(
							Item().
								Icon("list").
								Name("Select columns...").
								On("click", func(ctx app.Context, e app.Event) {
									slog.InfoContext(ctx.Context, "MyUITable: Render: item clicked")

									thisElement := ctx.JSSrc()
									slog.InfoContext(ctx.Context, "MyUITable: Render", "thisElement", thisElement.Get("className").String())
									parentElement := ctx.JSSrc().Call("closest", ".myui-table-header")
									slog.InfoContext(ctx.Context, "MyUITable: Render", "parentElement", parentElement.Get("className").String())
									if parentElement.IsNull() {
										return
									}
									popoverElement := parentElement.Call("querySelector", "[popover][data-popover-name='table-columns-menu']")
									if popoverElement.IsNull() {
										return
									}
									slog.InfoContext(ctx.Context, "MyUITable: Render", "popoverElement", popoverElement)
									options := app.ValueOf(map[string]any{})
									options.Set("source", thisElement)

									popoverElement.Call("togglePopover", options)
								}),
						),
					app.Div().
						Attr("popover", "auto").
						DataSet("popover-name", "table-columns-menu").
						Body(
							Multiselect().
								Label("Columns").
								AllowedValue(func() []SelectOption {
									columns := []SelectOption{}
									for _, column := range t.columns {
										columns = append(columns, SelectOption{
											Label: column.Name,
											Value: column.Name,
										})
									}
									return columns
								}()...).
								SelectedValue(popoverSelectedColumns...).
								On("change", SelectedValuesTo(&popoverSelectedColumns)),
							Button().
								Label("Apply").
								On("click", func(ctx app.Context, e app.Event) {
									slog.InfoContext(ctx.Context, "MyUITable: Render: item clicked")
									slog.InfoContext(ctx.Context, "MyUITable: Render: Apply", "popoverSelectedColumns", popoverSelectedColumns)

									popoverElement := ctx.JSSrc().Call("closest", "[popover]")
									slog.InfoContext(ctx.Context, "MyUITable: Render", "popoverElement", popoverElement.Get("className").String())
									if popoverElement.IsNull() {
										return
									}
									popoverElement.Call("hidePopover")

									t.IVisibleColumnNames = popoverSelectedColumns
									if t.IBindVisibleColumnNames != nil {
										*t.IBindVisibleColumnNames = t.IVisibleColumnNames
									}
									slog.InfoContext(ctx.Context, "MyUITable: Render: Apply", "visibleColumnNames", t.IVisibleColumnNames)
									ctx.Update()
								}),
						),
				),
			app.If(len(t.actions) > 0, func() app.UI {
				return app.Div().
					Class("myui-table-actions").
					Body(
						app.Range(t.actions).Slice(func(i int) app.UI {
							action := t.actions[i]
							button := Button().
								Label(action.Name).
								Icon(action.Icon).
								On("click", func(ctx app.Context, e app.Event) {
									if action.Function == nil {
										ctx.PreventUpdate()
										return
									}
									action.Function(ctx)
								})
							if action.To != nil {
								button.To(action.To())
							}
							return button
						}),
					)
			}),
			app.Table().
				Body(
					app.THead().
						Body(
							app.Tr().
								Body(
									app.Range(visibleColumns).Slice(func(i int) app.UI {
										column := visibleColumns[i]
										return app.Th().
											Text(column.Name)
									}),
									app.If(len(t.rowActions) > 0, func() app.UI {
										return app.Th().
											Text("Actions")
									}),
								),
						),
					app.TBody().
						Body(
							app.Range(rows).Slice(func(i int) app.UI {
								row := rows[i]
								return app.Tr().
									Body(
										app.Range(visibleColumns).Slice(func(i int) app.UI {
											column := visibleColumns[i]
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
										app.If(len(t.rowActions) > 0, func() app.UI {
											return app.Td().
												Body(
													app.Range(t.rowActions).Slice(func(i int) app.UI {
														rowAction := t.rowActions[i]
														button := Button().
															Label(rowAction.Name).
															Icon(rowAction.Icon)
														if rowAction.To != nil {
															button.To(rowAction.To(row))
														}
														if rowAction.Function != nil {
															button.On("click", func(ctx app.Context, e app.Event) {
																rowAction.Function(ctx, row)
															})
														}
														return button
													}),
												)
										}),
									)
							}),
						),
				),
			app.If(paginated, func() app.UI {
				return app.Div().
					Class("myui-table-pagination").
					Body(
						Button().
							Label("Previous").
							Disabled(t.pageIndex < 1).
							On("click", func(ctx app.Context, e app.Event) {
								t.pageIndex--
								ctx.Update()
							}),
						app.Span().
							Style("display", "flex").
							Style("align-items", "center").
							Text("Page"),
						app.Select().
							Body(
								app.Range(pageIndexes).Slice(func(i int) app.UI {
									index := pageIndexes[i]
									return app.Option().
										Value(index).
										Selected(index == t.pageIndex).
										Text(fmt.Sprintf("%d", index+1)).Selected(index == t.pageIndex)
								}),
							).
							OnChange(func(ctx app.Context, e app.Event) {
								v := e.Get("target").Get("value").String()
								index, err := strconv.ParseUint(v, 10, 64)
								if err != nil {
									return
								}
								t.pageIndex = uint(index)
								ctx.Update()
							}),
						app.Span().
							Style("display", "flex").
							Style("align-items", "center").
							Text(fmt.Sprintf("/%d", totalPages)),
						Button().
							Label("Next").
							Disabled(t.pageIndex >= totalPages-1).
							On("click", func(ctx app.Context, e app.Event) {
								t.pageIndex++
								ctx.Update()
							}),
						app.Span().
							Style("flex-grow", "1"),
						app.Span().
							Style("display", "flex").
							Style("align-items", "center").
							Text("Page size:"),
						app.Select().
							Body(
								app.Range(pageSizes).Slice(func(i int) app.UI {
									pageSize := pageSizes[i]
									return app.Option().
										Value(pageSize).
										Selected(pageSize == t.pageSize).
										Text(fmt.Sprintf("%d", pageSize)).Selected(pageSize == t.pageSize)
								}),
							).
							OnChange(func(ctx app.Context, e app.Event) {
								v := e.Get("target").Get("value").String()
								pageSize, err := strconv.ParseUint(v, 10, 64)
								if err != nil {
									return
								}
								t.pageSize = uint(pageSize)
								ctx.Update()
							}),
					)
			}),
		)
}
