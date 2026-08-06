package page

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

type OrganizationIDGroupIDPersonPage struct {
	app.Compo
	component.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	groupID        string
	group          *downballotapi.Group
	permissionSet  permissionset.PermissionSet

	Filter         string
	Limit          uint
	splitEvenOdd   bool
	PossibleFields []string
	Persons        []*downballotapi.Person
	Filters        []*downballotapi.Filter

	PersonsTableColumns        []blazar.TableColumn[*downballotapi.Person]
	PersonsTableVisibleColumns []string

	filterOpen  bool
	mapOpen     bool
	resultsOpen bool
	pageSize    uint
}

var _ app.Navigator = (*OrganizationIDGroupIDPersonPage)(nil)
var _ app.Mounter = (*OrganizationIDGroupIDPersonPage)(nil)

func (c *OrganizationIDGroupIDPersonPage) OnMount(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnMount")

	c.EmbeddedPage.OnMount(ctx)
}

func (c *OrganizationIDGroupIDPersonPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav", "GroupID", c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.groupID == "" {
		return
	}

	ctx.GetState("organization/"+c.organizationID+"/permission-set", &c.permissionSet)

	c.Limit = 1000
	c.splitEvenOdd = true
	c.filterOpen = false
	c.mapOpen = true
	c.resultsOpen = true

	var persistFilter string
	ctx.GetState("persist-organization-id-group-id-person-page-filter", &persistFilter)
	if persistFilter != "" {
		c.Filter = persistFilter
	}

	var persistLimit uint
	ctx.GetState("persist-organization-id-group-id-person-page-limit", &persistLimit)
	if persistLimit != 0 {
		c.Limit = persistLimit
	}

	var persistSplitEvenOdd *bool
	ctx.GetState("persist-organization-id-group-id-person-page-split-even-odd", &persistSplitEvenOdd)
	if persistSplitEvenOdd == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist split even odd is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist split even odd is not nil.", "persistSplitEvenOdd", deref.String(persistSplitEvenOdd))
		c.splitEvenOdd = *persistSplitEvenOdd
	}

	var persistFilterOpen *bool
	ctx.GetState("persist-organization-id-group-id-person-page-filter-open", &persistFilterOpen)
	if persistFilterOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist filter open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist filter open is not nil.", "persistFilterOpen", deref.String(persistFilterOpen))
		c.filterOpen = *persistFilterOpen
	}

	var persistMapOpen *bool
	ctx.GetState("persist-organization-id-group-id-person-page-map-open", &persistMapOpen)
	if persistMapOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist map open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist map open is not nil.", "persistMapOpen", deref.String(persistMapOpen))
		c.mapOpen = *persistMapOpen
	}

	var persistResultsOpen *bool
	ctx.GetState("persist-organization-id-group-id-person-page-results-open", &persistResultsOpen)
	if persistResultsOpen == nil {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist results open is nil.")
	} else {
		slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Persist results open is not nil.", "persistResultsOpen", deref.String(persistResultsOpen))
		c.resultsOpen = *persistResultsOpen
	}

	var persistPageSize uint
	ctx.GetState("persist-organization-id-group-id-person-page-page-size", &persistPageSize)
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
				c.PersonsTableColumns = append(c.PersonsTableColumns, blazar.TableColumn[*downballotapi.Person]{
					Name: name,
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

			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Async", "len(PersonsTableColumns)", len(c.PersonsTableColumns))

			c.PossibleFields = possibleFields
			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Async", "len(PossibleFields)", len(c.PossibleFields))

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
			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Async", "len(PersonsTableVisibleColumns)", len(c.PersonsTableVisibleColumns))
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Loading complete.  Searching for persons.")
			c.search(ctx)
		})
	})
}

func (c *OrganizationIDGroupIDPersonPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonPage: Render")

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

	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonPage: Render", "PersonsTableVisibleColumns", c.PersonsTableVisibleColumns)

	addFieldDialog := component.AddFieldMultipleDialog().
		OrganizationID(c.organizationID)

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

				summary += " | Limit: "
				if c.Limit == 0 {
					summary += "n/a"
				} else {
					summary += fmt.Sprintf("%d", c.Limit)
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
						blazar.Input[uint]().
							Label("Limit").
							Type("number").
							Placeholder("1000").
							Bind(&c.Limit).
							On("change", func(ctx app.Context, e app.Event) {
								ctx.SetState("persist-organization-id-group-id-person-page-limit", c.Limit).Persist()
							}),
						blazar.Input[bool]().
							Label("Split Even and Odd addresses").
							Bind(&c.splitEvenOdd).
							On("change", func(ctx app.Context, e app.Event) {
								ctx.SetState("persist-organization-id-group-id-person-page-split-even-odd", c.splitEvenOdd).Persist()
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
				blazar.FormAction{
					Name:     "CSV",
					Icon:     component.IconDownload,
					Function: c.csv,
				},
				blazar.FormAction{
					Name:     "Mailing Labels",
					Icon:     component.IconMailingLabel,
					Function: c.mailingLabels,
				},
			),
		blazar.Collapse().
			Label("Map").
			Bind(&c.mapOpen).
			OnOpenChange(func(ctx app.Context, open bool) {
				ctx.SetState("persist-organization-id-group-id-person-page-map-open", open).Persist()
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
				ctx.SetState("persist-organization-id-group-id-person-page-results-open", open).Persist()
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
						ctx.SetState("persist-organization-id-group-id-person-page-page-size", pageSize).Persist()
					}),
				addFieldDialog,
			),
	)
}

func (c *OrganizationIDGroupIDPersonPage) search(ctx app.Context) {
	queryParameters := url.Values{}
	queryParameters.Set("filter", c.Filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
	var output downballotapi.ListPersonsResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "OrganizationIDGroupIDPersonPage: search: Could not get persons", "err", err)
		return
	}

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: search: Setting persons", "len(persons)", len(output.Persons))
	c.Persons = output.Persons

	slices.SortFunc(c.Persons, func(left, right *downballotapi.Person) int {
		leftAddress := street.Canon(left.Fields["residential_address"], c.splitEvenOdd)
		rightAddress := street.Canon(right.Fields["residential_address"], c.splitEvenOdd)

		return strings.Compare(leftAddress, rightAddress)
	})

	ctx.Update()
}

func (c *OrganizationIDGroupIDPersonPage) mailingLabels(ctx app.Context) {
	ctx.PreventUpdate()

	queryParameters := url.Values{}
	queryParameters.Set("filter", c.Filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))

	ctx.Navigate("/organization/" + c.organizationID + "/group/" + c.groupID + "/person-mailing-labels?" + queryParameters.Encode())
}

func (c *OrganizationIDGroupIDPersonPage) csv(ctx app.Context) {
	queryParameters := url.Values{}
	queryParameters.Set("filter", c.Filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
	var output restapiclient.RawBytes
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output, restapiclient.OptionHeader("Accept", "text/csv"))
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
