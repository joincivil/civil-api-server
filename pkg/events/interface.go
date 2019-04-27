package events

import (
	"github.com/joincivil/civil-events-processor/pkg/processor"
)

// EventHandler is an interface to a governance event handler.
type EventHandler interface {
	// Handle runs the logic to handle the event as appropriate for the event
	// Returns a bool whether the event was handled and an error if occurred
	Handle(event *processor.PubSubMessage) (bool, error)
	// Name returns a readable name for this particular event handler
	Name() string
}
