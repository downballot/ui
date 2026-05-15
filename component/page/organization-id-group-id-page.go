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

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/googlemap"
	"github.com/downballot/ui/myui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/tekkamanendless/restapiclient"
)

type OrganizationIDGroupIDPage struct {
	app.Compo

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
}

func (c *OrganizationIDGroupIDPage) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate")
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "OrganizationID", c.OrganizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPage: OnUpdate", "GroupID", c.GroupID)

	if c.OrganizationID == "" {
		return
	}

	c.Limit = 1000

	ctx.Async(func() {
		var output downballotapi.GetOrganizationResponse
		err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+c.OrganizationID, nil, &output)
		if err != nil {
			slog.ErrorContext(ctx.Context, "Could not get organizations", "err", err)
			return
		}
	})
	ctx.Async(func() {
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
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Group should be set", "group", c.Group)

			//ctx.Update()
		})
	})
	ctx.Async(func() {
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
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Children should be set", "groups", c.Children)

			//ctx.Update()
		})
	})
	ctx.Async(func() {
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
				"residential_address_development",
				"candidate.notes",
			}
			slices.Sort(c.SelectedFields)
		})
		ctx.Defer(func(ctx app.Context) {
			slog.InfoContext(ctx.Context, "Defer: Children should be set", "groups", c.Children)

			//ctx.Update()
		})
	})
}

func (c *OrganizationIDGroupIDPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPage: Render")

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

	return app.Div().Body(
		app.If(c.Group == nil, func() app.UI {
			return app.Div().Text("Not found")
		}).Else(func() app.UI {
			return app.Div().Body(
				myui.Table[*downballotapi.Group]().
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
					}).Render(),
				app.Div().Body(
					app.Div().Text("Filter:"),
					app.Div().Body(
						app.Input().
							Placeholder("key = 'value' or ...").
							Value(c.Filter).
							OnChange(c.ValueTo(&c.Filter)),
					),
					app.Div().Text("Limit:"),
					app.Div().Body(
						app.Input().
							Type("number").
							Value(c.Limit).
							OnChange(c.ValueTo(&c.Limit)),
					),

					app.Div().Body(
						app.Range(c.PossibleFields).Slice(func(i int) app.UI {
							possibleField := c.PossibleFields[i]

							return app.Label().Body(
								app.Input().
									Type("checkbox").
									Checked(slices.Contains(c.SelectedFields, possibleField)).
									OnChange(func(ctx app.Context, e app.Event) {
										checked := e.Value.Get("target").Get("checked").Bool()
										if checked {
											c.SelectedFields = append(c.SelectedFields, possibleField)
										} else {
											c.SelectedFields = slices.DeleteFunc(c.SelectedFields, func(item string) bool { return item == possibleField })
										}
										slices.Sort(c.SelectedFields)
									}),
								app.Text(possibleField),
							)
						}),
					),
					app.Div().Body(
						app.Button().
							Text("Search").
							OnClick(func(ctx app.Context, e app.Event) {
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

								ctx.Dispatch(func(ctx app.Context) {
									slog.InfoContext(ctx.Context, "Dispatch: Setting persons", "persons", output.Persons)
									c.Persons = output.Persons
								})
								ctx.Defer(func(ctx app.Context) {
									slog.InfoContext(ctx.Context, "Defer: Persons should be set", "persons", c.Persons)

									//ctx.Update()
								})
							}),
						app.Button().
							Text("CSV").
							OnClick(func(ctx app.Context, e app.Event) {
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
					return app.Div().Text(c.Error)
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
				myui.Table[*downballotapi.Person]().
					Rows(c.Persons).
					Columns(columns).
					Render(),
			)
		}),
	)
}
