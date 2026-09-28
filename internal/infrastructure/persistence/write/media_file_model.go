package write

import (
	"time"

	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
)

type MediaFileModel struct {
	ID               uuid.UUID `gorm:"column:id;primaryKey"`
	ProjectID        uuid.UUID `gorm:"column:project_id"`
	UserID           uuid.UUID `gorm:"column:user_id"`
	StorageKey       string    `gorm:"column:storage_key"`
	ThumbnailKey     string    `gorm:"column:thumbnail_key"`
	OriginalFilename *string   `gorm:"column:original_filename"`
	MimeType         *string   `gorm:"column:mime_type"`
	DurationMs       int64     `gorm:"column:duration_ms"`
	Width            *int      `gorm:"column:width"`
	Height           *int      `gorm:"column:height"`
	FPS              *float64  `gorm:"column:fps"`
	VideoCodec       *string   `gorm:"column:video_codec"`
	AudioCodec       *string   `gorm:"column:audio_codec"`
	SizeBytes        *int64    `gorm:"column:size_bytes"`
	Status           string    `gorm:"column:status"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (MediaFileModel) TableName() string {
	return "media_file"
}

func mediaFileModelFromDomain(m *domainmediafile.MediaFile) *MediaFileModel {
	filename := m.OriginalFilename
	mimeType := m.MimeType
	sizeBytes := m.SizeBytes
	return &MediaFileModel{
		ID:               m.ID,
		ProjectID:        m.ProjectID,
		UserID:           m.UserID,
		StorageKey:       m.StorageKey,
		ThumbnailKey:     m.ThumbnailKey,
		OriginalFilename: &filename,
		MimeType:         &mimeType,
		DurationMs:       m.DurationMs,
		SizeBytes:        &sizeBytes,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func mediaFileDomainFromModel(m *MediaFileModel) *domainmediafile.MediaFile {
	media := &domainmediafile.MediaFile{
		ID:           m.ID,
		ProjectID:    m.ProjectID,
		UserID:       m.UserID,
		StorageKey:   m.StorageKey,
		ThumbnailKey: m.ThumbnailKey,
		DurationMs:   m.DurationMs,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.OriginalFilename != nil {
		media.OriginalFilename = *m.OriginalFilename
	}
	if m.MimeType != nil {
		media.MimeType = *m.MimeType
	}
	if m.SizeBytes != nil {
		media.SizeBytes = *m.SizeBytes
	}
	return media
}
