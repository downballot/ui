package slot

import "github.com/maxence-charriere/go-app/v11/pkg/app"

// Slotted can be used as an embedded component to add slot contents to a component.
type Slotted struct {
	components []app.UI
}

// AddSlotContents adds the given components to the slot.
//
// For example, call this within your component's Body method.
func (s *Slotted) AddSlotContents(components ...app.UI) *Slotted {
	s.components = components
	return s
}

// SlotContents returns the non-nil components in the slot, suitable for use in a parent component's Body method.
//
// For example, call this within your component's Render method.
func (s *Slotted) SlotContents() []app.UI {
	return app.FilterUIElems(s.components...)
}
