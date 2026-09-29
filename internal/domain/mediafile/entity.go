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
	Width            *int
	Height           *int
	FPS              *float64
	VideoCodec       string
	AudioCodec       string
	AudioSampleRate  *int
	AudioChannels    *int
	SizeBytes        int64
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	events []event.DomainEvent
}

// ProbeMetadata is the technical fingerprint produced by ffprobe.
type ProbeMetadata struct {
	DurationMs      int64
	Width           *int
	Height          *int
	FPS             *float64
	VideoCodec      string
	AudioCodec      string
	AudioSampleRate *int
	AudioChannels   *int
	SizeBytes       int64
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

func (m *MediaFile) MarkProbing() error {
	switch m.Status {
	case StatusUploaded:
		now := time.Now().UTC()
		m.Status = StatusProbing
		m.UpdatedAt = now
		m.recordEvent(MediaFileProbing{
			ID:          uuid.New().String(),
			MediaFileID: m.ID.String(),
			ProjectID:   m.ProjectID.String(),
			UserID:      m.UserID.String(),
			Status:      m.Status,
			Timestamp:   now,
		})
		return nil
	case StatusProbing:
		return nil
	default:
		return ErrInvalidTransition
	}
}

func (m *MediaFile) ApplyProbe(meta ProbeMetadata) error {
	if m.Status != StatusUploaded && m.Status != StatusProbing {
		return ErrInvalidTransition
	}
	if err := ValidateProbeMetadata(meta); err != nil {
		return err
	}

	now := time.Now().UTC()
	m.DurationMs = meta.DurationMs
	m.Width = meta.Width
	m.Height = meta.Height
	m.FPS = meta.FPS
	m.VideoCodec = meta.VideoCodec
	m.AudioCodec = meta.AudioCodec
	m.AudioSampleRate = meta.AudioSampleRate
	m.AudioChannels = meta.AudioChannels
	if meta.SizeBytes > 0 {
		m.SizeBytes = meta.SizeBytes
	}
	m.Status = StatusReady
	m.UpdatedAt = now
	m.recordEvent(MediaFileReady{
		ID:              uuid.New().String(),
		MediaFileID:     m.ID.String(),
		ProjectID:       m.ProjectID.String(),
		UserID:          m.UserID.String(),
		Status:          m.Status,
		DurationMs:      m.DurationMs,
		Width:           m.Width,
		Height:          m.Height,
		FPS:             m.FPS,
		VideoCodec:      m.VideoCodec,
		AudioCodec:      m.AudioCodec,
		AudioSampleRate: m.AudioSampleRate,
		AudioChannels:   m.AudioChannels,
		SizeBytes:       m.SizeBytes,
		Timestamp:       now,
	})
	return nil
}

func (m *MediaFile) MarkProbeFailed(reason string) error {
	if m.Status != StatusUploaded && m.Status != StatusProbing {
		return ErrInvalidTransition
	}
	if reason == "" {
		reason = ErrProbeInvalid.Error()
	}
	now := time.Now().UTC()
	m.Status = StatusFailed
	m.UpdatedAt = now
	m.recordEvent(MediaFileProbeFailed{
		ID:          uuid.New().String(),
		MediaFileID: m.ID.String(),
		ProjectID:   m.ProjectID.String(),
		UserID:      m.UserID.String(),
		Status:      m.Status,
		Reason:      reason,
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

// ValidateProbeMetadata ensures the probed file is a usable video source.
func ValidateProbeMetadata(meta ProbeMetadata) error {
	if meta.DurationMs <= 0 {
		return ErrProbeInvalid
	}
	if meta.VideoCodec == "" {
		return ErrProbeInvalid
	}
	if meta.Width == nil || meta.Height == nil || *meta.Width <= 0 || *meta.Height <= 0 {
		return ErrProbeInvalid
	}
	return nil
}
