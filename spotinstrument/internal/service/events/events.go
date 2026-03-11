package events

import "time"

var (
	MARKETS_UPDATE_EVENTS = "markets_update_event"
)

type MarketUpdateEvent struct {
	MarketUuid string
	UpdateAt   time.Time
}

func (e *MarketUpdateEvent) EventType() string {
	return MARKETS_UPDATE_EVENTS
}
