package write

import (
	"time"

	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

type DetectedTranscriptIssueModel struct {
	ID             uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID    uuid.UUID `gorm:"column:media_file_id"`
	TranscriptID   uuid.UUID `gorm:"column:transcript_id"`
	IssueType      string    `gorm:"column:issue_type"`
	Text           string    `gorm:"column:text"`
	SourceStartMs  int64     `gorm:"column:source_start_ms"`
	SourceEndMs    int64     `gorm:"column:source_end_ms"`
	Confidence     *float64  `gorm:"column:confidence"`
	WordStartIndex int       `gorm:"column:word_start_index"`
	WordEndIndex   int       `gorm:"column:word_end_index"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (DetectedTranscriptIssueModel) TableName() string { return "detected_transcript_issue" }

func detectedTranscriptIssueModelFromDomain(i *domaintranscriptissue.Issue) *DetectedTranscriptIssueModel {
	return &DetectedTranscriptIssueModel{
		ID:             i.ID,
		MediaFileID:    i.MediaFileID,
		TranscriptID:   i.TranscriptID,
		IssueType:      i.Type,
		Text:           i.Text,
		SourceStartMs:  i.SourceStartMs,
		SourceEndMs:    i.SourceEndMs,
		Confidence:     i.Confidence,
		WordStartIndex: i.WordStartIndex,
		WordEndIndex:   i.WordEndIndex,
		CreatedAt:      i.CreatedAt,
	}
}
