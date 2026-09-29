package silence

import "time"

const EventTypeSilenceDetected = "media_file.silence_detected.v1"

type SilenceDetected struct {
	ID           string    `json:"eventId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	SilenceCount int       `json:"silenceCount"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e SilenceDetected) EventID() string       { return e.ID }
func (e SilenceDetected) EventType() string     { return EventTypeSilenceDetected }
func (e SilenceDetected) AggregateID() string   { return e.MediaFileID }
func (e SilenceDetected) OccurredAt() time.Time { return e.Timestamp }
