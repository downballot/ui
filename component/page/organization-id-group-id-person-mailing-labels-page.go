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
	"github.com/downballot/ui/component"
	"github.com/downballot/ui/myui"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type OrganizationIDGroupIDPersonMailingLabelsPage struct {
	app.Compo

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

	FilterOpen bool
	FormatOpen bool

	Format string

	Addresses []string
}

var _ app.Navigator = (*OrganizationIDGroupIDPersonMailingLabelsPage)(nil)

func (c *OrganizationIDGroupIDPersonMailingLabelsPage) OnNav(ctx app.Context) {
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonMailingLabelsPage: OnNav")

	router.GetActiveRoute(ctx).ReadVariable("organization_id", &c.organizationID)
	router.GetActiveRoute(ctx).ReadVariable("group_id", &c.groupID)

	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonMailingLabelsPage: OnNav", "OrganizationID", c.organizationID)
	slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonMailingLabelsPage: OnNav", "GroupID", c.groupID)

	if c.organizationID == "" {
		return
	}

	if c.groupID == "" {
		return
	}

	if c.Format == "" {
		c.Format = "5164"
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

			c.PossibleFields = possibleFields
			slog.InfoContext(ctx.Context, "OrganizationIDGroupIDPersonMailingLabelsPage: OnNav: Async", "len(PossibleFields)", len(c.PossibleFields))
		})
		wg.Wait()

		ctx.Dispatch(func(ctx app.Context) {
			c.loaded = true

			slog.InfoContext(ctx.Context, "Dispatch: Loading complete.  Searching for persons.")
			c.search(ctx)
		})
	})
}

func (c *OrganizationIDGroupIDPersonMailingLabelsPage) Render() app.UI {
	slog.InfoContext(context.TODO(), "OrganizationIDGroupIDPersonMailingLabelsPage: Render")

	var allPossibleFilterStrings []string
	for _, filter := range c.Filters {
		allPossibleFilterStrings = append(allPossibleFilterStrings, filter.Filter)
	}
	slices.Sort(allPossibleFilterStrings)

	if !c.loaded {
		return myui.Page().Body(
			app.Div().Text("Loading..."),
		)
	}

	if c.group == nil {
		return myui.Page().Body(
			myui.StatusBar().
				Text("Not found").
				Bad(),
		)
	}

	return app.Div().
		Body(
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
								AllowedValue(func() []myui.SelectOption {
									var allowedValues []myui.SelectOption
									allowedValues = append(allowedValues, myui.SelectOption{Label: "Select a filter or create your own", Value: ""})
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
			myui.Collapse().
				Label("Format").
				Bind(&c.FormatOpen).
				SummaryText(c.Format).
				Body(
					myui.Select().
						Label("Format").
						AllowedValue(
							myui.SelectOption{Label: "Select a format", Value: "", Disabled: true},
							myui.SelectOption{Label: "5164", Value: "5164"},
						).
						Bind(&c.Format),
				),
			myui.Form().
				Class("no-print").
				Spacer(false).
				Action(
					myui.FormAction{
						Name:     "Search",
						Icon:     "search",
						Function: c.search,
					},
				),
			app.If(c.Error != "", func() app.UI {
				return myui.StatusBar().
					Text(c.Error).
					Bad()
			}),
			component.MailingLabels().
				Format(c.Format).
				ReturnAddress("Downballot IO\nEaton Place\nNewark, DE 19711").
				Addresses(c.Addresses),
		)
}

func (c *OrganizationIDGroupIDPersonMailingLabelsPage) search(ctx app.Context) {
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

	c.Addresses = []string{}
	for _, person := range c.Persons {
		finalAddress := person.Fields["name"]

		{
			var addressLines []string

			lines := strings.Split(person.Fields["residential_address"], ",")
			for len(lines) > 0 {
				if len(lines) == 2 {
					addressLine := lines[0] + ", " + lines[1]
					lines = lines[2:]

					addressLines = append(addressLines, addressLine)
				} else {
					addressLine := lines[0]
					lines = lines[1:]

					addressLines = append(addressLines, addressLine)
				}
			}

			for _, addressLine := range addressLines {
				finalAddress += "\n" + addressLine
			}
		}

		c.Addresses = append(c.Addresses, finalAddress)
	}

	ctx.Update()
}
