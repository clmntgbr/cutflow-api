package mediafile

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	domainmediaconfig "go-api/internal/domain/mediaconfig"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainsilence "go-api/internal/domain/silence"
	domaintimeline "go-api/internal/domain/timeline"
	domaintranscript "go-api/internal/domain/transcript"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"
	domainviral "go-api/internal/domain/viral"

	"github.com/google/uuid"
)

var ErrEditorNotReady = errors.New("editor not ready")

type GetEditorStateQuery struct {
	MediaFileID uuid.UUID
	UserID      uuid.UUID
}

type EditorMediaView struct {
	ID         uuid.UUID
	Name       string
	URL        string
	DurationMs int64
	Width      *int
	Height     *int
}

type EditorTimelineSegmentView struct {
	ID            uuid.UUID
	Index         int
	SourceStartMs int64
	SourceEndMs   int64
	OutputStartMs int64
	OutputEndMs   int64
}

type EditorTimelineView struct {
	ID         uuid.UUID
	Version    int
	DurationMs int64
	Segments   []EditorTimelineSegmentView
}

type EditorSilenceConfigView struct {
	Enabled          bool
	ThresholdMode    string
	DetectionLevel   string
	MinDurationMs    int
	PaddingBeforeMs  int
	PaddingAfterMs   int
	ThresholdDB      *float64
	NoiseFloorDB     *float64
	CalculatedThresholdDB *float64
}

type EditorConfigView struct {
	Silence     EditorSilenceConfigView
	Filler      EditorToggleConfigView
	Repetition  EditorToggleConfigView
	Subtitles   EditorSubtitlesConfigView
	Output      EditorOutputConfigView
}

type EditorToggleConfigView struct {
	Enabled bool
}

type EditorSubtitlesConfigView struct {
	Enabled  bool
	MaxWords int
	StyleID  string
}

type EditorOutputConfigView struct {
	AspectRatio string
}

type EditorDecisionView struct {
	ID              uuid.UUID
	Type            string
	Label           *string
	SourceStartMs   int64
	SourceEndMs     int64
	AutomaticAction *string
	EffectiveAction string
	ModifiedByUser  bool
}

type EditorWordView struct {
	ID            uuid.UUID
	Text          string
	SourceStartMs int64
	SourceEndMs   int64
	Confidence    *float64
}

type EditorSubtitlesView struct {
	Words []EditorWordView
}

type EditorViralCandidateView struct {
	ID            uuid.UUID
	SourceStartMs int64
	SourceEndMs   int64
	Score         float64
	Title         string
	Hook          string
	Selected      bool
}

type EditorStateView struct {
	Media           EditorMediaView
	Timeline        EditorTimelineView
	Configuration   EditorConfigView
	Decisions       []EditorDecisionView
	Subtitles       EditorSubtitlesView
	ViralCandidates []EditorViralCandidateView
}

type EditorMediaLoader interface {
	FindOwnedForEditor(ctx context.Context, id, userID uuid.UUID) (*domainmediafile.MediaFileEditorView, error)
}

type EditorTimelineLoader interface {
	GetActiveWithSegments(ctx context.Context, mediaFileID uuid.UUID) (*EditorTimelineView, error)
}

type EditorViralLister interface {
	ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]*domainviral.Candidate, error)
}

type GetEditorStateHandler struct {
	mediaRepo      EditorMediaLoader
	configRepo     domainmediaconfig.MediaConfigurationWriteRepository
	silenceRepo    domainsilence.DetectedSilenceWriteRepository
	issueRepo      domaintranscriptissue.IssueWriteRepository
	overrideRepo   interface {
		ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]domaintimeline.Override, error)
	}
	timelineRepo   EditorTimelineLoader
	transcriptRepo domaintranscript.TranscriptWriteRepository
	viralRepo      EditorViralLister
	storage        port.Storage
}

func NewGetEditorStateHandler(
	mediaRepo EditorMediaLoader,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	issueRepo domaintranscriptissue.IssueWriteRepository,
	overrideRepo interface {
		ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]domaintimeline.Override, error)
	},
	timelineRepo EditorTimelineLoader,
	transcriptRepo domaintranscript.TranscriptWriteRepository,
	viralRepo EditorViralLister,
	storage port.Storage,
) *GetEditorStateHandler {
	return &GetEditorStateHandler{
		mediaRepo:      mediaRepo,
		configRepo:     configRepo,
		silenceRepo:    silenceRepo,
		issueRepo:      issueRepo,
		overrideRepo:   overrideRepo,
		timelineRepo:   timelineRepo,
		transcriptRepo: transcriptRepo,
		viralRepo:      viralRepo,
		storage:        storage,
	}
}

