package mediafile

import (
	"path/filepath"
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type MediaFile struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	UserID           uuid.UUID
	OriginalFilename string
	StorageKey       string
	ThumbnailKey     string
	MimeType         string
	DurationMs       int64
	SizeBytes        int64
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	events []event.DomainEvent
}

func NewPresignedMediaFile(
	projectID, userID uuid.UUID,
	filename, contentType string,
	sizeBytes int64,
) (*MediaFile, error) {
	filename = SanitizeFilename(filename)
	if filename == "" || filename == "." || filename == string(filepath.Separator) {
		return nil, ErrInvalidFilename
	}
	if !IsVideoFilename(filename) && !IsVideoContentType(contentType) {
		return nil, ErrUnsupportedType
	}

	now := time.Now().UTC()
	id := uuid.New()
	m := &MediaFile{
		ID:               id,
		ProjectID:        projectID,
		UserID:           userID,
		OriginalFilename: filename,
		StorageKey:       NewStorageKey(id),
		MimeType:         ContentTypeFromFilename(filename, contentType),
		SizeBytes:        sizeBytes,
		Status:           StatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.recordCreated()
	return m, nil
}

func (m *MediaFile) PullEvents() []event.DomainEvent {
	events := m.events
	m.events = nil
	return events
}

func (m *MediaFile) recordEvent(e event.DomainEvent) {
	m.events = append(m.events, e)
}

func (m *MediaFile) recordCreated() {
	m.recordEvent(MediaFileCreated{
		ID:          uuid.New().String(),
		MediaFileID: m.ID.String(),
		ProjectID:   m.ProjectID.String(),
		UserID:      m.UserID.String(),
		Filename:    m.OriginalFilename,
		StorageKey:  m.StorageKey,
		Status:      m.Status,
		Timestamp:   m.CreatedAt,
	})
}

func (m *MediaFile) MarkUploaded(contentType string, sizeBytes int64) error {
	if contentType != "" {
		m.MimeType = contentType
	}
	if sizeBytes > 0 {
		m.SizeBytes = sizeBytes
	}

	switch m.Status {
	case StatusPending:
		now := time.Now().UTC()
		m.Status = StatusUploaded
		m.UpdatedAt = now
		m.recordEvent(MediaFileUploaded{
			ID:          uuid.New().String(),
			MediaFileID: m.ID.String(),
			ProjectID:   m.ProjectID.String(),
			UserID:      m.UserID.String(),
			StorageKey:  m.StorageKey,
			ContentType: m.MimeType,
			SizeBytes:   m.SizeBytes,
			Status:      m.Status,
			Timestamp:   now,
		})
		return nil
	case StatusUploaded, StatusProbing, StatusReady, StatusProcessing, StatusCompleted:
		m.UpdatedAt = time.Now().UTC()
		return nil
	default:
		return ErrInvalidTransition
	}
}

func (m *MediaFile) MarkUploadExpired() error {
	if m.Status != StatusPending {
		return ErrInvalidTransition
	}
	now := time.Now().UTC()
	m.Status = StatusFailed
	m.UpdatedAt = now
	m.recordEvent(MediaFileUploadExpired{
		ID:          uuid.New().String(),
		MediaFileID: m.ID.String(),
		ProjectID:   m.ProjectID.String(),
		UserID:      m.UserID.String(),
		Timestamp:   now,
	})
	return nil
}

func (m *MediaFile) SetThumbnailKey(key string) {
	if key == "" || m.ThumbnailKey == key {
		return
	}
	now := time.Now().UTC()
	m.ThumbnailKey = key
	m.UpdatedAt = now
	m.recordEvent(MediaFileThumbnailReady{
		ID:           uuid.New().String(),
		MediaFileID:  m.ID.String(),
		ProjectID:    m.ProjectID.String(),
		UserID:       m.UserID.String(),
		ThumbnailKey: key,
		Status:       m.Status,
		Timestamp:    now,
	})
}
