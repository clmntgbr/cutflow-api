package mediaaudio

import (
	"context"
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusQueued     = "queued"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"

	ExtractedCodec      = "opus"
	ExtractedSampleRate = 16000
	ExtractedChannels   = 1
	ExtractedBitrate    = "32k"
	ExtractedObjectName = "audio.opus"
	ExtractedContentType = "audio/ogg"
)

type MediaAudio struct {
	ID          uuid.UUID
	MediaFileID uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	StorageKey  string
	Codec       string
	SampleRate  int
	Channels    int
	SizeBytes   int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	events []event.DomainEvent
}

func NewPending(mediaFileID, projectID, userID uuid.UUID) *MediaAudio {
	now := time.Now().UTC()
	id := uuid.New()
	return &MediaAudio{
		ID:          id,
		MediaFileID: mediaFileID,
		ProjectID:   projectID,
		UserID:      userID,
		StorageKey:  NewStorageKey(mediaFileID),
		Codec:       ExtractedCodec,
		SampleRate:  ExtractedSampleRate,
		Channels:    ExtractedChannels,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func NewStorageKey(mediaFileID uuid.UUID) string {
	return "videos/" + mediaFileID.String() + "/" + ExtractedObjectName
}

func (m *MediaAudio) PullEvents() []event.DomainEvent {
	events := m.events
	m.events = nil
	return events
}

func (m *MediaAudio) recordEvent(e event.DomainEvent) {
	m.events = append(m.events, e)
}

func (m *MediaAudio) MarkProcessing() {
	if m.Status == StatusProcessing || m.Status == StatusCompleted {
		return
	}
	m.Status = StatusProcessing
	m.UpdatedAt = time.Now().UTC()
}

func (m *MediaAudio) MarkCompleted(sizeBytes int64) {
	now := time.Now().UTC()
	m.SizeBytes = sizeBytes
	m.Status = StatusCompleted
	m.UpdatedAt = now
	m.recordEvent(MediaAudioReady{
		ID:          uuid.New().String(),
		MediaAudioID: m.ID.String(),
		MediaFileID: m.MediaFileID.String(),
		ProjectID:   m.ProjectID.String(),
		UserID:      m.UserID.String(),
		StorageKey:  m.StorageKey,
		Codec:       m.Codec,
		SampleRate:  m.SampleRate,
		Channels:    m.Channels,
		SizeBytes:   m.SizeBytes,
		Status:      m.Status,
		Timestamp:   now,
	})
}

func (m *MediaAudio) MarkFailed(reason string) {
	now := time.Now().UTC()
	m.Status = StatusFailed
	m.UpdatedAt = now
	m.recordEvent(MediaAudioFailed{
		ID:           uuid.New().String(),
		MediaAudioID: m.ID.String(),
		MediaFileID:  m.MediaFileID.String(),
		ProjectID:    m.ProjectID.String(),
		UserID:       m.UserID.String(),
		Status:       m.Status,
		Reason:       reason,
		Timestamp:    now,
	})
}

type MediaAudioWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, audio *MediaAudio) error
	Update(ctx context.Context, audio *MediaAudio) error
	GetByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*MediaAudio, error)
}
