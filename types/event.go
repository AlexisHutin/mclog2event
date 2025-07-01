package types

// EventPayload is the structure for the event sent to the webhook.
type EventPayload struct {
	EventType string    `json:"event_type"`
	EventData map[string]string `json:"data"`
}
