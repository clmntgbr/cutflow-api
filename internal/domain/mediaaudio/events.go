package mediaaudio

import "time"

const (
	EventTypeMediaAudioReady  = "media_file.audio_ready.v1"
	EventTypeMediaAudioFailed = "media_file.audio_failed.v1"
)

type MediaAudioReady struct {
	ID           string    `json:"eventId"`
	MediaAudioID string    `json:"mediaAudioId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	StorageKey   string    `json:"storageKey"`
	Codec        string    `json:"codec"`
	SampleRate   int       `json:"sampleRate"`
	Channels     int       `json:"channels"`
	SizeBytes    int64     `json:"sizeBytes"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e MediaAudioReady) EventID() string       { return e.ID }
func (e MediaAudioReady) EventType() string     { return EventTypeMediaAudioReady }
func (e MediaAudioReady) AggregateID() string   { return e.MediaFileID }
func (e MediaAudioReady) OccurredAt() time.Time { return e.Timestamp }

type MediaAudioFailed struct {
	ID           string    `json:"eventId"`
	MediaAudioID string    `json:"mediaAudioId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e MediaAudioFailed) EventID() string       { return e.ID }
func (e MediaAudioFailed) EventType() string     { return EventTypeMediaAudioFailed }
func (e MediaAudioFailed) AggregateID() string   { return e.MediaFileID }
func (e MediaAudioFailed) OccurredAt() time.Time { return e.Timestamp }
