package googlemap

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type HTMLGoogleMap struct {
	app.Compo

	id string

	APIKeyValue  string
	CenterValue  Coordinate
	MarkersValue []Marker
}

func GoogleMap() *HTMLGoogleMap {
	return &HTMLGoogleMap{
		id: "id-" + uuid.New().String(),
	}
}

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

type Marker struct {
	Coordinate
	Title   string
	Body    func() []app.UI
	OnClick func(ctx app.Context, event app.Event)
}

//go:embed dynamic-import.js
var rawDynamicImportCode string

func (c *HTMLGoogleMap) APIKey(value string) *HTMLGoogleMap {
	c.APIKeyValue = value
	return c
}

func (c *HTMLGoogleMap) Center(value Coordinate) *HTMLGoogleMap {
	c.CenterValue = value
	return c
}

func (c *HTMLGoogleMap) Markers(value []Marker) *HTMLGoogleMap {
	c.MarkersValue = value
	return c
}

func (c *HTMLGoogleMap) OnUpdate(ctx app.Context) {
	slog.InfoContext(ctx.Context, "GoogleMap: OnUpdate")

	ctx.Async(func() {
		slog.InfoContext(context.TODO(), "GoogleMap: OnUpdate")
		var bounds app.Value
		if v := app.Window().Get("google"); v.Truthy() {
			if v = v.Get("maps"); v.Truthy() {
				if v = v.Get("LatLngBounds"); v.Truthy() {
					bounds = v.New()
					slog.InfoContext(context.TODO(), "GoogleMap: OnUpdate", "bounds", bounds)
					for _, marker := range c.MarkersValue {
						bounds.Call("extend", app.Window().Get("google").Get("maps").Get("LatLng").New(marker.Latitude, marker.Longitude))
					}

					mapElement := app.Window().GetElementByID(c.id)
					if mapElement.Truthy() {
						if innerMap := mapElement.Get("innerMap"); innerMap.Truthy() {
							slog.InfoContext(context.TODO(), "GoogleMap: OnUpdate", "innerMap", innerMap)
							innerMap.Call("fitBounds", bounds)
						}
					}
				}
			}
		}
	})
}

func (c *HTMLGoogleMap) Render() app.UI {
	slog.InfoContext(context.TODO(), "GoogleMap: Render")

	var center app.Value
	var bounds app.Value
	center = app.ValueOf(fmt.Sprintf("%f,%f", c.CenterValue.Latitude, c.CenterValue.Longitude))
	if len(c.MarkersValue) > 0 {
		if v := app.Window().Get("google"); v.Truthy() {
			if v = v.Get("maps"); v.Truthy() {
				if v = v.Get("LatLngBounds"); v.Truthy() {
					bounds = v.New()
					slog.InfoContext(context.TODO(), "GoogleMap: bounds", "len(c.MarkersValue)", len(c.MarkersValue), "bounds", bounds)
					for _, marker := range c.MarkersValue {
						bounds.Call("extend", app.Window().Get("google").Get("maps").Get("LatLng").New(marker.Latitude, marker.Longitude))
					}
					newCenter := bounds.Call("getCenter")
					slog.InfoContext(context.TODO(), "GoogleMap: center", "center", center)

					center = app.ValueOf(fmt.Sprintf("%f,%f", newCenter.Call("lat").Float(), newCenter.Call("lng").Float()))
				}
			}
		}
	}

	return app.Div().
		Class("google-map").
		Style("width", "100%").
		Style("height", "100%").
		Body(
			app.Script().Text(strings.ReplaceAll(rawDynamicImportCode, "YOUR_API_KEY", c.APIKeyValue)),
			app.Elem("gmp-map").
				ID(c.id).
				Attr("center", center).
				Attr("zoom", "13").
				Attr("map-id", "MY_MAP_ID").
				Attr("gesture-handling", "none").
				Style("width", "100%").
				Style("height", "100%").
				Body(
					app.Range(c.MarkersValue).Slice(func(i int) app.UI {
						marker := c.MarkersValue[i]
						element := app.Elem("gmp-advanced-marker").
							Attr("position", fmt.Sprintf("%f,%f", marker.Latitude, marker.Longitude)).
							Attr("title", marker.Title).
							Attr("gmp-clickable", "true").
							DataSet("markerindex", fmt.Sprintf("%d", i)).
							On("gmp-click", func(ctx app.Context, event app.Event) {
								// For whatever reason, the event handler is not updated when the elemnt itself is updated.
								// In this case, we need to have a generic function that will work no matter what, so we're going
								// to store the marker index in the data set of the element itself, and then we'll ask for the current
								// element that was clicked on.
								markerIndexString := event.Get("target").Get("dataset").Get("markerindex").String()
								markerIndex, err := strconv.Atoi(markerIndexString)
								if err != nil {
									slog.ErrorContext(ctx.Context, "GoogleMap: Marker clicked; could not parse marker index", "error", err)
									ctx.PreventUpdate()
									return
								}
								slog.InfoContext(ctx.Context, "GoogleMap: Marker clicked", "markerIndex", markerIndex)
								marker := c.MarkersValue[markerIndex]
								slog.InfoContext(ctx.Context, "GoogleMap: Marker clicked", "marker", marker)
								if marker.OnClick == nil {
									ctx.PreventUpdate()
									return
								}
								marker.OnClick(ctx, event)
							})

						if marker.Body != nil {
							element = element.Body(marker.Body()...)
						}
						return element
					}),
				),
			app.Script().Text(`
async function init() {
    // Import the needed libraries.
	console.log("GoogleMap: importing libraries");
    await google.maps.importLibrary('maps');
	await google.maps.importLibrary('marker');

    // Access the map.
    const mapElement = document.querySelector('#`+c.id+`');
	console.log("GoogleMap: mapElement:", mapElement);
	if (!mapElement) {
		return
	}
    // Access the underlying map object.
    const innerMap = mapElement.innerMap;
	console.log("GoogleMap: innerMap:", innerMap);
	if (!innerMap) {
		return
	}
	innerMap.setOptions({
		gestureHandling: 'greedy',
	});
}

void init();
			`),
		)
}
