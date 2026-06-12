package component

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type htmlMailingLabels struct {
	app.Compo

	IFormat        string
	IReturnAddress string

	addresses []string
}

func MailingLabels() *htmlMailingLabels {
	return &htmlMailingLabels{
		IFormat: "5164",
	}
}

func (c *htmlMailingLabels) Format(format string) *htmlMailingLabels {
	c.IFormat = format
	return c
}

func (c *htmlMailingLabels) ReturnAddress(returnAddress string) *htmlMailingLabels {
	c.IReturnAddress = returnAddress
	return c
}

func (c *htmlMailingLabels) Addresses(addresses []string) *htmlMailingLabels {
	c.addresses = addresses
	return c
}

func (c *htmlMailingLabels) Render() app.UI {
	slog.InfoContext(context.TODO(), "htmlMailingLabels: Render")

	itemsPerPage := 1
	pageWidth := "8.5in"
	pageHeight := "11in"
	pagePaddingTop := "0"
	pagePaddingBottom := "0"
	pagePaddingLeft := "0"
	pagePaddingRight := "0"
	pageGridTemplateColumns := "1fr"
	pageGridColumnGap := "0"
	labelWidth := "1in"
	labelHeight := "1in"
	labelPadding := "0"

	switch c.IFormat {
	case "5164":
		itemsPerPage = 6
		pageWidth = "8.5in"
		pageHeight = "11in"
		pagePaddingTop = "0.5in"
		pagePaddingBottom = "0.5in"
		pagePaddingLeft = "0.5in"
		pagePaddingRight = "0.5in"
		pageGridTemplateColumns = "4in 4in"
		pageGridColumnGap = "0.2in"

		labelWidth = "4in"
		labelHeight = "3.33in"
		labelPadding = "0.3in"
	}

	pages := slices.Collect(slices.Chunk(c.addresses, itemsPerPage))

	slog.InfoContext(context.TODO(), "htmlMailingLabels: Render", "addresses", len(c.addresses))
	slog.InfoContext(context.TODO(), "htmlMailingLabels: Render", "pages", len(pages))

	return app.Div().
		Class("mailing-labels").
		Body(
			app.Range(pages).Slice(func(i int) app.UI {
				page := pages[i]

				return app.Div().
					Class("mailing-labels__page").
					Style("width", pageWidth).
					Style("height", pageHeight).
					Style("padding-top", pagePaddingTop).
					Style("padding-bottom", pagePaddingBottom).
					Style("padding-left", pagePaddingLeft).
					Style("padding-right", pagePaddingRight).
					Style("box-sizing", "border-box").
					Style("background", "white").
					Style("page-break-after", "always").
					Style("display", "grid").
					Style("grid-template-columns", pageGridTemplateColumns).
					Style("grid-column-gap", pageGridColumnGap).
					Body(
						app.Range(page).Slice(func(i int) app.UI {
							address := page[i]

							createAddressBody := func(address string) []app.UI {
								addressLines := strings.Split(address, "\n")

								var addressBody []app.UI
								for _, line := range addressLines {
									if len(addressBody) > 0 {
										addressBody = append(addressBody, app.Br())
									}
									addressBody = append(addressBody, app.Text(line))
								}

								return addressBody
							}

							deliveryAddressBody := createAddressBody(address)
							returnAddressBody := createAddressBody(c.IReturnAddress)

							return app.Div().
								Class("mailing-labels__label").
								Style("width", labelWidth).
								Style("height", labelHeight).
								Style("padding", labelPadding).
								Style("box-sizing", "border-box").
								Style("overflow", "hidden").
								Style("display", "flex").
								Style("flex-direction", "column").
								Style("justify-content", "center").
								Body(
									app.Div().
										Class("mailing-labels__return-address").
										Style("font-size", "9pt").
										Style("margin-bottom", "0.4in").
										Style("text-transform", "uppercase").
										Body(returnAddressBody...),
									app.Div().
										Class("mailing-labels__delivery-address").
										Style("font-size", "14pt").
										Style("font-weight", "bold").
										Style("text-transform", "uppercase").
										Style("padding-left", "0.4in").
										Body(deliveryAddressBody...),
								)
						}),
					)
			}),
		)
}
