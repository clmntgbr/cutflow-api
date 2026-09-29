package project

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ProjectWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
}

type ProjectReadRepository interface {
	FindByID(ctx context.Context, id, userID uuid.UUID) (*ProjectDetailView, error)
	List(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]ProjectListView, int64, error)
}

// ProjectListView is a lean row for short list calls.
type ProjectListView struct {
	ID           uuid.UUID
	Name         string
	Status       string
	CreatedAt    time.Time
	MediaFileID  uuid.UUID
	ThumbnailKey string
	Jobs         []ProjectJobView
}

type ProjectDetailView struct {
	ID             uuid.UUID
	Name           string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MediaFiles     []ProjectMediaFileView
	Jobs           []ProjectJobView
	Timelines      []ProjectTimelineView
	Configurations []ProjectMediaConfigurationView
}

type ProjectMediaFileView struct {
	ID               uuid.UUID
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	DurationMs       int64
	Width            *int
	Height           *int
	FPS              *float64
	VideoCodec       string
	AudioCodec       string
	AudioSampleRate  *int
	AudioChannels    *int
	StorageKey       string
	OriginalURL      string
	ThumbnailKey     string
	Status           string
	CreatedAt        time.Time
}

type ProjectJobView struct {
	ID           uuid.UUID
	MediaFileID  uuid.UUID
	Name         string
	Status       string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
}

type ProjectTimelineView struct {
	ID            uuid.UUID
	MediaFileID   uuid.UUID
	Version       int
	DurationMs    int64
	Fingerprint   string
	EngineVersion string
	IsActive      bool
	Segments      []ProjectTimelineSegmentView
	Decisions     []ProjectEditDecisionView
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProjectTimelineSegmentView struct {
	Index         int
	MediaFileID   uuid.UUID
	SourceStartMs int64
	SourceEndMs   int64
	OutputStartMs int64
	OutputEndMs   int64
}

type ProjectEditDecisionView struct {
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

type ProjectMediaConfigurationView struct {
	ID                           uuid.UUID
	MediaFileID                  uuid.UUID
	SilenceRemovalEnabled        bool
	SilenceThresholdMode         string
	SilenceThresholdDB           *float64
	NoiseFloorDB                 *float64
	CalculatedSilenceThresholdDB *float64
	SilenceDetectionLevel        string
	SilencePaddingBeforeMs       int
	SilencePaddingAfterMs        int
	SilenceMinDurationMs         int
	SpeechMinDurationMs          int
	FillerRemovalEnabled         bool
	RepetitionRemovalEnabled     bool
	SubtitlesEnabled             bool
	SubtitleMaxWords             int
	ViralDetectionEnabled        bool
	ViralClipMinDurationMs       int
	ViralClipMaxDurationMs       int
	ViralMaxCandidates           int
	ViralMinScore                float64
}
