package write

import (
	"time"

	domainmediaaudio "go-api/internal/domain/mediaaudio"

	"github.com/google/uuid"
)

type MediaAudioModel struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID uuid.UUID `gorm:"column:media_file_id"`
	ProjectID   uuid.UUID `gorm:"column:project_id"`
	UserID      uuid.UUID `gorm:"column:user_id"`
	StorageKey  string    `gorm:"column:storage_key"`
	Codec       *string   `gorm:"column:codec"`
	SampleRate  *int      `gorm:"column:sample_rate"`
	Channels    *int      `gorm:"column:channels"`
	SizeBytes   *int64    `gorm:"column:size_bytes"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (MediaAudioModel) TableName() string { return "media_audio" }

func mediaAudioModelFromDomain(a *domainmediaaudio.MediaAudio) *MediaAudioModel {
	codec := a.Codec
	sampleRate := a.SampleRate
	channels := a.Channels
	sizeBytes := a.SizeBytes
	model := &MediaAudioModel{
		ID:          a.ID,
		MediaFileID: a.MediaFileID,
		ProjectID:   a.ProjectID,
		UserID:      a.UserID,
		StorageKey:  a.StorageKey,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
	if a.Codec != "" {
		model.Codec = &codec
	}
	if a.SampleRate > 0 {
		model.SampleRate = &sampleRate
	}
	if a.Channels > 0 {
		model.Channels = &channels
	}
	if a.SizeBytes > 0 {
		model.SizeBytes = &sizeBytes
	}
	return model
}

func mediaAudioDomainFromModel(m *MediaAudioModel) *domainmediaaudio.MediaAudio {
	audio := &domainmediaaudio.MediaAudio{
		ID:          m.ID,
		MediaFileID: m.MediaFileID,
		ProjectID:   m.ProjectID,
		UserID:      m.UserID,
		StorageKey:  m.StorageKey,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.Codec != nil {
		audio.Codec = *m.Codec
	}
	if m.SampleRate != nil {
		audio.SampleRate = *m.SampleRate
	}
	if m.Channels != nil {
		audio.Channels = *m.Channels
	}
	if m.SizeBytes != nil {
		audio.SizeBytes = *m.SizeBytes
	}
	return audio
}
