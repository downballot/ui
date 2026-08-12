package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/downballot/ui/googlemap"
	"github.com/downballot/ui/street"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/deref"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/restapiclient"
)

type OrganizationIDPersonSearchPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	permissionSet  permissionset.PermissionSet

	Filter         string
	Limit          uint
	splitEvenOdd   bool
	PossibleFields []string
	Persons        []*downballotapi.Person

	PersonsTableColumns        []blazar.TableColumn[*downballotapi.Person]
	PersonsTableVisibleColumns []string

	mapOpen     bool
	resultsOpen bool
	pageSize    uint
}

var _ app.Navigator = (*OrganizationIDPersonSearchPage)(nil)
var _ app.Mounter = (*OrganizationIDPersonSearchPage)(nil)

func (c *OrganizationIDPersonSearchPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnMount")

	c.EmbeddedPage.OnMount(ctx)
}

func (c *OrganizationIDPersonSearchPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)

	slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav", "OrganizationID", c.organizationID)

	if c.organizationID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	c.Limit = 50
	c.splitEvenOdd = true
	c.mapOpen = true
	c.resultsOpen = true

	var persistLimit uint
	ctx.GetState("persist-organization-id-person-search-page-limit", &persistLimit)
	if persistLimit != 0 {
		c.Limit = persistLimit
	}

	var persistMapOpen *bool
	ctx.GetState("persist-organization-id-person-search-page-map-open", &persistMapOpen)
	if persistMapOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Persist map open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Persist map open is not nil.", "persistMapOpen", deref.String(persistMapOpen))
		c.mapOpen = *persistMapOpen
	}

	var persistResultsOpen *bool
	ctx.GetState("persist-organization-id-person-search-page-results-open", &persistResultsOpen)
	if persistResultsOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Persist results open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Persist results open is not nil.", "persistResultsOpen", deref.String(persistResultsOpen))
		c.resultsOpen = *persistResultsOpen
	}

	var persistPageSize uint
	ctx.GetState("persist-organization-id-person-search-page-page-size", &persistPageSize)
	if persistPageSize != 0 {
		c.pageSize = persistPageSize
	} else {
		c.pageSize = 10
	}

	if value := ctx.Page().URL().Query().Get("filter"); value != "" {
		c.Filter = value
	}
	if value := ctx.Page().URL().Query().Get("limit"); value != "" {
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not parse limit", "err", err)
		} else {
			c.Limit = uint(uintValue)
		}
	}

	ctx.Async(func() {
		var wg sync.WaitGroup
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

			slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Async", "len(PersonsTableColumns)", len(c.PersonsTableColumns))

			c.PossibleFields = possibleFields
			slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Async", "len(PossibleFields)", len(c.PossibleFields))

			c.PersonsTableVisibleColumns = []string{
				"Voter ID",
				"name",
				"phone_number",
				"residential_address",
			}
			if fieldNameMap["computed.age"] {
				c.PersonsTableVisibleColumns = append(c.PersonsTableVisibleColumns, "computed.age")
			} else {
				c.PersonsTableVisibleColumns = append(c.PersonsTableVisibleColumns, "birthday_year")
			}
			if fieldNameMap["computed.likely"] {
				c.PersonsTableVisibleColumns = append(c.PersonsTableVisibleColumns, "computed.likely")
			}
			slices.Sort(c.PersonsTableVisibleColumns)
			slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: OnNav: Async", "len(PersonsTableVisibleColumns)", len(c.PersonsTableVisibleColumns))
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Loading complete.")
		})
	})
}

func (c *OrganizationIDPersonSearchPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDPersonSearchPage: Render")

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

	if !c.loaded {
		return c.EmbeddedPage.Wrap(
			app.Div().Text("Loading..."),
		)
	}

	slog.InfoContext(context.TODO(), "OrganizationIDPersonSearchPage: Render", "PersonsTableVisibleColumns", c.PersonsTableVisibleColumns)

	addFieldDialog := component.AddFieldMultipleDialog().
		OrganizationID(c.organizationID)

	return c.EmbeddedPage.Wrap(
		app.Div().
			Style("display", "flex").
			Style("flex-direction", "column").
			Body(
				blazar.Input[string]().
					Label("Search").
					Type("text").
					Placeholder("name, phone number, address, etc.").
					Clearable(true).
					Bind(&c.Filter).
					On("change", func(ctx app.Context, e app.Event) {
						ctx.SetState("persist-organization-id-person-search-page-filter", c.Filter).Persist()
						ctx.Update() // Update so that the other input can be updated.
					}),
				blazar.Input[uint]().
					Label("Limit").
					Type("number").
					Placeholder("1000").
					Bind(&c.Limit).
					On("change", func(ctx app.Context, e app.Event) {
						ctx.SetState("persist-organization-id-person-search-page-limit", c.Limit).Persist()
					}),
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
				blazar.FormAction{
					Name:     "CSV",
					Icon:     component.IconDownload,
					Function: c.csv,
				},
			),
		blazar.Collapse().
			Label("Map").
			Bind(&c.mapOpen).
			OnOpenChange(func(ctx app.Context, open bool) {
				ctx.SetState("persist-organization-id-person-search-page-map-open", open).Persist()
			}).
			Body(
				app.Div().
					Class("map-container").
					Style("width", "100%").
					Style("height", "600px").
					Body(
						googlemap.GoogleMap().
							APIKey(app.Getenv("GOOGLE_MAPS_API_KEY")).
							Center(center).
							Markers(markers),
					),
			),
		blazar.Collapse().
			Label("Results").
			Bind(&c.resultsOpen).
			OnOpenChange(func(ctx app.Context, open bool) {
				ctx.SetState("persist-organization-id-person-search-page-results-open", open).Persist()
			}).
			SummaryText("Results: "+fmt.Sprintf("%d", len(c.Persons))).
			Body(
				blazar.Table[*downballotapi.Person]().
					PageSize(c.pageSize).
					Columns(c.PersonsTableColumns).
					VisibleColumns(c.PersonsTableVisibleColumns).
					RowIDFunction(func(row *downballotapi.Person) string {
						return row.VoterID
					}).
					Rows(c.Persons).
					MultiRowAction(
						blazar.MultiRowAction[*downballotapi.Person]{
							Name: "Edit",
							Icon: component.IconEdit,
							Function: func(ctx app.Context, rows []*downballotapi.Person) {
								voterIDs := []string{}
								for _, row := range rows {
									voterIDs = append(voterIDs, row.VoterID)
								}
								addFieldDialog.Open(ctx, voterIDs)
							},
							Disabled: !c.permissionSet.Match(iam.IAMPersonUpdate),
						},
					).
					OnPageSizeChange(func(ctx app.Context, pageSize uint) {
						ctx.SetState("persist-organization-id-person-search-page-page-size", pageSize).Persist()
					}),
				addFieldDialog,
			),
	)
}

