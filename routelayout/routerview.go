package routelayout

import "github.com/maxence-charriere/go-app/v10/pkg/app"

// RouterViewInterface is the interface that must be implemented in order for a component to be
// properly used in a route as a layout.
//
// Instead of implementing this interface, a component should embed `RouterViewComponent`.
type RouterViewInterface interface {
	RouterView() app.Composer
	SetRouterView(app.Composer)
}

// RouterViewComponent can be embedded in a layout component.
//
// The router wil set the specific router view component with `SetComponent`, and the
// embedding component can call `RouterViewComponent` in its `Render` function to get the component
// that should be rendered in the router view.
type RouterViewComponent struct {
	component app.Composer
}

var _ RouterViewInterface = (*RouterViewComponent)(nil)

// RouterView returns the router view comonent.  Put this where you wan the route component
// to be rendered.
//
// In Vue, this would be the `<router-view>` component.
func (v *RouterViewComponent) RouterView() app.Composer {
	return v.component
}

// SetRouterView is used by the router to set the component that will be returned by `RouterView`.
//
// This should not be called by anything but the router.
func (v *RouterViewComponent) SetRouterView(component app.Composer) {
	v.component = component
}
