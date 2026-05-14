package googlemap

import (
	"context"
	_ "embed"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type HTMLGoogleMap struct {
	app.Compo

	IDValue string

	APIKeyValue string
	CenterValue *Coordinate
}

func GoogleMap() *HTMLGoogleMap {
	return &HTMLGoogleMap{
		IDValue: "id-" + uuid.New().String(),
	}
}

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

//go:embed dynamic-import.js
var rawDynamicImportCode string

func (c *HTMLGoogleMap) APIKey(value string) *HTMLGoogleMap {
	c.APIKeyValue = value
	return c
}

func (c *HTMLGoogleMap) Center(value *Coordinate) *HTMLGoogleMap {
	c.CenterValue = value
	return c
}

func (c *HTMLGoogleMap) Render() app.UI {
	slog.InfoContext(context.TODO(), "GoogleMap: Render")

	return app.Div().
		Class("google-map").
		Style("width", "100%").
		Style("height", "100%").
		Body(
			app.Script().Text(strings.ReplaceAll(rawDynamicImportCode, "YOUR_API_KEY", c.APIKeyValue)),
			app.Elem("gmp-map").ID(c.IDValue),
			app.Script().Text(`
async function init() {
    // Import the needed libraries.
	console.log("GoogleMap: importing libraries");
    await google.maps.importLibrary('maps');

    // Access the map.
    const mapElement = document.querySelector('#`+c.IDValue+`');
	console.log("GoogleMap: mapElement:", mapElement);
    // Access the underlying map object.
    const innerMap = mapElement.innerMap;
	console.log("GoogleMap: innerMap:", innerMap);
}

void init();
			`),
		)
}
