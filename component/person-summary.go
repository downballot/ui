package component

import (
	"fmt"
	"net/url"

	"github.com/downballot/downballot/downballotapi"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type personSummary struct {
	app.Compo

	IPerson *downballotapi.Person
}

var _ app.Composer = (*personSummary)(nil)

func PersonSummary() *personSummary {
	return &personSummary{}
}

func (c *personSummary) Person(person *downballotapi.Person) *personSummary {
	c.IPerson = person
	return c
}

func (c *personSummary) Render() app.UI {
	if c.IPerson == nil {
		return nil
	}
	if c.IPerson.Fields == nil {
		return nil
	}

	icon := "circle-question"
	switch c.IPerson.Fields["candidate.support"] {
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

	var registrationString string
	if c.IPerson.Fields["computed.age"] != "" {
		registrationString = c.IPerson.Fields["computed.age"] + " year-old"
	} else {
		registrationString = "Born in " + c.IPerson.Fields["birthday_year"]
	}
	registrationString += ", registered as " + c.IPerson.Fields["political_party"]

	developmentString := "Lives in "
	if c.IPerson.Fields["residential_address_development"] != "" {
		developmentString += c.IPerson.Fields["residential_address_development"] + " "
	}
	developmentString += "(" + c.IPerson.Fields["district_representative"] + ", " + c.IPerson.Fields["district_senate"] + ")"

	summaryItems := []app.UI{
		app.Div().
			Class("person-summary__header").
			Body(
				blazar.Icon().
					Icon(icon),
				app.Div().
					Text(c.IPerson.Fields["name"]),
			),
		app.Div().
			Class("person-summary__registration").
			Text(registrationString),
		app.Div().
			Class("person-summary__address").
			Body(
				app.A().
					Href(fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(c.IPerson.Fields["residential_address"]))).
					Target("_blank").
					Text(c.IPerson.Fields["residential_address"]),
			),
		app.If(developmentString != "", func() app.UI {
			return app.Div().
				Class("person-summary__address").
				Body(
					app.Span().
						Text(developmentString),
				)
		}),
		app.If(c.IPerson.Fields["phone_number"] != "", func() app.UI {
			return app.Div().
				Class("person-summary__phone").
				Text(c.IPerson.Fields["phone_number"])
		}),
		app.If(c.IPerson.Fields["email_address"] != "", func() app.UI {
			return app.Div().
				Class("person-summary__email").
				Text(c.IPerson.Fields["email_address"])
		}),
		app.Div().
			Class("person-summary-chips").
			Body(
				app.If(c.IPerson.Fields["candidate.connected"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("handshake"),
							app.Text("Connected"),
						)
				}),
				app.If(c.IPerson.Fields["candidate.support"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Support: " + c.IPerson.Fields["candidate.support"])
				}),
				app.If(c.IPerson.Fields["candidate.cat"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("cat"),
							app.Text("Cat"),
						)
				}),
				app.If(c.IPerson.Fields["candidate.dog"] == "true", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Body(
							blazar.Icon().
								Icon("dog"),
							app.Text("Dog"),
						)
				}),
				app.If(c.IPerson.Fields["candidate.date_called"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Called on " + c.IPerson.Fields["candidate.date_called"])
				}),
				app.If(c.IPerson.Fields["candidate.date_canvassed"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Canvassed on " + c.IPerson.Fields["candidate.date_canvassed"])
				}),
				app.If(c.IPerson.Fields["candidate.date_texted"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("Texted on " + c.IPerson.Fields["candidate.date_texted"])
				}),
				app.If(c.IPerson.Fields["candidate.date_whatsapped"] != "", func() app.UI {
					return app.Div().
						Class("person-summary-chip").
						Text("WhatsApped on " + c.IPerson.Fields["candidate.date_whatsapped"])
				}),
			),
		app.If(c.IPerson.Fields["candidate.notes"] != "", func() app.UI {
			return app.Div().
				Class("person-summary__notes").
				Body(
					blazar.Icon().
						Solid(false).
						Icon("note-sticky"),
					app.Text(c.IPerson.Fields["candidate.notes"]),
				)
		}),
	}

	return app.Div().
		Body(
			app.Div().
				Class("person-summary").
				Body(summaryItems...),
		)
}
