package page

import (
	"github.com/downballot/ui/go-ui/component/customlayout"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func init() {
	app.Route("/", func() app.Composer { return &customlayout.DownballotLayout{} })
}
