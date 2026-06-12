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

	ITitle                  string
	IColumns                []TableColumn[T]
	IVisibleColumnNames     []string
	IBindVisibleColumnNames *[]string
	IRows                   []T
	pageSize                uint
	pageIndex               uint
	IActions                []TableAction
	IRowActions             []RowAction[T]
	IEmptyMessage           string
	BindValue               *TableBinding[T]
}

type TableBinding[T any] struct {
	PageSize           uint
	PageIndex          uint
	VisibleColumnNames []string
	Rows               []T
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

func (t *MyUITable[T]) Bind(bindValue *TableBinding[T]) *MyUITable[T] {
	t.BindValue = bindValue
	if t.BindValue != nil {
		t.pageIndex = t.BindValue.PageIndex
		t.pageSize = t.BindValue.PageSize
	}
	return t
}

func (t *MyUITable[T]) Title(title string) *MyUITable[T] {
	t.ITitle = title
	return t
}

func (t *MyUITable[T]) Rows(rows []T) *MyUITable[T] {
	t.IRows = rows
	if t.BindValue != nil {
		t.BindValue.Rows = rows
	}
	return t
}

func (t *MyUITable[T]) Columns(columns []TableColumn[T]) *MyUITable[T] {
	t.IColumns = columns
	return t
}

func (t *MyUITable[T]) VisibleColumns(visibleColumnNames []string) *MyUITable[T] {
	t.IVisibleColumnNames = visibleColumnNames
	if t.BindValue != nil {
		t.BindValue.VisibleColumnNames = visibleColumnNames
	}
	return t
}

func (t *MyUITable[T]) BindVisibleColumns(visibleColumnNames *[]string) *MyUITable[T] {
	t.IBindVisibleColumnNames = visibleColumnNames
	t.IVisibleColumnNames = *visibleColumnNames
	return t
}

func (t *MyUITable[T]) EmptyMessage(emptyMessage string) *MyUITable[T] {
	t.IEmptyMessage = emptyMessage
	return t
}

func (t *MyUITable[T]) PageIndex(pageIndex uint) *MyUITable[T] {
	t.pageIndex = pageIndex
	if t.BindValue != nil {
		t.BindValue.PageIndex = pageIndex
	}
	return t
}

func (t *MyUITable[T]) PageSize(pageSize uint) *MyUITable[T] {
	t.pageSize = pageSize
	if t.BindValue != nil {
		t.BindValue.PageSize = pageSize
	}
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

func (t *MyUITable[T]) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "MyUITable: OnUpdate", "IPageIndex", t.pageIndex, "IPageSize", t.pageSize, "rows", len(t.IRows))

	if t.BindValue != nil {
		t.pageIndex = t.BindValue.PageIndex
		t.pageSize = t.BindValue.PageSize
	}

	totalPages := uint(1)
	if t.pageSize > 0 {
		totalPages = uint(uint(len(t.IRows)) / t.pageSize)
		if uint(len(t.IRows))%t.pageSize > 0 {
			totalPages++
		}
	}

	if totalPages > 0 {
		if t.pageIndex >= totalPages {
			t.pageIndex = totalPages - 1
			if t.BindValue != nil {
				t.BindValue.PageIndex = t.pageIndex
			}
			slog.InfoContext(ctx.Context, "MyUITable: OnUpdate: Setting IPageIndex.", "IPageIndex", t.pageIndex, "totalPages", totalPages)
		}
	}
}

func (t *MyUITable[T]) Render() app.UI {
	if t.BindValue != nil {
		t.pageIndex = t.BindValue.PageIndex
	}
	if t.BindValue != nil {
		t.pageSize = t.BindValue.PageSize
	}
	slog.InfoContext(context.TODO(), "MyUITable: Render", "IPageIndex", t.pageIndex, "IPageSize", t.pageSize, "rows", len(t.IRows))
	slog.InfoContext(context.TODO(), "MyUITable: Render", "IVisibleColumnNames", t.IVisibleColumnNames)
	slog.InfoContext(context.TODO(), "MyUITable: Render", "IBindVisibleColumnNames", t.IBindVisibleColumnNames)
	if t.BindValue == nil {
		slog.InfoContext(context.TODO(), "MyUITable: Render: BindValue is nil")
	} else {
		slog.InfoContext(context.TODO(), "MyUITable: Render", "BindValue", *t.BindValue)
	}

	visibleColumns := []TableColumn[T]{}
	for _, column := range t.IColumns {
		if len(t.IVisibleColumnNames) == 0 || slices.Contains(t.IVisibleColumnNames, column.Name) {
			visibleColumns = append(visibleColumns, column)
		}
	}
	slog.InfoContext(context.TODO(), "MyUITable: Render", "visibleColumns", visibleColumns)

	rows := t.IRows
	paginated := t.pageSize > 0
	totalPages := uint(1)
	if t.pageSize > 0 && uint(len(t.IRows)) > t.pageSize {
		totalPages = uint(uint(len(t.IRows)) / t.pageSize)
		if uint(len(t.IRows))%t.pageSize > 0 {
			totalPages++
		}

		pages := slices.Collect(slices.Chunk(t.IRows, int(t.pageSize)))
		if t.pageIndex >= uint(len(pages)) {
			t.pageIndex = uint(len(pages)) - 1
			if t.BindValue != nil {
				t.BindValue.PageIndex = t.pageIndex
			}
		}
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

	tableMenuItems := []app.UI{}
	if t.IBindVisibleColumnNames != nil {
		tableMenuItems = append(tableMenuItems, Item().
			Icon("list").
			Label("Select columns...").
			On("click", func(ctx app.Context, e app.Event) {
				slog.InfoContext(ctx.Context, "MyUITable: Render: item clicked")

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
									Bind(&popoverSelectedColumns),
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
										if t.BindValue != nil {
											t.BindValue.VisibleColumnNames = t.IVisibleColumnNames
										}
										slog.InfoContext(ctx.Context, "MyUITable: Render: Apply", "visibleColumnNames", t.IVisibleColumnNames)
										ctx.Update()
									}),
							),
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
							app.If(len(rows) == 0, func() app.UI {
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
							app.Range(rows).Slice(func(i int) app.UI {
								row := rows[i]
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
								t.pageIndex--
								if t.BindValue != nil {
									t.BindValue.PageIndex = t.pageIndex
								}
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
								if t.BindValue != nil {
									t.BindValue.PageIndex = uint(index)
								}
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
								slog.InfoContext(ctx.Context, "MyUITable: Render: Next", "IPageIndex", t.pageIndex)
								if t.BindValue != nil {
									t.BindValue.PageIndex = t.pageIndex
									slog.InfoContext(ctx.Context, "MyUITable: Render: Next", "BindValue", *t.BindValue)
								}
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
								if t.BindValue != nil {
									t.BindValue.PageSize = uint(pageSize)
								}
								slog.InfoContext(ctx.Context, "MyUITable: Setting IPageSize via select.", "IPageSize", t.pageSize)
								ctx.Update()
							}),
					)
			}),
		)
}
