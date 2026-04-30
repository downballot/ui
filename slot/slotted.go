package slot

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Slotted struct {
	components []app.UI
}

func (s *Slotted) AddSlotContents(components ...app.UI) *Slotted {
	s.components = append(s.components, components...)
	return s
}

func (s *Slotted) SlotContents() []app.UI {
	return s.components
}
