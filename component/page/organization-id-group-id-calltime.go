package page

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/downballot/ui/googlemap"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/deref"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDCalltimePage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	groupID        string
	group          *downballotapi.Group
	permissionSet  permissionset.PermissionSet

	Filter  string
	Persons []*downballotapi.Person
	Filters []*downballotapi.Filter

	PersonsTableColumns        []blazar.TableColumn[*downballotapi.Person]
	PersonsTableVisibleColumns []string

	filterOpen bool
}

var _ app.Navigator = (*OrganizationIDGroupIDCalltimePage)(nil)
var _ app.Mounter = (*OrganizationIDGroupIDCalltimePage)(nil)

func (c *OrganizationIDGroupIDCalltimePage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnMount")

	c.EmbeddedPage.OnMount(ctx)
}

func (c *OrganizationIDGroupIDCalltimePage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav", "GroupID", c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.groupID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	c.filterOpen = false

	var persistFilter string
	ctx.GetState("persist-organization-id-group-id-person-page-filter", &persistFilter)
	if persistFilter != "" {
		c.Filter = persistFilter
	}

	var persistFilterOpen *bool
	ctx.GetState("persist-organization-id-group-id-person-page-filter-open", &persistFilterOpen)
	if persistFilterOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav: Persist filter open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav: Persist filter open is not nil.", "persistFilterOpen", deref.String(persistFilterOpen))
		c.filterOpen = *persistFilterOpen
	}

	if value := ctx.Page().URL().Query().Get("filter"); value != "" {
		c.Filter = value
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting group", "group", output.Group)
				c.group = output.Group
			})
		})
		wg.Go(func() {
			var output downballotapi.ListFiltersResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/filter", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get filters", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting filters", "len(filters)", len(output.Filters))
				c.Filters = output.Filters
			})
		})
		wg.Go(func() {
			var output downballotapi.ListPersonFieldsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person-field", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
				return
			}

			var possibleFields []string
			for _, personField := range output.PersonFields {
				possibleFields = append(possibleFields, personField.Name)
			}
			slices.Sort(possibleFields)

			fieldNameMap := map[string]bool{}
			for _, personField := range output.PersonFields {
				fieldNameMap[personField.Name] = true
			}

			c.PersonsTableColumns = []blazar.TableColumn[*downballotapi.Person]{
				{
					Name: "Voter ID",
					Value: func(row *downballotapi.Person) any {
						return row.VoterID
					},
					To: func(row *downballotapi.Person) string {
						return fmt.Sprintf("/organization/%s/person/%s", c.organizationID, row.VoterID)
					},
				},
			}

			for _, name := range possibleFields {
				displayName := name
				for _, personField := range output.PersonFields {
					if personField.Name == name {
						if personField.DisplayName != "" {
							displayName = personField.DisplayName
						}
						break
					}
				}
				c.PersonsTableColumns = append(c.PersonsTableColumns, blazar.TableColumn[*downballotapi.Person]{
					Name:        name,
					DisplayName: displayName,
					Value: func(row *downballotapi.Person) any {
						if name == "computed.likely" {
							if row.Fields[name] == "true" {
								return blazar.Icon().Icon("star")
							}
							return ""
						}
						return row.Fields[name]
					},
				})
			}

			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: OnNav: Async", "len(PersonsTableColumns)", len(c.PersonsTableColumns))
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Loading complete.  Searching for persons.")
			c.search(ctx)
		})
	})
}

