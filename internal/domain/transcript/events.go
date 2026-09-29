package transcript

import "time"

const (
	EventTypeTranscriptReady  = "media_file.transcript_ready.v1"
	EventTypeTranscriptFailed = "media_file.transcript_failed.v1"
)

type TranscriptReady struct {
	ID           string    `json:"eventId"`
	TranscriptID string    `json:"transcriptId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	Language     string    `json:"language,omitempty"`
	WordCount    int       `json:"wordCount"`
	SRTKey       string    `json:"srtKey"`
	ASSKey       string    `json:"assKey"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e TranscriptReady) EventID() string       { return e.ID }
func (e TranscriptReady) EventType() string     { return EventTypeTranscriptReady }
func (e TranscriptReady) AggregateID() string   { return e.MediaFileID }
func (e TranscriptReady) OccurredAt() time.Time { return e.Timestamp }

type TranscriptFailed struct {
	ID           string    `json:"eventId"`
	TranscriptID string    `json:"transcriptId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e TranscriptFailed) EventID() string       { return e.ID }
func (e TranscriptFailed) EventType() string     { return EventTypeTranscriptFailed }
func (e TranscriptFailed) AggregateID() string   { return e.MediaFileID }
func (e TranscriptFailed) OccurredAt() time.Time { return e.Timestamp }
