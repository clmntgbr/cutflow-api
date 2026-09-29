package transcriptissue

import "time"

const EventTypeTranscriptAnalysisReady = "media_file.transcript_analysis_ready.v1"

type TranscriptAnalysisReady struct {
	ID              string    `json:"eventId"`
	MediaFileID     string    `json:"mediaFileId"`
	ProjectID       string    `json:"projectId"`
	UserID          string    `json:"userId"`
	TranscriptID    string    `json:"transcriptId"`
	FillerCount     int       `json:"fillerCount"`
	RepetitionCount int       `json:"repetitionCount"`
	FalseStartCount int       `json:"falseStartCount"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e TranscriptAnalysisReady) EventID() string       { return e.ID }
func (e TranscriptAnalysisReady) EventType() string     { return EventTypeTranscriptAnalysisReady }
func (e TranscriptAnalysisReady) AggregateID() string   { return e.MediaFileID }
func (e TranscriptAnalysisReady) OccurredAt() time.Time { return e.Timestamp }
