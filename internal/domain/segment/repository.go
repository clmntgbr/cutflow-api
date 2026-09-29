package segment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusQueued     = "queued"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
)

type Segment struct {
	ID                 uuid.UUID
	MediaFileID        uuid.UUID
	SegmentIndex       int
	StartMs            int64
	EndMs              int64
	ProcessingStartMs  int64
	ProcessingEndMs    int64
	Status             string
	CreatedAt          time.Time
}

func NewFromWindow(mediaFileID uuid.UUID, window Window) *Segment {
	return &Segment{
		ID:                uuid.New(),
		MediaFileID:       mediaFileID,
		SegmentIndex:      window.Index,
		StartMs:           window.StartMs,
		EndMs:             window.EndMs,
		ProcessingStartMs: window.ProcessingStartMs,
		ProcessingEndMs:   window.ProcessingEndMs,
		Status:            StatusPending,
		CreatedAt:         time.Now().UTC(),
	}
}

type SegmentWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	SaveBatch(ctx context.Context, segments []*Segment) error
	CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error)
}
