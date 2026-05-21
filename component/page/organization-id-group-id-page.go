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
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/tekkamanendless/restapiclient"
)

type OrganizationIDGroupIDPage struct {
	app.Compo

	Loaded bool

	OrganizationID string `route:"organization_id"`
	Organization   *downballotapi.Organization
	GroupID        string `route:"group_id"`
	Group          *downballotapi.Group
	Children       []*downballotapi.Group

	Filter         string
	Limit          uint
	PossibleFields []string
	SelectedFields []string
	Error          string
	Persons        []*downballotapi.Person

	PersonsTable *myui.MyUITable[*downballotapi.Person]
}

var _ app.Navigator = (*OrganizationIDGroupIDPage)(nil)

func (c *OrganizationIDGroupIDPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnNav", "OrganizationID", c.OrganizationID)

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

	ctx.Update()
}

func (c *OrganizationIDGroupIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "GroupID", c.GroupID)

	if c.OrganizationID == "" {
		return
	}

	c.Limit = 1000

	c.PersonsTable = myui.Table[*downballotapi.Person]().
		PageSize(10)

	ctx.Async(func() {
		var wg sync.WaitGroup
		wg.Go(func() {
			var output downballotapi.GetOrganizationResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
				return
			}
		})
		wg.Go(func() {
			var output downballotapi.GetGroupResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting group", "group", output.Group)
				c.Group = output.Group
			})
		})
		wg.Go(func() {
			var output downballotapi.ListGroupsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group?parent_id="+c.GroupID, nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get groups", "err", err)
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting children", "groups", output.Groups)
				c.Children = output.Groups
			})
		})
		wg.Go(func() {
			var output downballotapi.ListPersonFieldsResponse
			err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/person-field", nil, &output)
			if err != nil {
				slog.ErrorContext(ctx.Context, "Could not get person fields", "err", err)
				return
			}

			var possibleFields []string
			for _, personField := range output.PersonFields {
				possibleFields = append(possibleFields, personField.Name)
			}
			slices.Sort(possibleFields)

			ctx.Dispatch(func(ctx app.Context) {
				slog.InfoContext(ctx.Context, "Dispatch: Setting person fields", "person fields", output.PersonFields)
				c.PossibleFields = possibleFields

				c.SelectedFields = []string{
					"name",
					"phone_number",
					"residential_address",
					"residential_address_development",
					"candidate.notes",
				}
				slices.Sort(c.SelectedFields)
			})
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.Loaded = true
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

func (c *OrganizationIDGroupIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render")

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

					app.Window().Call("open", fmt.Sprintf("/organization/%s/person/%s", c.OrganizationID, person.VoterID), "_blank")
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

	if !c.Loaded {
		return nil
	}

	if c.Group == nil {
		return myui.StatusBar().
			Text("Not found").
			Bad()
	}

	return myui.Page().
		Body(
			myui.Table[*downballotapi.Group]().
				Title("Sub-groups").
				Rows(c.Children).
				Columns([]myui.TableColumn[*downballotapi.Group]{
					{
						Name: "ID",
						Value: func(row *downballotapi.Group) any {
							return row.ID
						},
					},
					{
						Name: "Name",
						Value: func(row *downballotapi.Group) any {
							return row.Name
						},
						To: func(row *downballotapi.Group) string {
							return fmt.Sprintf("/organization/%s/group/%s", c.OrganizationID, row.ID)
						},
					},
				}).
				Action(myui.TableAction{
					Name: "New group",
					Icon: "plus",
					To: func() string {
						return fmt.Sprintf("/organization/%s/group/new?parent_id=%s", c.OrganizationID, c.GroupID)
					},
				}).
				Render(),
			app.Div().
				Style("display", "flex").
				Style("flex-direction", "column").
				Body(
					myui.Input().
						Label("Filter").
						Type("text").
						Placeholder("key = 'value' or ...").
						Value(c.Filter).
						On("change", c.ValueTo(&c.Filter)),
					myui.Input().
						Label("Limit").
						Type("number").
						Placeholder("1000").
						Value(fmt.Sprintf("%d", c.Limit)).
						On("change", c.ValueTo(&c.Limit)),
					myui.Multiselect().
						Label("Fields").
						AllowedValue(func() []myui.SelectOption {
							selectOptions := []myui.SelectOption{}
							for _, field := range c.PossibleFields {
								selectOptions = append(selectOptions, myui.SelectOption{
									Label: field,
									Value: field,
								})
							}
							return selectOptions
						}()...).
						SelectedValue(c.SelectedFields...).
						On("change", myui.SelectedValuesTo(&c.SelectedFields)),
					app.Div().Body(
						myui.Button().
							Label("Search").
							On("click", func(ctx app.Context, e app.Event) {
								queryParameters := url.Values{}
								queryParameters.Set("filter", c.Filter)
								queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
								//queryParameters.Set("fields", strings.Join(c.SelectedFields, ",")) // Always get all fields.
								var output downballotapi.ListPersonsResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID+"/person?"+queryParameters.Encode(), nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get persons", "err", err)
									c.Error = err.Error()
									return
								}
								c.Error = ""

								ctx.Dispatch(func(ctx app.Context) {
									slog.InfoContext(ctx.Context, "Dispatch: Setting persons", "persons", output.Persons)
									c.Persons = output.Persons

									slices.SortFunc(c.Persons, func(left, right *downballotapi.Person) int {
										leftAddress := streetCanon(left.Fields["residential_address"])
										rightAddress := streetCanon(right.Fields["residential_address"])

										return strings.Compare(leftAddress, rightAddress)
									})

									columns := []myui.TableColumn[*downballotapi.Person]{
										{
											Name: "Voter ID",
											Value: func(row *downballotapi.Person) any {
												return row.VoterID
											},
											To: func(row *downballotapi.Person) string {
												return fmt.Sprintf("/organization/%s/person/%s", c.OrganizationID, row.VoterID)
											},
										},
									}

									for _, name := range c.SelectedFields {
										columns = append(columns, myui.TableColumn[*downballotapi.Person]{
											Name: name,
											Value: func(row *downballotapi.Person) any {
												return row.Fields[name]
											},
										})
									}

									c.PersonsTable.Columns(columns)
									c.PersonsTable.Rows(output.Persons)
								})
							}),
						myui.Button().
							Label("CSV").
							On("click", func(ctx app.Context, e app.Event) {
								queryParameters := url.Values{}
								queryParameters.Set("filter", c.Filter)
								queryParameters.Set("limit", fmt.Sprintf("%d", c.Limit))
								queryParameters.Set("fields", strings.Join(c.SelectedFields, ","))
								var output restapiclient.RawBytes
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID+"/group/"+c.GroupID+"/person?"+queryParameters.Encode(), nil, &output, restapiclient.OptionHeader("Accept", "text/csv"))
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
							}),
					),
				),
			app.If(c.Error != "", func() app.UI {
				return myui.StatusBar().
					Text(c.Error).
					Bad()
			}),
			app.Div().
				Class("map-container").
				Style("width", "100%").
				Style("height", "600px").
				Body(
					googlemap.GoogleMap().
						APIKey(app.Getenv("GOOGLE_MAPS_API_KEY")).
						Center(center).
						Markers(markers),
					//Render(),
				),
			app.Div().
				Body(
					app.Text("Total: "+fmt.Sprintf("%d", len(c.Persons))),
				),
			c.PersonsTable.Render(),
		)
}
