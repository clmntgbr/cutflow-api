package timeline

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	EngineVersion = "timeline-engine-v1"

	ActionRemove = "remove"
	ActionKeep   = "keep"

	SourceAutomatic = "automatic"
	SourceUser      = "user"

	DecisionSilence     = "silence"
	DecisionFiller      = "filler"
	DecisionRepetition  = "repetition"
	DecisionFalseStart  = "false_start"
	DecisionManual      = "manual"
)

// Decision is a SOURCE TIME montage proposition (automatic or user).
type Decision struct {
	ID            uuid.UUID
	MediaFileID   uuid.UUID
	Type          string
	SourceStartMs int64
	SourceEndMs   int64
	Action        string
	Source        string
	Confidence    *float64
	Reasons       []string
}

func (d Decision) DurationMs() int64 {
	if d.SourceEndMs <= d.SourceStartMs {
		return 0
	}
	return d.SourceEndMs - d.SourceStartMs
}

// Override is an explicit user KEEP/REMOVE on a SOURCE TIME range.
type Override struct {
	ID            uuid.UUID
	MediaFileID   uuid.UUID
	Type          string
	SourceStartMs int64
	SourceEndMs   int64
	Action        string
}

// Segment is a KEEP window mapped to OUTPUT TIME.
type Segment struct {
	Index         int
	MediaFileID   uuid.UUID
	SourceStartMs int64
	SourceEndMs   int64
	OutputStartMs int64
	OutputEndMs   int64
}

func (s Segment) SourceDurationMs() int64 {
	return s.SourceEndMs - s.SourceStartMs
}

func (s Segment) OutputDurationMs() int64 {
	return s.OutputEndMs - s.OutputStartMs
}

// Timeline is a versioned montage description (never mutates media files).
type Timeline struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	MediaFileID  uuid.UUID
	Version      int
	DurationMs   int64
	Fingerprint  string
	EngineVersion string
	IsActive     bool
	Segments     []Segment
	Decisions    []Decision
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TimelineWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	GetActiveByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*Timeline, error)
	FindByFingerprint(ctx context.Context, mediaFileID uuid.UUID, fingerprint string) (*Timeline, error)
	ExistsByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (bool, error)
	NextVersion(ctx context.Context, mediaFileID uuid.UUID) (int, error)
	Save(ctx context.Context, tl *Timeline) error
	DeactivateOthers(ctx context.Context, mediaFileID, keepID uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
}
