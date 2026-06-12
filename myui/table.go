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

	ITitle        string
	IColumns      []TableColumn[T]
	IRows         []T
	IActions      []TableAction
	IRowActions   []RowAction[T]
	IEmptyMessage string

	visibleColumnNames         []string // This is the list of columns that are currently visible in the table.
	popoverSelectedColumnNames []string // This is the list of columns that are currently selected in the popover.
	pageSize                   uint     // This is the number of rows to display per page.
	pageIndex                  uint     // This is the index of the current page.
}

type TableAction struct {
	Name     string
	Icon     string
	To       string
	Function func(ctx app.Context)
	Disabled bool
}

type RowAction[T any] struct {
	Name     string
	Icon     string
	To       func(row T) string
	Function func(ctx app.Context, row T)
	Disabled bool
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
	t.ITitle = title
	return t
}

func (t *MyUITable[T]) Rows(rows []T) *MyUITable[T] {
	t.IRows = rows
	return t
}

func (t *MyUITable[T]) Columns(columns []TableColumn[T]) *MyUITable[T] {
	t.IColumns = columns

	if len(t.popoverSelectedColumnNames) == 0 {
		t.popoverSelectedColumnNames = make([]string, 0, len(t.IColumns))
		for _, column := range t.IColumns {
			t.popoverSelectedColumnNames = append(t.popoverSelectedColumnNames, column.Name)
		}
	}
	slog.InfoContext(context.TODO(), "MyUITable: Columns", "self", fmt.Sprintf("%p", t), "popoverSelectedColumnNames", t.popoverSelectedColumnNames)
	return t
}

func (t *MyUITable[T]) VisibleColumns(visibleColumnNames []string) *MyUITable[T] {
	t.visibleColumnNames = visibleColumnNames

	if len(t.visibleColumnNames) > 0 {
		t.popoverSelectedColumnNames = make([]string, len(t.visibleColumnNames))
		copy(t.popoverSelectedColumnNames, t.visibleColumnNames)
	}
	slog.InfoContext(context.TODO(), "MyUITable: VisibleColumns", "self", fmt.Sprintf("%p", t), "popoverSelectedColumnNames", t.popoverSelectedColumnNames)
	return t
}

func (t *MyUITable[T]) EmptyMessage(emptyMessage string) *MyUITable[T] {
	t.IEmptyMessage = emptyMessage
	return t
}

func (t *MyUITable[T]) PageIndex(pageIndex uint) *MyUITable[T] {
	t.setPageIndex(pageIndex)
	return t
}

func (t *MyUITable[T]) PageSize(pageSize uint) *MyUITable[T] {
	t.setPageSize(pageSize)
	return t
}

func (t *MyUITable[T]) Action(actions ...TableAction) *MyUITable[T] {
	t.IActions = actions
	return t
}

func (t *MyUITable[T]) RowAction(rowActions ...RowAction[T]) *MyUITable[T] {
	t.IRowActions = rowActions
	return t
}

func (t *MyUITable[T]) totalPages() uint {
	totalPages := uint(1)
	if t.pageSize > 0 {
		totalPages = uint(uint(len(t.IRows)) / t.pageSize)
		if uint(len(t.IRows))%t.pageSize > 0 {
			totalPages++
		}
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return totalPages
}

func (t *MyUITable[T]) previousPage() {
	if t.pageIndex > 0 {
		t.pageIndex--
	}
}

func (t *MyUITable[T]) nextPage() {
	t.pageIndex++

	totalPages := t.totalPages()
	if t.pageIndex >= totalPages {
		t.pageIndex = totalPages - 1
	}
}

func (t *MyUITable[T]) setPageIndex(pageIndex uint) {
	t.pageIndex = pageIndex

	totalPages := t.totalPages()
	if t.pageIndex >= totalPages {
		t.pageIndex = totalPages - 1
	}
}

func (t *MyUITable[T]) setPageSize(pageSize uint) {
	if pageSize > 0 {
		t.pageSize = pageSize

		totalPages := t.totalPages()
		if t.pageIndex >= totalPages {
			t.pageIndex = totalPages - 1
		}
	}
}

func (t *MyUITable[T]) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUITable: OnUpdate", "self", fmt.Sprintf("%p", t), "pageIndex", t.pageIndex, "pageSize", t.pageSize, "rows", len(t.IRows))
	slog.InfoContext(ctx.Context, "MyUITable: OnUpdate", "self", fmt.Sprintf("%p", t), "visibleColumnNames", len(t.visibleColumnNames), "popoverSelectedColumnNames", len(t.popoverSelectedColumnNames))

	if len(t.popoverSelectedColumnNames) == 0 {
		if len(t.visibleColumnNames) == 0 {
			t.popoverSelectedColumnNames = make([]string, 0, len(t.IColumns))
			for _, column := range t.IColumns {
				t.popoverSelectedColumnNames = append(t.popoverSelectedColumnNames, column.Name)
			}
		} else {
			t.popoverSelectedColumnNames = make([]string, len(t.visibleColumnNames))
			copy(t.popoverSelectedColumnNames, t.visibleColumnNames)
		}
	}
	slog.InfoContext(ctx.Context, "MyUITable: OnUpdate", "self", fmt.Sprintf("%p", t), "popoverSelectedColumnNames", t.popoverSelectedColumnNames)
}

