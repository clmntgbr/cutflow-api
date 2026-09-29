package timeline

import "time"

const (
	EventTypeTimelineRebuildRequested = "media_file.timeline_rebuild_requested.v1"
	EventTypeTimelineUpdated          = "media_file.timeline_updated.v1"
	EventTypeTimelineFailed           = "media_file.timeline_failed.v1"
)

type RebuildRequested struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	JobID       string    `json:"jobId"`
	Reason      string    `json:"reason,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e RebuildRequested) EventID() string       { return e.ID }
func (e RebuildRequested) EventType() string     { return EventTypeTimelineRebuildRequested }
func (e RebuildRequested) AggregateID() string   { return e.MediaFileID }
func (e RebuildRequested) OccurredAt() time.Time { return e.Timestamp }

type Updated struct {
	ID          string    `json:"eventId"`
	TimelineID  string    `json:"timelineId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Version     int       `json:"version"`
	DurationMs  int64     `json:"durationMs"`
	Fingerprint string    `json:"fingerprint"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e Updated) EventID() string       { return e.ID }
func (e Updated) EventType() string     { return EventTypeTimelineUpdated }
func (e Updated) AggregateID() string   { return e.MediaFileID }
func (e Updated) OccurredAt() time.Time { return e.Timestamp }

type Failed struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	JobID       string    `json:"jobId"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e Failed) EventID() string       { return e.ID }
func (e Failed) EventType() string     { return EventTypeTimelineFailed }
func (e Failed) AggregateID() string   { return e.MediaFileID }
func (e Failed) OccurredAt() time.Time { return e.Timestamp }