var phoneNumberRegex = regexp.MustCompile(`^[0-9() +-]+$`)

func formatPhoneNumber(input string) string {
	var output string
	for _, char := range input {
		if char >= '0' && char <= '9' {
			output += string(char)
		}
	}
	if len(output) == 11 {
		output = strings.TrimPrefix(output, "1")
	}
	if len(output) == 10 {
		return output[:3] + "-" + output[3:6] + "-" + output[6:]
	}
	return ""
}

func (c *OrganizationIDPersonSearchPage) computeFilter() string {
	if c.Filter == "" {
		return ""
	}

	if phoneNumberRegex.MatchString(c.Filter) {
		slog.InfoContext(context.TODO(), "OrganizationIDPersonSearchPage: computeFilter: Phone number", "Filter", c.Filter)
		phoneNumber := formatPhoneNumber(c.Filter)
		slog.InfoContext(context.TODO(), "OrganizationIDPersonSearchPage: computeFilter: formatted phone number", "Phone number", phoneNumber)
		if phoneNumber != "" {
			return "phone_number = '" + phoneNumber + "'"
		}
	}

	slog.InfoContext(context.TODO(), "OrganizationIDPersonSearchPage: computeFilter: Filter", "Filter", c.Filter)
	var orParts []string
	orParts = append(orParts, "residential_address ~ '"+c.Filter+"*'")
	orParts = append(orParts, "name ~ '"+strings.Join(strings.Split(c.Filter, " "), "*")+"'")
	return strings.Join(orParts, " or ")
}

func (c *OrganizationIDPersonSearchPage) search(ctx app.Context) {
	filter := c.computeFilter()
	if filter == "" {
		c.Persons = []*downballotapi.Person{}
		ctx.Update()
		return
	}

	queryParameters := url.Values{}
	queryParameters.Set("filter", filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
	var output downballotapi.ListPersonsResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person?"+queryParameters.Encode(), nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "OrganizationIDPersonSearchPage: search: Could not get persons", "err", err)
		return
	}

	slog.InfoContext(ctx.Context, "OrganizationIDPersonSearchPage: search: Setting persons", "len(persons)", len(output.Persons))
	c.Persons = output.Persons

	slices.SortFunc(c.Persons, func(left, right *downballotapi.Person) int {
		leftAddress := street.Canon(left.Fields["residential_address"], c.splitEvenOdd)
		rightAddress := street.Canon(right.Fields["residential_address"], c.splitEvenOdd)

		return strings.Compare(leftAddress, rightAddress)
	})

	ctx.Update()
}

func (c *OrganizationIDPersonSearchPage) csv(ctx app.Context) {
	filter := c.computeFilter()
	if filter == "" {
		return
	}

	queryParameters := url.Values{}
	queryParameters.Set("filter", filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
	var output restapiclient.RawBytes
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/person?"+queryParameters.Encode(), nil, &output, restapiclient.OptionHeader("Accept", "text/csv"))
	if err != nil {
		slog.ErrorContext(ctx.Context, "Could not get persons", "err", err)
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		slog.InfoContext(ctx.Context, "Dispatch: Saving CSV")
		blobConstructor := app.Window().Get("Blob")
		arrayConstructor := app.Window().Get("Array")
		array := arrayConstructor.New(string(output))
		blob := blobConstructor.New(array, map[string]any{"type": "text/csv"})

		aElement := app.Window().Get("document").Call("createElement", "a")
		aElement.Set("style", "display: none;")
		aElement.Set("href", app.Window().Get("URL").Call("createObjectURL", blob))
		aElement.Set("download", "persons.csv")

		app.Window().Get("document").Get("body").Call("appendChild", aElement)
		aElement.Call("click")
		app.Window().Get("document").Get("body").Call("removeChild", aElement)
	})
}
