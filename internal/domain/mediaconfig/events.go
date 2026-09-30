package mediaconfig

import "time"

const EventTypeConfigurationUpdated = "media_file.configuration_updated.v1"

// ConfigurationUpdated is emitted when editor configuration is saved and a timeline
// rebuild (or silence redetect) has been requested. Frontend can treat it like
// silence_detected: wait for media_file.timeline_updated then GET /editor.
type ConfigurationUpdated struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	JobID       string    `json:"jobId"`
	Reason      string    `json:"reason,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e ConfigurationUpdated) EventID() string       { return e.ID }
func (e ConfigurationUpdated) EventType() string     { return EventTypeConfigurationUpdated }
func (e ConfigurationUpdated) AggregateID() string   { return e.MediaFileID }
func (e ConfigurationUpdated) OccurredAt() time.Time { return e.Timestamp }
