package viral

import "time"

const EventTypeViralReady = "media_file.viral_ready.v1"

type ViralReady struct {
	ID             string    `json:"eventId"`
	MediaFileID    string    `json:"mediaFileId"`
	ProjectID      string    `json:"projectId"`
	UserID         string    `json:"userId"`
	CandidateCount int       `json:"candidateCount"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ViralReady) EventID() string       { return e.ID }
func (e ViralReady) EventType() string     { return EventTypeViralReady }
func (e ViralReady) AggregateID() string   { return e.MediaFileID }
func (e ViralReady) OccurredAt() time.Time { return e.Timestamp }
