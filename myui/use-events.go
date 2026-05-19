package myui

import (
	"reflect"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type HasEvents[T any] interface {
	On(name string, eventHandler app.EventHandler, options ...app.EventOption) T
}

type UseEvents struct {
	events []eventHandlerEvent
}

type eventHandlerEvent struct {
	name         string
	eventHandler app.EventHandler
	options      []app.EventOption
}

func (c *UseEvents) On(name string, eventHandler app.EventHandler, options ...app.EventOption) *UseEvents {
	c.events = append(c.events, eventHandlerEvent{
		name:         name,
		eventHandler: eventHandler,
		options:      options,
	})
	return c
}

func (c *UseEvents) Wrap(element app.UI) app.UI {
	func() {
		//slog.Debug("Wrap", "element", element)

		elementValue := reflect.ValueOf(element)
		//slog.Debug("Wrap", "elementValue", elementValue)

		// TODO: Can we remove this block?
		/*for elementValue.Kind() == reflect.Pointer {
			elementValue = elementValue.Elem()
		}
		slog.Info("Wrap", "elementValue", elementValue)
		if elementValue.Kind() != reflect.Struct {
			return
		}
		*/

		methodValue := elementValue.MethodByName("On")
		if methodValue.IsZero() {
			return
		}
		//slog.Debug("Wrap", "methodValue", methodValue)
		methodType := methodValue.Type()
		if methodType.NumIn() != 3 {
			return
		}
		in1 := methodType.In(0)
		if in1.Kind() != reflect.String {
			return
		}
		in2 := methodType.In(1)
		var eventHandler app.EventHandler
		if in2 != reflect.TypeOf(eventHandler) {
			return
		}
		in3 := methodType.In(2)
		var options []app.EventOption
		if in3 != reflect.TypeOf(options) {
			return
		}

		for _, event := range c.events {
			//slog.Debug("REGISTERING EVENT", "name", event.name)
			reflectOptions := []reflect.Value{
				reflect.ValueOf(event.name),
				reflect.ValueOf(event.eventHandler),
			}
			for _, option := range event.options {
				reflectOptions = append(reflectOptions, reflect.ValueOf(option))
			}
			methodValue.Call(reflectOptions)
		}
	}()
	return element
}
