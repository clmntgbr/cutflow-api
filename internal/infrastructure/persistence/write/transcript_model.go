package write

import (
	"time"

	domaintranscript "go-api/internal/domain/transcript"

	"github.com/google/uuid"
)

type TranscriptModel struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	ProjectID     uuid.UUID `gorm:"column:project_id"`
	UserID        uuid.UUID `gorm:"column:user_id"`
	Language      *string   `gorm:"column:language"`
	Text          *string   `gorm:"column:text"`
	SRTStorageKey string    `gorm:"column:srt_storage_key"`
	ASSStorageKey string    `gorm:"column:ass_storage_key"`
	ProviderJobID *string   `gorm:"column:provider_job_id"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (TranscriptModel) TableName() string { return "transcript" }

type TranscriptWordModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TranscriptID  uuid.UUID `gorm:"column:transcript_id"`
	WordIndex     int       `gorm:"column:word_index"`
	Text          string    `gorm:"column:text"`
	SourceStartMs int64     `gorm:"column:source_start_ms"`
	SourceEndMs   int64     `gorm:"column:source_end_ms"`
	Confidence    *float64  `gorm:"column:confidence"`
	Kind          string    `gorm:"column:kind"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (TranscriptWordModel) TableName() string { return "transcript_word" }

func transcriptModelFromDomain(t *domaintranscript.Transcript) *TranscriptModel {
	model := &TranscriptModel{
		ID:            t.ID,
		MediaFileID:   t.MediaFileID,
		ProjectID:     t.ProjectID,
		UserID:        t.UserID,
		SRTStorageKey: t.SRTStorageKey,
		ASSStorageKey: t.ASSStorageKey,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
	if t.Language != "" {
		lang := t.Language
		model.Language = &lang
	}
	if t.Text != "" {
		text := t.Text
		model.Text = &text
	}
	if t.ProviderJobID != "" {
		job := t.ProviderJobID
		model.ProviderJobID = &job
	}
	return model
}

func transcriptDomainFromModel(m *TranscriptModel) *domaintranscript.Transcript {
	t := &domaintranscript.Transcript{
		ID:            m.ID,
		MediaFileID:   m.MediaFileID,
		ProjectID:     m.ProjectID,
		UserID:        m.UserID,
		SRTStorageKey: m.SRTStorageKey,
		ASSStorageKey: m.ASSStorageKey,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.Language != nil {
		t.Language = *m.Language
	}
	if m.Text != nil {
		t.Text = *m.Text
	}
	if m.ProviderJobID != nil {
		t.ProviderJobID = *m.ProviderJobID
	}
	return t
}