func (h *GetEditorStateHandler) Handle(ctx context.Context, q GetEditorStateQuery) (*EditorStateView, error) {
	media, err := h.mediaRepo.FindOwnedForEditor(ctx, q.MediaFileID, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("load media: %w", err)
	}
	if media == nil {
		return nil, domainmediafile.ErrMediaNotFound
	}

	timeline, err := h.timelineRepo.GetActiveWithSegments(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load timeline: %w", err)
	}
	if timeline == nil {
		return nil, ErrEditorNotReady
	}

	cfg, err := h.configRepo.GetByMediaFileID(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg == nil {
		cfg = domainmediaconfig.NewDefault(q.MediaFileID)
	}

	silences, err := h.silenceRepo.ListByMediaFileID(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load silences: %w", err)
	}
	issues, err := h.issueRepo.ListByMediaFileID(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load issues: %w", err)
	}
	overrides, err := h.overrideRepo.ListByMediaFileID(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load overrides: %w", err)
	}

	auto := domaintimeline.BuildDecisions(domaintimeline.BuildInput{
		MediaFileID:          q.MediaFileID,
		MediaDurationMs:      media.DurationMs,
		SilenceRemoval:       cfg.SilenceRemovalEnabled,
		FillerRemoval:        cfg.FillerRemovalEnabled,
		RepetitionRemoval:    cfg.RepetitionRemovalEnabled,
		SilenceMinDurationMs: cfg.SilenceMinDurationMs,
		SilencePadBeforeMs:   cfg.SilencePaddingBeforeMs,
		SilencePadAfterMs:    cfg.SilencePaddingAfterMs,
		SpeechMinDurationMs:  cfg.SpeechMinDurationMs,
		Silences:             domainsilence.IntervalsFromDetected(silences),
		Issues:               issues,
		Overrides:            nil, // automatic proposals only
	})
	resolved := domaintimeline.ResolveEditorDecisions(auto, overrides, domaintimeline.LabelsFromIssues(issues))

	url := ""
	if media.StorageKey != "" && h.storage != nil {
		presigned, err := h.storage.PresignedGetURL(ctx, media.StorageKey, time.Hour)
		if err != nil {
			log.Printf("failed to presign editor media url mediaFileId=%s: %v", q.MediaFileID, err)
		} else {
			url = presigned
		}
	}

	words, err := h.loadWords(ctx, q.MediaFileID)
	if err != nil {
		return nil, err
	}

	virals, err := h.viralRepo.ListByMediaFileID(ctx, q.MediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load viral candidates: %w", err)
	}

	return &EditorStateView{
		Media: EditorMediaView{
			ID:         media.ID,
			Name:       media.OriginalFilename,
			URL:        url,
			DurationMs: media.DurationMs,
			Width:      media.Width,
			Height:     media.Height,
		},
		Timeline:      *timeline,
		Configuration: mapEditorConfig(cfg),
		Decisions:     mapEditorDecisions(resolved),
		Subtitles:     EditorSubtitlesView{Words: words},
		ViralCandidates: mapViralCandidates(virals),
	}, nil
}

func (h *GetEditorStateHandler) loadWords(ctx context.Context, mediaFileID uuid.UUID) ([]EditorWordView, error) {
	transcript, err := h.transcriptRepo.GetByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return nil, fmt.Errorf("load transcript: %w", err)
	}
	if transcript == nil || transcript.Status != domaintranscript.StatusCompleted {
		return []EditorWordView{}, nil
	}
	words, err := h.transcriptRepo.ListWords(ctx, transcript.ID)
	if err != nil {
		return nil, fmt.Errorf("load words: %w", err)
	}
	out := make([]EditorWordView, 0, len(words))
	for _, w := range words {
		out = append(out, EditorWordView{
			ID:            uuid.NewSHA1(mediaFileID, []byte(fmt.Sprintf("%d:%d:%s", w.WordIndex, w.SourceStartMs, w.Text))),
			Text:          w.Text,
			SourceStartMs: w.SourceStartMs,
			SourceEndMs:   w.SourceEndMs,
			Confidence:    w.Confidence,
		})
	}
	return out, nil
}

func mapEditorConfig(cfg *domainmediaconfig.MediaConfiguration) EditorConfigView {
	return EditorConfigView{
		Silence: EditorSilenceConfigView{
			Enabled:               cfg.SilenceRemovalEnabled,
			ThresholdMode:         cfg.SilenceThresholdMode,
			DetectionLevel:        cfg.SilenceDetectionLevel,
			MinDurationMs:         cfg.SilenceMinDurationMs,
			PaddingBeforeMs:       cfg.SilencePaddingBeforeMs,
			PaddingAfterMs:        cfg.SilencePaddingAfterMs,
			ThresholdDB:           cfg.SilenceThresholdDB,
			NoiseFloorDB:          cfg.NoiseFloorDB,
			CalculatedThresholdDB: cfg.CalculatedSilenceThresholdDB,
		},
		Filler:     EditorToggleConfigView{Enabled: cfg.FillerRemovalEnabled},
		Repetition: EditorToggleConfigView{Enabled: cfg.RepetitionRemovalEnabled},
		Subtitles: EditorSubtitlesConfigView{
			Enabled:  cfg.SubtitlesEnabled,
			MaxWords: cfg.SubtitleMaxWords,
			StyleID:  "default",
		},
		Output: EditorOutputConfigView{AspectRatio: "original"},
	}
}

func mapEditorDecisions(in []domaintimeline.EditorDecision) []EditorDecisionView {
	out := make([]EditorDecisionView, 0, len(in))
	for _, d := range in {
		out = append(out, EditorDecisionView{
			ID:              d.ID,
			Type:            d.Type,
			Label:           d.Label,
			SourceStartMs:   d.SourceStartMs,
			SourceEndMs:     d.SourceEndMs,
			AutomaticAction: d.AutomaticAction,
			EffectiveAction: d.EffectiveAction,
			ModifiedByUser:  d.ModifiedByUser,
		})
	}
	return out
}

func mapViralCandidates(in []*domainviral.Candidate) []EditorViralCandidateView {
	out := make([]EditorViralCandidateView, 0, len(in))
	for _, c := range in {
		if c == nil {
			continue
		}
		out = append(out, EditorViralCandidateView{
			ID:            c.ID,
			SourceStartMs: c.SourceStartMs,
			SourceEndMs:   c.SourceEndMs,
			Score:         c.Score,
			Title:         c.Title,
			Hook:          c.Hook,
			Selected:      c.Selected,
		})
	}
	return out
}
