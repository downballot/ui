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
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/googlemap"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/restapiclient"
)

type OrganizationIDGroupIDPersonPage struct {
	app.Compo
	myui.EmbeddedPage

	loaded bool

	organizationID string
	organization   *downballotapi.Organization
	groupID        string
	group          *downballotapi.Group

	Filter         string
	Limit          uint
	PossibleFields []string
	Error          string
	Persons        []*downballotapi.Person
	Filters        []*downballotapi.Filter

	PersonsTableColumns        []myui.TableColumn[*downballotapi.Person]
	PersonsTableVisibleColumns []string
	BindPersonsTable           myui.TableBinding[*downballotapi.Person]

	FilterOpen  bool
	MapOpen     bool
	ResultsOpen bool
}

var _ app.Navigator = (*OrganizationIDGroupIDPersonPage)(nil)

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

	c.Limit = 1000
	c.FilterOpen = false
	c.MapOpen = true
	c.ResultsOpen = true
	c.BindPersonsTable.PageIndex = 0
	c.BindPersonsTable.PageSize = 10

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

			c.PersonsTableColumns = []myui.TableColumn[*downballotapi.Person]{
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
				c.PersonsTableColumns = append(c.PersonsTableColumns, myui.TableColumn[*downballotapi.Person]{
					Name: name,
					Value: func(row *downballotapi.Person) any {
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
				"residential_address_development",
				"candidate.notes",
			}
			slices.Sort(c.PersonsTableVisibleColumns)
			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonPage: OnNav: Async", "len(PersonsTableVisibleColumns)", len(c.PersonsTableVisibleColumns))
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Loading complete.  Searching for persons.")
			c.search(ctx, app.Event{})
		})
	})
}

func streetCanon(address string) string {
	parts := strings.SplitN(address, " ", 2)
	addressNumberString := parts[0]
	remainder := parts[1]
	{
		addressNumber, err := strconv.ParseInt(addressNumberString, 10, 64)
		if err == nil {
			addressNumberString = fmt.Sprintf("%09d", addressNumber)
			if addressNumber%2 == 0 {
				addressNumberString = "e" + addressNumberString
			} else {
				addressNumberString = "o" + addressNumberString
			}
		}
	}
	parts = strings.SplitN(remainder, ",", 2)
	streetInfo := parts[0]
	remainder = parts[1]
	return streetInfo + " " + addressNumberString + "," + remainder
}

func (c *OrganizationIDGroupIDPersonPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonPage: Render")

	markers := []googlemap.Marker{}
	center := googlemap.Coordinate{
		Latitude:  39.713171422509426,
		Longitude: -75.75937795659787,
	}
	{
		totalLatitude := 0.0
		totalLongitude := 0.0
		total := 0

		coordinatesField := "coordinates" // TODO: Look this up instead.

		for _, person := range c.Persons {
			title := person.Fields["name"]
			if person.Fields["residential_address"] != "" {
				title += "\n" + person.Fields["residential_address"]
			}
			if person.Fields["residential_address_development"] != "" {
				title += "\n" + person.Fields["residential_address_development"]
			}

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
			markers = append(markers, googlemap.Marker{
				Coordinate: googlemap.Coordinate{
					Latitude:  latitude,
					Longitude: longitude,
				},
				Title: title,
				OnClick: func(ctx app.Context, event app.Event) {
					ctx.PreventUpdate()

					app.Window().Call("open", fmt.Sprintf("/organization/%s/person/%s", c.organizationID, person.VoterID), "_blank")
				},
			})
			totalLatitude += latitude
			totalLongitude += longitude
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
			myui.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonPage: Render", "PersonsTableVisibleColumns", c.PersonsTableVisibleColumns)
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonPage: Render", "BindPersonsTable", c.BindPersonsTable)

	return c.EmbeddedPage.Wrap(
		myui.Collapse().
			Label("Filter").
			Bind(&c.FilterOpen).
			SummaryText(func() string {
				summary := "Filter: "
				if c.Filter == "" {
					summary += "n/a"
				} else {
					summary += c.Filter
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
						myui.Select().
							Name("saved_filter").
							Label("Saved Filter").
							AllowedValue(
								myui.SelectOption{Label: "Select a filter or create your own", Value: ""},
							).
							AllowedValue(func() []myui.SelectOption {
								var allowedValues []myui.SelectOption
								for _, filter := range c.Filters {
									allowedValues = append(allowedValues, myui.SelectOption{Label: filter.Name, Value: filter.Filter})
								}
								return allowedValues
							}()...).
							Bind(&c.Filter).
							On("change", func(ctx app.Context, e app.Event) {
								c.ValueTo(&c.Filter)(ctx, e)
								ctx.SetState("persist-organization-id-group-id-person-page-filter", c.Filter).Persist()
								ctx.Update() // Update so that the other input can be updated.
							}),
						myui.Input[string]().
							Label("Filter").
							Type("text").
							Placeholder("key = 'value' or ...").
							Bind(&c.Filter).
							On("change", func(ctx app.Context, e app.Event) {
								ctx.SetState("persist-organization-id-group-id-person-page-filter", c.Filter).Persist()
								ctx.Update() // Update so that the other input can be updated.
							}),
						myui.Input[uint]().
							Label("Limit").
							Type("number").
							Placeholder("1000").
							Bind(&c.Limit).
							On("change", func(ctx app.Context, e app.Event) {
								ctx.SetState("persist-organization-id-group-id-person-page-limit", c.Limit).Persist()
							}),
					),
			),
		app.Div().
			Class("no-print").
			Body(
				myui.Button().
					Label("Search").
					On("click", c.search),
				myui.Button().
					Label("CSV").
					On("click", c.csv),
			),
		app.If(c.Error != "", func() app.UI {
			return myui.StatusBar().
				Text(c.Error).
				Bad()
		}),
		myui.Collapse().
			Label("Map").
			Bind(&c.MapOpen).
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
		myui.Collapse().
			Label("Results").
			Bind(&c.ResultsOpen).
			SummaryText("Results: "+fmt.Sprintf("%d", len(c.Persons))).
			Body(
				myui.Table[*downballotapi.Person]().
					Bind(&c.BindPersonsTable).
					Columns(c.PersonsTableColumns).
					BindVisibleColumns(&c.PersonsTableVisibleColumns).
					Rows(c.Persons),
			),
	)
}

func (c *OrganizationIDGroupIDPersonPage) search(ctx app.Context, e app.Event) {
	queryParameters := url.Values{}
	queryParameters.Set("filter", c.Filter)
	queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
	var output downballotapi.ListPersonsResponse
	err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.organizationID+"/group/"+c.groupID+"/person?"+queryParameters.Encode(), nil, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "Could not get persons", "err", err)
		c.Error = err.Error()
		return
	}
	c.Error = ""

	slog.InfoContext(ctx.Context, "Setting persons", "len(persons)", len(output.Persons))
	c.Persons = output.Persons

	slices.SortFunc(c.Persons, func(left, right *downballotapi.Person) int {
		leftAddress := streetCanon(left.Fields["residential_address"])
		rightAddress := streetCanon(right.Fields["residential_address"])

		return strings.Compare(leftAddress, rightAddress)
	})

	ctx.Update()
}

func (c *OrganizationIDGroupIDPersonPage) csv(ctx app.Context, e app.Event) {
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
