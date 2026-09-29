package job

import "time"

const (
	EventTypeJobCreated       = "job.created.v1"
	EventTypeJobStatusChanged = "job.status_changed.v1"
)

type JobCreated struct {
	ID          string    `json:"eventId"`
	JobID       string    `json:"jobId"`
	ProjectID   string    `json:"projectId"`
	MediaFileID string    `json:"mediaFileId"`
	UserID      string    `json:"userId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e JobCreated) EventID() string       { return e.ID }
func (e JobCreated) EventType() string     { return EventTypeJobCreated }
func (e JobCreated) AggregateID() string   { return e.JobID }
func (e JobCreated) OccurredAt() time.Time { return e.Timestamp }

type JobStatusChanged struct {
	ID           string    `json:"eventId"`
	JobID        string    `json:"jobId"`
	ProjectID    string    `json:"projectId"`
	MediaFileID  string    `json:"mediaFileId"`
	UserID       string    `json:"userId"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e JobStatusChanged) EventID() string       { return e.ID }
func (e JobStatusChanged) EventType() string     { return EventTypeJobStatusChanged }
func (e JobStatusChanged) AggregateID() string   { return e.JobID }
func (e JobStatusChanged) OccurredAt() time.Time { return e.Timestamp }