func (t *MyUITable[T]) Render() app.UI {
	slog.InfoContext(context.TODO(), "MyUITable: Render", "self", fmt.Sprintf("%p", t), "pageIndex", t.pageIndex, "pageSize", t.pageSize, "rows", len(t.IRows))
	slog.InfoContext(context.TODO(), "MyUITable: Render", "self", fmt.Sprintf("%p", t), "visibleColumnNames", t.visibleColumnNames)
	slog.InfoContext(context.TODO(), "MyUITable: Render", "self", fmt.Sprintf("%p", t), "popoverSelectedColumnNames", t.popoverSelectedColumnNames)

	visibleColumns := []TableColumn[T]{}
	for _, column := range t.IColumns {
		if len(t.visibleColumnNames) == 0 || slices.Contains(t.visibleColumnNames, column.Name) {
			visibleColumns = append(visibleColumns, column)
		}
	}
	slog.InfoContext(context.TODO(), "MyUITable: Render", "self", fmt.Sprintf("%p", t), "visibleColumns", visibleColumns)

	rowsToRender := t.IRows
	paginated := t.pageSize > 0
	totalPages := t.totalPages()
	if t.pageSize > 0 && uint(len(t.IRows)) > t.pageSize {
		pages := slices.Collect(slices.Chunk(t.IRows, int(t.pageSize)))
		if t.pageIndex >= uint(len(pages)) {
			t.pageIndex = uint(len(pages)) - 1
		}
		rowsToRender = pages[t.pageIndex]
	}

	// Build the list of page indexes (so we can render a dropdown).
	pageIndexes := []uint{}
	for i := uint(0); i < totalPages; i++ {
		pageIndexes = append(pageIndexes, i)
	}

	// This is a static list of page sizes that we will render in a dropdown.
	pageSizes := []uint{1, 10, 50, 100, 500, 10000, 100000, 1000000}

	tableMenuItems := []app.UI{}
	{
		tableMenuItems = append(tableMenuItems,
			Item().
				Icon("list").
				Label("Select columns...").
				On("click", func(ctx app.Context, e app.Event) {
					slog.InfoContext(ctx.Context, "MyUITable: Render: item clicked")

					ctx.PreventUpdate()

					thisElement := ctx.JSSrc()
					slog.InfoContext(ctx.Context, "MyUITable: Render", "thisElement", thisElement.Get("className").String())
					parentElement := ctx.JSSrc().Call("closest", ".myui-table__header")
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
		)
	}

	emptyMessage := t.IEmptyMessage
	if emptyMessage == "" {
		emptyMessage = "No results found"
	}

	var visibleActions []TableAction
	for _, action := range t.IActions {
		if action.Disabled {
			continue
		}
		visibleActions = append(visibleActions, action)
	}

	var visibleRowActions []RowAction[T]
	for _, rowAction := range t.IRowActions {
		if rowAction.Disabled {
			continue
		}
		visibleRowActions = append(visibleRowActions, rowAction)
	}

	return app.Div().
		Class("myui-table").
		Body(
			app.If(t.ITitle != "" || len(tableMenuItems) > 0, func() app.UI {
				return app.Div().
					Class("myui-table__header").
					Body(
						app.Div().
							Class("myui-table__title").
							Text(t.ITitle),
						app.Span().Style("flex", "1"),
						app.If(len(tableMenuItems) > 0, func() app.UI {
							return Button().
								Round(true).
								Flat(true).
								Icon("ellipsis-vertical").
								On("click", func(ctx app.Context, e app.Event) {
									thisElement := ctx.JSSrc()
									slog.InfoContext(ctx.Context, "MyUITable: Render", "thisElement", thisElement.Get("className").String())
									parentElement := ctx.JSSrc().Call("closest", ".myui-table__header")
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
								})
						}),
						app.If(len(tableMenuItems) > 0, func() app.UI {
							return app.Div().
								Attr("popover", "auto").
								DataSet("popover-name", "table-menu").
								Body(
									tableMenuItems...,
								)
						}),
						app.Div().
							Attr("popover", "auto").
							DataSet("popover-name", "table-columns-menu").
							Body(
								Multiselect().
									Label("Columns").
									AllowedValue(func() []SelectOption {
										columns := []SelectOption{}
										for _, column := range t.IColumns {
											columns = append(columns, SelectOption{
												Label: column.Name,
												Value: column.Name,
											})
										}
										return columns
									}()...).
									Bind(&t.popoverSelectedColumnNames),
								Button().
									Label("Apply").
									On("click", func(ctx app.Context, e app.Event) {
										slog.InfoContext(ctx.Context, "MyUITable: table-columns-menu: Apply", "popoverSelectedColumnNames", t.popoverSelectedColumnNames)

										popoverElement := ctx.JSSrc().Call("closest", "[popover]")
										slog.InfoContext(ctx.Context, "MyUITable: table-columns-menu", "popoverElement", popoverElement.Get("className").String())
										if popoverElement.IsNull() {
											return
										}
										popoverElement.Call("hidePopover")

										t.visibleColumnNames = make([]string, len(t.popoverSelectedColumnNames))
										copy(t.visibleColumnNames, t.popoverSelectedColumnNames)

										slog.InfoContext(ctx.Context, "MyUITable: table-columns-menu: Apply", "visibleColumnNames", t.visibleColumnNames)
										ctx.Update()
									}),
							).
							On("toggle", func(ctx app.Context, e app.Event) {
								newState := e.Get("newState").String()
								slog.InfoContext(ctx.Context, "MyUITable: table-columns-menu: on toggle", "newState", newState)
								if newState != "closed" {
									return
								}

								parentElement := ctx.JSSrc().Call("closest", ".myui-table__header")
								slog.InfoContext(ctx.Context, "MyUITable: Render", "parentElement", parentElement.Get("className").String())
								if parentElement.IsNull() {
									return
								}

								popoverElement := parentElement.Call("querySelector", "[popover][data-popover-name='table-menu']")
								if popoverElement.IsNull() {
									return
								}

								popoverElement.Call("hidePopover")
							}),
					)
			}),
			app.If(len(visibleActions) > 0, func() app.UI {
				return app.Div().
					Class("myui-table__actions").
					Body(
						app.Range(visibleActions).Slice(func(i int) app.UI {
							action := visibleActions[i]

							button := Button().
								Label(action.Name).
								Icon(action.Icon).
								To(action.To).
								On("click", func(ctx app.Context, e app.Event) {
									if action.Function == nil {
										ctx.PreventUpdate()
										return
									}
									action.Function(ctx)
								})
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
									app.If(len(visibleRowActions) > 0, func() app.UI {
										return app.Th().
											Text("Actions")
									}),
								),
						),
					app.TBody().
						Body(
							app.If(len(rowsToRender) == 0, func() app.UI {
								return app.Tr().
									Body(
										app.Td().
											ColSpan(len(visibleColumns) + 1 /* +1 for the actions column */).
											Body(
												app.Div().
													Class("myui-table__empty-message").
													Text(emptyMessage),
											),
									)
							}),
							app.Range(rowsToRender).Slice(func(i int) app.UI {
								row := rowsToRender[i]
								return app.Tr().
									Body(
										app.Range(visibleColumns).Slice(func(i int) app.UI {
											column := visibleColumns[i]
											return app.Td().
												Body(
													app.If(column.Value != nil, func() app.UI {
														value := column.Value(row)
														valueAsUI, valueIsUI := value.(app.UI)

														return app.If(column.To != nil, func() app.UI {
															element := app.A().
																Href(column.To(row))
															if valueIsUI {
																element.Body(valueAsUI)
															} else {
																element.Text(value)
															}
															return element
														}).Else(func() app.UI {
															element := app.Span()
															if valueIsUI {
																element.Body(valueAsUI)
															} else {
																element.Text(value)
															}
															return element
														})
													}),
												)
										}),
										app.If(len(visibleRowActions) > 0, func() app.UI {
											return app.Td().
												Body(
													app.Range(visibleRowActions).Slice(func(i int) app.UI {
														rowAction := visibleRowActions[i]
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
					Class("myui-table__pagination").
					Body(
						Button().
							Label("Previous").
							Disabled(t.pageIndex < 1).
							On("click", func(ctx app.Context, e app.Event) {
								t.previousPage()
								ctx.Update()
							}),
						app.Span().
							Style("display", "flex").
							Style("align-items", "center").
							Text("Page"),
						app.Select().
							Disabled(totalPages <= 1).
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

								t.setPageIndex(uint(index))
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
								t.nextPage()
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

								slog.InfoContext(ctx.Context, "MyUITable: Setting pageSize via select.", "pageSize", pageSize)
								t.setPageSize(uint(pageSize))
								ctx.Update()
							}),
					)
			}),
		)
}