func (c *OrganizationIDGroupIDCalltimePage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDCalltimePage: Render")

	markers := []googlemap.Marker{}
	center := googlemap.Coordinate{
		Latitude:  39.713171422509426,
		Longitude: -75.75937795659787,
	}
	{
		coordinatesField := "coordinates" // TODO: Look this up instead.

		personsByLocation := map[string][]*downballotapi.Person{}
		for _, person := range c.Persons {
			coordinates := person.Fields[coordinatesField]
			if coordinates == "" {
				continue
			}
			parts := strings.SplitN(coordinates, ",", 2)
			if len(parts) != 2 {
				continue
			}
			latitudeString := parts[0]
			longitudeString := parts[1]
			if latitudeString == "" || longitudeString == "" {
				continue
			}
			latitude, err := strconv.ParseFloat(latitudeString, 64)
			if err != nil {
				continue
			}
			longitude, err := strconv.ParseFloat(longitudeString, 64)
			if err != nil {
				continue
			}

			locationString := fmt.Sprintf("%0.7f,%0.7f", latitude, longitude)
			personsByLocation[locationString] = append(personsByLocation[locationString], person)
		}

		type Location struct {
			Latitude  float64
			Longitude float64
			Title     string
			Persons   []*downballotapi.Person
		}

		locations := make([]*Location, 0, len(personsByLocation))
		for _, persons := range personsByLocation {
			var latitude float64
			var longitude float64
			{
				person := persons[0]
				coordinates := person.Fields[coordinatesField]
				if coordinates == "" {
					continue
				}
				parts := strings.SplitN(coordinates, ",", 2)
				if len(parts) != 2 {
					continue
				}
				latitudeString := parts[0]
				longitudeString := parts[1]
				if latitudeString == "" || longitudeString == "" {
					continue
				}
				var err error
				latitude, err = strconv.ParseFloat(latitudeString, 64)
				if err != nil {
					continue
				}
				longitude, err = strconv.ParseFloat(longitudeString, 64)
				if err != nil {
					continue
				}
			}

			var title string
			{
				person := persons[0]
				if person.Fields["residential_address"] != "" {
					title += "\n" + person.Fields["residential_address"]
				}
				if person.Fields["residential_address_development"] != "" {
					title += "\n" + person.Fields["residential_address_development"]
				}
				title = strings.TrimSpace(title)
			}
			locations = append(locations, &Location{
				Latitude:  latitude,
				Longitude: longitude,
				Title:     title,
				Persons:   persons,
			})
		}

		totalLatitude := 0.0
		totalLongitude := 0.0
		total := 0

		for _, location := range locations {
			title := location.Title
			if len(location.Persons) == 1 {
				title = location.Persons[0].Fields["name"] + "\n" + title
			}

			var streetNumber string
			{
				person := location.Persons[0]
				parts := strings.SplitN(person.Fields["residential_address"], " ", 2)
				streetNumber = parts[0]
			}
			var apartmentNumber string
			{
				person := location.Persons[0]
				parts := strings.SplitN(person.Fields["residential_address"], ",", 2)
				if len(parts) > 1 {
					streetPart := parts[0]
					parts = strings.SplitN(streetPart, "#", 2)
					if len(parts) > 1 {
						apartmentNumber = parts[1]
					}
				}
			}

			slices.SortFunc(location.Persons, func(left, right *downballotapi.Person) int {
				return strings.Compare(left.Fields["name"], right.Fields["name"])
			})

			var streetNumberBody []app.UI
			{
				streetNumberBody = []app.UI{
					app.Text(streetNumber),
				}
				if apartmentNumber != "" {
					streetNumberBody = append(streetNumberBody, app.Br(), app.Text("#"+apartmentNumber))
				}
			}

			markers = append(markers, googlemap.Marker{
				Coordinate: googlemap.Coordinate{
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
				},
				Title: title,
				Body: func() []app.UI {
					var streetNumberElement app.UI
					if len(location.Persons) == 1 {
						streetNumberElement = app.A().
							Href(fmt.Sprintf("/organization/%s/person/%s", c.organizationID, location.Persons[0].VoterID)).
							Body(streetNumberBody...)
					} else {
						streetNumberElement = app.Span().Body(streetNumberBody...)
					}

					return []app.UI{
						app.Div().
							Class("map-marker").
							Body(
								app.Div().
									Class("map-marker-street-number").
									Body(
										streetNumberElement,
									),
								app.If(len(location.Persons) > 1, func() app.UI {
									return app.Range(location.Persons).Slice(func(i int) app.UI {
										person := location.Persons[i]
										return app.Span().
											Class("map-marker-person").
											Body(
												app.A().
													Href(fmt.Sprintf("/organization/%s/person/%s", c.organizationID, person.VoterID)).
													Text(fmt.Sprintf("%c", 'A'+i)).
													Title(person.Fields["name"] + "\n" + location.Title),
											)
									})
								}),
							),
					}
				},
			})
			totalLatitude += location.Latitude
			totalLongitude += location.Longitude
			total++
		}

		if total > 0 {
			center.Latitude = totalLatitude / float64(total)
			center.Longitude = totalLongitude / float64(total)
		}
	}

	var allPossibleFilterStrings []string
	for _, filter := range c.Filters {
		allPossibleFilterStrings = append(allPossibleFilterStrings, filter.Filter)
	}
	slices.Sort(allPossibleFilterStrings)

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	if c.group == nil {
		return c.EmbeddedPage.Wrap(
			blazar.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDCalltimePage: Render", "PersonsTableVisibleColumns", c.PersonsTableVisibleColumns)

	var personItems []app.UI
	if len(c.Persons) > 0 {
		person := c.Persons[0]

		icon := "circle-question"
		switch person.Fields["candidate.support"] {
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

		summaryItems := []app.UI{
			app.Div().
				Class("person-summary-header").
				Body(
					blazar.Icon().
						Icon(icon),
					app.Div().
						Text(person.Fields["name"]),
				),
			app.Div().
				Class("person-summary-registration").
				Text("Registered as " + person.Fields["political_party"] + ", living in " + person.Fields["district_representative"] + ", " + person.Fields["district_senate"]),
			app.Div().
				Class("person-summary-address").
				Body(
					app.A().
						Href(fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(person.Fields["residential_address"]))).
						Target("_blank").
						Text(person.Fields["residential_address"]),
				),
			app.Div().
				Class("person-summary-phone").
				Text(person.Fields["phone_number"]),
			app.Div().
				Class("person-summary-chips").
				Body(
					app.If(person.Fields["candidate.connected"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("handshake"),
								app.Text("Connected"),
							)
					}),
					app.If(person.Fields["candidate.support"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Support: " + person.Fields["candidate.support"])
					}),
					app.If(person.Fields["candidate.cat"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("cat"),
								app.Text("Cat"),
							)
					}),
					app.If(person.Fields["candidate.dog"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("dog"),
								app.Text("Dog"),
							)
					}),
					app.If(person.Fields["candidate.date_called"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Called on " + person.Fields["candidate.date_called"])
					}),
					app.If(person.Fields["candidate.date_canvassed"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Canvassed on " + person.Fields["candidate.date_canvassed"])
					}),
					app.If(person.Fields["candidate.date_texted"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Texted on " + person.Fields["candidate.date_texted"])
					}),
				),
		}

		personItems = append(personItems,
			app.Div().
				Class("person-summary").
				Body(summaryItems...),
			app.If(person.Fields["candidate.notes"] != "", func() app.UI {
				return app.Hr()
			}),
			app.If(person.Fields["candidate.notes"] != "", func() app.UI {
				return app.Div().
					Class("person-summary-notes").
					Text(person.Fields["candidate.notes"])
			}),
			app.Hr(),
			app.Div().
				Style("display", "flex").
				Style("flex-direction", "row").
				Style("gap", "2em").
				Body(
					blazar.Button().
						Label("Call").
						Icon(component.IconPhone).
						To("tel:+1"+strings.ReplaceAll(person.Fields["phone_number"], "-", "")),
					app.Div().
						Style("font-size", "1.2em").
						Text(person.Fields["phone_number"]),
				),
			app.Hr(),
			blazar.Form().
				Class("no-print").
				Spacer(false).
				Action(
					blazar.FormAction{
						Name: "Edit Person",
						Icon: component.IconEdit,
						To:   fmt.Sprintf("/organization/%s/person/%s", c.organizationID, person.VoterID),
					},
					blazar.FormAction{
						Name: "Bad Number",
						Icon: component.IconEdit,
						Function: func(ctx app.Context) {
							result := app.Window().Call("confirm", "Are you sure you want to mark this as a bad number?")
							slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: Bad Number button clicked", "result", result.Bool())
							if !result.Bool() {
								slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: Bad number button clicked: User cancelled", "result", result.Bool())
								return
							}

							phoneNumber := person.Fields["phone_number"]
							if phoneNumber != "" {
								queryParameters := url.Values{}
								queryParameters.Set("filter", "phone_number = '"+phoneNumber+"'")
								queryParameters.Set("limit", fmt.Sprintf("%d", 100))
								var output downballotapi.ListPersonsResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: search: Could not get persons", "err", err)
									return
								}

								var voterIDs []string
								for _, person := range output.Persons {
									voterIDs = append(voterIDs, person.VoterID)
								}

								if len(voterIDs) > 0 {
									input := downballotapi.PostPersonUpdateRequest{
										VoterIDs: voterIDs,
										Fields: map[string]*string{
											"phone_number": nil,
										},
									}
									var output downballotapi.PostPersonUpdateResponse
									err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/person/update", input, &output)
									if err != nil {
										slog.ErrorContext(ctx.Context, "Could not update persons", "err", err)
										return
									}
								}
							}

							c.search(ctx)
						},
					},
					blazar.FormAction{
						Name: "Called",
						Icon: component.IconEdit,
						Function: func(ctx app.Context) {
							phoneNumber := person.Fields["phone_number"]
							if phoneNumber != "" {
								queryParameters := url.Values{}
								queryParameters.Set("filter", "phone_number = '"+phoneNumber+"'")
								queryParameters.Set("limit", fmt.Sprintf("%d", 100))
								var output downballotapi.ListPersonsResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: search: Could not get persons", "err", err)
									return
								}

								var voterIDs []string
								for _, person := range output.Persons {
									voterIDs = append(voterIDs, person.VoterID)
								}

								if len(voterIDs) > 0 {
									dateCalled := time.Now().Format("2006-01-02")
									input := downballotapi.PostPersonUpdateRequest{
										VoterIDs: voterIDs,
										Fields: map[string]*string{
											"candidate.date_called": &dateCalled,
										},
									}
									var output downballotapi.PostPersonUpdateResponse
									err := api.Do(ctx, http.MethodPost, "/api/v1/organization/"+c.organizationID+"/person/update", input, &output)
									if err != nil {
										slog.ErrorContext(ctx.Context, "Could not update persons", "err", err)
										return
									}
								}
							}

							c.search(ctx)
						},
					},
				),
		)
	}

	return c.EmbeddedPage.Wrap(
		blazar.Collapse().
			Label("Filter").
			Bind(&c.filterOpen).
			OnOpenChange(func(ctx app.Context, open bool) {
				ctx.SetState("persist-organization-id-group-id-person-page-filter-open", open).Persist()
			}).
			SummaryText(func() string {
				summary := "Filter: "
				if c.Filter == "" {
					summary += "n/a"
				} else {
					var namedFilter string
					for _, filter := range c.Filters {
						if filter.Filter == c.Filter {
							namedFilter = filter.Name
							break
						}
					}
					if namedFilter == "" {
						summary += c.Filter
					} else {
						summary += namedFilter
					}
				}

				return summary
			}()).
			Body(
				app.Div().
					Style("display", "flex").
					Style("flex-direction", "column").
					Body(
						blazar.Select().
							Name("saved_filter").
							Label("Saved Filter").
							AllowedValue(func() []blazar.SelectOption {
								var allowedValues []blazar.SelectOption
								allowedValues = append(allowedValues, blazar.SelectOption{Label: "Select a filter or create your own", Value: ""})
								for _, filter := range c.Filters {
									allowedValues = append(allowedValues, blazar.SelectOption{Label: filter.Name, Value: filter.Filter})
								}
								return allowedValues
							}()...).
							Bind(&c.Filter).
							On("change", func(ctx app.Context, e app.Event) {
								c.ValueTo(&c.Filter)(ctx, e)
								ctx.SetState("persist-organization-id-group-id-person-page-filter", c.Filter).Persist()
								ctx.Update() // Update so that the other input can be updated.
							}),
						blazar.Input[string]().
							Label("Filter").
							Type("text").
							Placeholder("key = 'value' or ...").
							Clearable(true).
							Bind(&c.Filter).
							On("change", func(ctx app.Context, e app.Event) {
								ctx.SetState("persist-organization-id-group-id-person-page-filter", c.Filter).Persist()
								ctx.Update() // Update so that the other input can be updated.
							}),
					),
			),
		blazar.Form().
			Class("no-print").
			Spacer(false).
			Action(
				blazar.FormAction{
					Name:     "Search",
					Icon:     component.IconSearch,
					Function: c.search,
				},
			),
		app.If(len(personItems) > 0, func() app.UI {
			return app.Div().
				Body(personItems...)
		}),
	)
}

func (c *OrganizationIDGroupIDCalltimePage) search(ctx app.Context) {
	queryParameters := url.Values{}
	queryParameters.Set("filter", c.Filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", 100))
	var output downballotapi.ListPersonsResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: search: Could not get persons", "err", err)
		return
	}

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: search: Setting persons", "len(persons)", len(output.Persons))
	c.Persons = output.Persons

	rand.Shuffle(len(c.Persons), func(i, j int) {
		c.Persons[i], c.Persons[j] = c.Persons[j], c.Persons[i]
	})

	ctx.Update()
}
