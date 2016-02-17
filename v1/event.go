package payjp

import (
	"encoding/json"
	"time"
)

type Event struct {
	CreatedEpoch    int             `json:"created"`
	Data            json.RawMessage `json:"data"`
	ID              string          `json:"id"`
	LiveMode        bool            `json:"livemode"`
	Object          string          `json:"object"`
	PendingWebHooks int             `json:"pending_webhooks"`
	Type            string          `json:"type"`

	CreatedAt time.Time
}

type event Event

func (e *Event) UnmarshalJSON(b []byte) error {
	raw := event{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "event" {
		*e = Event(raw)
		e.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
