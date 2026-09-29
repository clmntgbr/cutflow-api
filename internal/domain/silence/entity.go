package silence

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Interval struct {
	StartMs int64
	EndMs   int64
}

func (i Interval) DurationMs() int64 {
	if i.EndMs <= i.StartMs {
		return 0
	}
	return i.EndMs - i.StartMs
}

type DetectedSilence struct {
	ID          uuid.UUID
	MediaFileID uuid.UUID
	StartMs     int64
	EndMs       int64
	CreatedAt   time.Time
}

func NewDetectedSilence(mediaFileID uuid.UUID, startMs, endMs int64) *DetectedSilence {
	return &DetectedSilence{
		ID:          uuid.New(),
		MediaFileID: mediaFileID,
		StartMs:     startMs,
		EndMs:       endMs,
		CreatedAt:   time.Now().UTC(),
	}
}

func (s *DetectedSilence) DurationMs() int64 {
	return Interval{StartMs: s.StartMs, EndMs: s.EndMs}.DurationMs()
}

func (s *DetectedSilence) Interval() Interval {
	return Interval{StartMs: s.StartMs, EndMs: s.EndMs}
}

type DetectedSilenceWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	ReplaceForMediaFile(ctx context.Context, mediaFileID uuid.UUID, rows []*DetectedSilence) error
	CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error)
	ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]*DetectedSilence, error)
}
