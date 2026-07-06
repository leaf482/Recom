package features

import (
	"fmt"

	"echorec/events"
)

func ScoreDelta(eventType string) (int, error) {
	switch eventType {
	case events.EventTypePlay:
		return 1, nil
	case events.EventTypeSkip:
		return -2, nil
	case events.EventTypeLike:
		return 4, nil
	case events.EventTypeSave:
		return 5, nil
	case events.EventTypeReplay:
		return 6, nil
	default:
		return 0, fmt.Errorf("unknown event type: %s", eventType)
	}
}
