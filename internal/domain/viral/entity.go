package viral

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Candidate is a short-form clip proposal in SOURCE TIME (before Timeline mapping).
type Candidate struct {
	ID              uuid.UUID
	MediaFileID     uuid.UUID
	ProjectID       uuid.UUID
	SourceStartMs   int64
	SourceEndMs     int64
	Score           float64
	HookScore       *float64
	StandaloneScore *float64
	PayoffScore     *float64
	InterestScore   *float64
	Title           string
	Hook            string
	Reason          string
	Provider        string
	Model           string
	Selected        bool
	CreatedAt       time.Time
}

func (c *Candidate) DurationMs() int64 {
	if c.SourceEndMs <= c.SourceStartMs {
		return 0
	}
	return c.SourceEndMs - c.SourceStartMs
}

type CandidateWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	ReplaceForMediaFile(ctx context.Context, mediaFileID uuid.UUID, rows []*Candidate) error
	CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error)
}
