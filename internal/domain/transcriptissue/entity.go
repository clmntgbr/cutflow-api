package transcriptissue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TypeFiller     = "filler"
	TypeRepetition = "repetition"
	TypeFalseStart = "false_start"
)

// Issue is a SourceTime span proposed for removal (filler / repetition / false start).
type Issue struct {
	ID             uuid.UUID
	MediaFileID    uuid.UUID
	TranscriptID   uuid.UUID
	Type           string
	Text           string
	SourceStartMs  int64
	SourceEndMs    int64
	Confidence     *float64
	WordStartIndex int
	WordEndIndex   int
	CreatedAt      time.Time
}

func NewIssue(
	mediaFileID, transcriptID uuid.UUID,
	issueType, text string,
	startMs, endMs int64,
	confidence *float64,
	wordStart, wordEnd int,
) *Issue {
	return &Issue{
		ID:             uuid.New(),
		MediaFileID:    mediaFileID,
		TranscriptID:   transcriptID,
		Type:           issueType,
		Text:           text,
		SourceStartMs:  startMs,
		SourceEndMs:    endMs,
		Confidence:     confidence,
		WordStartIndex: wordStart,
		WordEndIndex:   wordEnd,
		CreatedAt:      time.Now().UTC(),
	}
}

type IssueWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	ReplaceForMediaFile(ctx context.Context, mediaFileID uuid.UUID, rows []*Issue) error
	CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error)
	ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]*Issue, error)
}
