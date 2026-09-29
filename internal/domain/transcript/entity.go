package transcript

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

	KindSpeech     = "speech"
	KindFiller     = "filler"
	KindRepetition = "repetition"

	SRTObjectName = "subtitles.srt"
	ASSObjectName = "subtitles.ass"
)

type Word struct {
	WordIndex      int
	Text           string
	SourceStartMs  int64
	SourceEndMs    int64
	Confidence     *float64
	Kind           string
}

type Transcript struct {
	ID             uuid.UUID
	MediaFileID    uuid.UUID
	ProjectID      uuid.UUID
	UserID         uuid.UUID
	Language       string
	Text           string
	SRTStorageKey  string
	ASSStorageKey  string
	ProviderJobID  string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Words          []Word

	events []event.DomainEvent
}

func NewPending(mediaFileID, projectID, userID uuid.UUID) *Transcript {
	now := time.Now().UTC()
	return &Transcript{
		ID:            uuid.New(),
		MediaFileID:   mediaFileID,
		ProjectID:     projectID,
		UserID:        userID,
		SRTStorageKey: NewSRTStorageKey(mediaFileID),
		ASSStorageKey: NewASSStorageKey(mediaFileID),
		Status:        StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func NewSRTStorageKey(mediaFileID uuid.UUID) string {
	return "videos/" + mediaFileID.String() + "/" + SRTObjectName
}

func NewASSStorageKey(mediaFileID uuid.UUID) string {
	return "videos/" + mediaFileID.String() + "/" + ASSObjectName
}

func (t *Transcript) PullEvents() []event.DomainEvent {
	events := t.events
	t.events = nil
	return events
}

func (t *Transcript) recordEvent(e event.DomainEvent) {
	t.events = append(t.events, e)
}

func (t *Transcript) MarkProcessing(providerJobID string) {
	t.Status = StatusProcessing
	t.ProviderJobID = providerJobID
	t.UpdatedAt = time.Now().UTC()
}

func (t *Transcript) MarkCompleted(language, text, srtKey, assKey string, words []Word) {
	now := time.Now().UTC()
	t.Language = language
	t.Text = text
	t.SRTStorageKey = srtKey
	t.ASSStorageKey = assKey
	t.Words = words
	t.Status = StatusCompleted
	t.UpdatedAt = now
	t.recordEvent(TranscriptReady{
		ID:          uuid.New().String(),
		TranscriptID: t.ID.String(),
		MediaFileID: t.MediaFileID.String(),
		ProjectID:   t.ProjectID.String(),
		UserID:      t.UserID.String(),
		Language:    t.Language,
		WordCount:   len(words),
		SRTKey:      t.SRTStorageKey,
		ASSKey:      t.ASSStorageKey,
		Status:      t.Status,
		Timestamp:   now,
	})
}

func (t *Transcript) MarkFailed(reason string) {
	now := time.Now().UTC()
	t.Status = StatusFailed
	t.UpdatedAt = now
	t.recordEvent(TranscriptFailed{
		ID:           uuid.New().String(),
		TranscriptID: t.ID.String(),
		MediaFileID:  t.MediaFileID.String(),
		ProjectID:    t.ProjectID.String(),
		UserID:       t.UserID.String(),
		Status:       t.Status,
		Reason:       reason,
		Timestamp:    now,
	})
}

type TranscriptWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, transcript *Transcript) error
	Update(ctx context.Context, transcript *Transcript) error
	ReplaceWords(ctx context.Context, transcriptID uuid.UUID, words []Word) error
	ListWords(ctx context.Context, transcriptID uuid.UUID) ([]Word, error)
	GetByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*Transcript, error)
}
