package payjp

import (
	"encoding/json"
	"time"
)

type SubscriptionResponse struct {
	CanceledEpoch           int    `json:"canceled_at"`
	CreatedEpoch            int    `json:"created"`
	CurrentPeriodEndEpoch   int    `json:"current_period_end"`
	CurrentPeriodStartEpoch int    `json:"current_period_start"`
	Customer                string `json:"customer"`
	Id                      string `json:"id"`
	LiveMode                bool   `json:"livemode"`
	Object                  string `json:"object"`
	PausedEpoch             int    `json:"paused_at"`
	Plan                    Plan   `json:"plan"`
	Prorate                 bool   `json:"prorate"`
	ResumedEpoch            int    `json:"resumed_at"`
	Start                   int    `json:"start"`
	Status                  string `json:"status"`
	TrialEndEpoch           int    `json:"trial_end"`
	TrialStartEpoch         int    `json:"trial_start"`

	CanceledAt           time.Time
	CreatedAt            time.Time
	CurrentPeriodEndAt   time.Time
	CurrentPeriodStartAt time.Time
	PausedAt             time.Time
	ResumedAt            time.Time
	TrialEndAt           time.Time
	TrialStartAt         time.Time

	service *Service
}

type subscription SubscriptionResponse

func (s *SubscriptionResponse) UnmarshalJSON(b []byte) error {
	raw := subscription{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "subscription" {
		*s = SubscriptionResponse(raw)
		s.CanceledAt = time.Unix(int64(raw.CanceledEpoch), 0)
		s.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		s.CurrentPeriodEndAt = time.Unix(int64(raw.CurrentPeriodEndEpoch), 0)
		s.CurrentPeriodStartAt = time.Unix(int64(raw.CurrentPeriodStartEpoch), 0)
		s.PausedAt = time.Unix(int64(raw.PausedEpoch), 0)
		s.ResumedAt = time.Unix(int64(raw.ResumedEpoch), 0)
		s.TrialEndAt = time.Unix(int64(raw.TrialEndEpoch), 0)
		s.TrialStartAt = time.Unix(int64(raw.TrialStartEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type SubscriptionList struct {
	Count   int                     `json:"count"`
	Data    []*SubscriptionResponse `json:"data"`
	HasMore bool                    `json:"has_more"`
	Object  string                  `json:"object"`
	URL     string                  `json:"url"`
}
