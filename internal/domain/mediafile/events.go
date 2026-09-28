package mediafile

import "time"

const (
	EventTypeMediaFileCreated       = "media_file.created.v1"
	EventTypeMediaFileUploaded      = "media_file.uploaded.v1"
	EventTypeMediaFileUploadExpired = "media_file.upload_expired.v1"
)

type MediaFileCreated struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Filename    string    `json:"filename"`
	StorageKey  string    `json:"storageKey"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileCreated) EventID() string       { return e.ID }
func (e MediaFileCreated) EventType() string     { return EventTypeMediaFileCreated }
func (e MediaFileCreated) AggregateID() string   { return e.MediaFileID }
func (e MediaFileCreated) OccurredAt() time.Time { return e.Timestamp }

type MediaFileUploaded struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	StorageKey  string    `json:"storageKey"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileUploaded) EventID() string       { return e.ID }
func (e MediaFileUploaded) EventType() string     { return EventTypeMediaFileUploaded }
func (e MediaFileUploaded) AggregateID() string   { return e.MediaFileID }
func (e MediaFileUploaded) OccurredAt() time.Time { return e.Timestamp }

type MediaFileUploadExpired struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileUploadExpired) EventID() string       { return e.ID }
func (e MediaFileUploadExpired) EventType() string     { return EventTypeMediaFileUploadExpired }
func (e MediaFileUploadExpired) AggregateID() string   { return e.MediaFileID }
func (e MediaFileUploadExpired) OccurredAt() time.Time { return e.Timestamp }
