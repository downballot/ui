package page

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component"
	"github.com/go-app-blazar/blazar/blazar"
	"github.com/go-app-blazar/blazar/deref"
	"github.com/go-app-blazar/blazar/htmlevent"
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
	persons []*downballotapi.Person
	person  *downballotapi.Person
	Filters []*downballotapi.Filter

	PersonsTableColumns []blazar.TableColumn[*downballotapi.Person]

	filterOpen  bool
	detailsOpen bool

	lastCalledVoterID string
	newNotes          string
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

	var personItems []app.UI
	if c.person != nil {
		icon := "circle-question"
		switch c.person.Fields["candidate.support"] {
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
						Text(c.person.Fields["name"]),
				),
			app.Div().
				Class("person-summary-registration").
				Text("Registered as " + c.person.Fields["political_party"] + ", living in " + c.person.Fields["district_representative"] + ", " + c.person.Fields["district_senate"]),
			app.Div().
				Class("person-summary-address").
				Body(
					app.A().
						Href(fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", url.QueryEscape(c.person.Fields["residential_address"]))).
						Target("_blank").
						Text(c.person.Fields["residential_address"]),
				),
			app.Div().
				Class("person-summary-phone").
				Text(c.person.Fields["phone_number"]),
			app.Div().
				Class("person-summary-chips").
				Body(
					app.If(c.person.Fields["candidate.connected"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("handshake"),
								app.Text("Connected"),
							)
					}),
					app.If(c.person.Fields["candidate.support"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Support: " + c.person.Fields["candidate.support"])
					}),
					app.If(c.person.Fields["candidate.cat"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("cat"),
								app.Text("Cat"),
							)
					}),
					app.If(c.person.Fields["candidate.dog"] == "true", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Body(
								blazar.Icon().
									Icon("dog"),
								app.Text("Dog"),
							)
					}),
					app.If(c.person.Fields["candidate.date_called"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Called on " + c.person.Fields["candidate.date_called"])
					}),
					app.If(c.person.Fields["candidate.date_canvassed"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Canvassed on " + c.person.Fields["candidate.date_canvassed"])
					}),
					app.If(c.person.Fields["candidate.date_texted"] != "", func() app.UI {
						return app.Div().
							Class("person-summary-chip").
							Text("Texted on " + c.person.Fields["candidate.date_texted"])
					}),
				),
		}

		personItems = append(personItems,
			app.Div().
				Class("person-summary").
				Body(summaryItems...),
			app.If(c.person.Fields["candidate.notes"] != "", func() app.UI {
				return app.Hr()
			}),
			app.If(c.person.Fields["candidate.notes"] != "", func() app.UI {
				return app.Div().
					Class("person-summary-notes").
					Text(c.person.Fields["candidate.notes"])
			}),
			app.Hr(),
			app.Div().
				Style("display", "flex").
				Style("flex-direction", "row").
				Style("gap", "2em").
				Style("font-size", "1.5em").
				Body(
					blazar.Button().
						Label("Call").
						Icon(component.IconPhone).
						To("tel:+1"+strings.ReplaceAll(c.person.Fields["phone_number"], "-", "")).
						On("click", func(ctx app.Context, e app.Event) {
							c.lastCalledVoterID = c.person.VoterID
						}),
					blazar.Copy().
						Style("font-size", "1.2em").
						Style("margin-top", "auto").
						Style("margin-bottom", "auto").
						Text(c.person.Fields["phone_number"]).
						Value(c.person.Fields["phone_number"]).
						OnClick(func(ctx app.Context, e htmlevent.PointerEvent) {
							c.lastCalledVoterID = c.person.VoterID
						}),
				),
			app.Hr(),
			app.If(c.lastCalledVoterID == c.person.VoterID, func() app.UI {
				return blazar.Form().
					Class("no-print").
					Style("font-size", "1.2em").
					Spacer(false).
					Action(
						blazar.FormAction{
							Name:            "Bad Number",
							Icon:            component.IconDelete,
							BackgroundColor: "red",
							Function: func(ctx app.Context) {
								result := app.Window().Call("confirm", "Are you sure you want to mark this as a bad number?")
								slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: Bad Number button clicked", "result", result.Bool())
								if !result.Bool() {
									slog.InfoContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: Bad number button clicked: User cancelled", "result", result.Bool())
									return
								}

								phoneNumber := c.person.Fields["phone_number"]
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
							Name:            "Called (continue)",
							Icon:            component.IconDone,
							BackgroundColor: "green",
							Function: func(ctx app.Context) {
								phoneNumber := c.person.Fields["phone_number"]
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
					)
			}),
			app.If(c.lastCalledVoterID == c.person.VoterID, func() app.UI {
				return blazar.Collapse().
					Label("Additional Details").
					Style("margin-top", "1em").
					Bind(&c.detailsOpen).
					Body(
						app.FieldSet().
							Body(
								app.Legend().Text("Information"),
								blazar.Form().
									Class("no-print").
									Spacer(false).
									Body(
										app.Text("For more details, click the link below."),
									).
									Action(
										blazar.FormAction{
											Name:   "Edit Person",
											Icon:   component.IconExternalLink,
											To:     fmt.Sprintf("/organization/%s/person/%s", c.organizationID, c.person.VoterID),
											Target: "_blank",
										},
									),
							),
						app.FieldSet().
							Body(
								app.Legend().Text("Fields"),
								blazar.InputWrapper().
									Label("Connected (saves automatically)").
									Body(
										blazar.Form().
											Class("no-print").
											Style("font-size", "1.2em").
											Spacer(false).
											Action(
												blazar.FormAction{
													Name: "Yes, we spoke with this person",
													Flat: c.person.Fields["candidate.connected"] != "true",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.connected"] == "true" {
															c.updatePerson(ctx, "candidate.connected", nil)
														} else {
															newValue := "true"
															c.updatePerson(ctx, "candidate.connected", &newValue)
														}
													},
												},
											),
									),
								blazar.InputWrapper().
									Label("Support (saves automatically)").
									Body(
										blazar.Form().
											Class("no-print").
											Style("font-size", "1.2em").
											Spacer(false).
											Action(
												blazar.FormAction{
													Name: "-2",
													Flat: c.person.Fields["candidate.support"] != "-2",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.support"] == "-2" {
															c.updatePerson(ctx, "candidate.support", nil)
														} else {
															newValue := "-2"
															c.updatePerson(ctx, "candidate.support", &newValue)
														}
													},
												},
												blazar.FormAction{
													Name: "-1",
													Flat: c.person.Fields["candidate.support"] != "-1",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.support"] == "-1" {
															c.updatePerson(ctx, "candidate.support", nil)
														} else {
															newValue := "-1"
															c.updatePerson(ctx, "candidate.support", &newValue)
														}
													},
												},
												blazar.FormAction{
													Name: "0",
													Flat: c.person.Fields["candidate.support"] != "0",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.support"] == "0" {
															c.updatePerson(ctx, "candidate.support", nil)
														} else {
															newValue := "0"
															c.updatePerson(ctx, "candidate.support", &newValue)
														}
													},
												},
												blazar.FormAction{
													Name: "+1",
													Flat: c.person.Fields["candidate.support"] != "+1",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.support"] == "+1" {
															c.updatePerson(ctx, "candidate.support", nil)
														} else {
															newValue := "+1"
															c.updatePerson(ctx, "candidate.support", &newValue)
														}
													},
												},
												blazar.FormAction{
													Name: "+2",
													Flat: c.person.Fields["candidate.support"] != "+2",
													Function: func(ctx app.Context) {
														if c.person.Fields["candidate.support"] == "+2" {
															c.updatePerson(ctx, "candidate.support", nil)
														} else {
															newValue := "+2"
															c.updatePerson(ctx, "candidate.support", &newValue)
														}
													},
												},
											),
									),
								blazar.Input[string]().
									Label("Notes").
									Bind(&c.newNotes),
								blazar.Button().
									Label("Save Notes").
									Icon(component.IconSave).
									On("click", func(ctx app.Context, e app.Event) {
										c.updatePerson(ctx, "candidate.notes", &c.newNotes)
									}),
							),
					)
			}),
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
	c.persons = output.Persons

	rand.Shuffle(len(c.persons), func(i, j int) {
		c.persons[i], c.persons[j] = c.persons[j], c.persons[i]
	})

	if len(c.persons) == 0 {
		c.person = nil
	} else {
		c.person = c.persons[0]
	}

	c.newNotes = c.person.Fields["candidate.notes"]

	ctx.Update()
}

func (c *OrganizationIDGroupIDCalltimePage) updatePerson(ctx app.Context, field string, value *string) {
	if c.person == nil {
		return
	}

	input := downballotapi.PatchPersonRequest{
		Fields: map[string]*string{
			field: value,
		},
	}
	var output downballotapi.GetPersonResponse
	err := api.Do(ctx, http.MethodPatch, "/api/v1/organization/"+c.organizationID+"/person/"+c.person.VoterID, input, &output)
	if err != nil {
		slog.ErrorContext(ctx.Context, "OrganizationIDGroupIDCalltimePage: updatePerson: Could not patch person", "err", err)
		return
	}

	c.person = output.Person

	c.newNotes = c.person.Fields["candidate.notes"]

	ctx.Update()
}
