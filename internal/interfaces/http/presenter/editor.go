package presenter

import (
	querymediafile "go-api/internal/application/query/mediafile"
)

type EditorStateResponse struct {
	Media           EditorMediaResponse           `json:"media"`
	Timeline        EditorTimelineResponse        `json:"timeline"`
	Configuration   EditorConfigurationResponse   `json:"configuration"`
	Decisions       []EditorDecisionResponse      `json:"decisions"`
	Subtitles       EditorSubtitlesResponse       `json:"subtitles"`
	ViralCandidates []EditorViralCandidateResponse `json:"viralCandidates"`
}

type EditorMediaResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	DurationMs int64  `json:"durationMs"`
	Width      *int   `json:"width"`
	Height     *int   `json:"height"`
}

type EditorTimelineResponse struct {
	ID         string                         `json:"id"`
	Version    int                            `json:"version"`
	DurationMs int64                          `json:"durationMs"`
	Segments   []EditorTimelineSegmentResponse `json:"segments"`
}

type EditorTimelineSegmentResponse struct {
	ID            string `json:"id"`
	Index         int    `json:"index"`
	SourceStartMs int64  `json:"sourceStartMs"`
	SourceEndMs   int64  `json:"sourceEndMs"`
	OutputStartMs int64  `json:"outputStartMs"`
	OutputEndMs   int64  `json:"outputEndMs"`
}

type EditorConfigurationResponse struct {
	Silence    EditorSilenceConfigResponse    `json:"silence"`
	Filler     EditorToggleConfigResponse     `json:"filler"`
	Repetition EditorToggleConfigResponse     `json:"repetition"`
	Subtitles  EditorSubtitlesConfigResponse  `json:"subtitles"`
	Output     EditorOutputConfigResponse     `json:"output"`
}

type EditorSilenceConfigResponse struct {
	Enabled                 bool     `json:"enabled"`
	ThresholdMode           string   `json:"thresholdMode"`
	DetectionLevel          string   `json:"detectionLevel"`
	MinDurationMs           int      `json:"minDurationMs"`
	PaddingBeforeMs         int      `json:"paddingBeforeMs"`
	PaddingAfterMs          int      `json:"paddingAfterMs"`
	ThresholdDB             *float64 `json:"thresholdDb"`
	NoiseFloorDB            *float64 `json:"noiseFloorDb"`
	CalculatedThresholdDB   *float64 `json:"calculatedSilenceThresholdDb"`
}

type EditorToggleConfigResponse struct {
	Enabled bool `json:"enabled"`
}

type EditorSubtitlesConfigResponse struct {
	Enabled  bool   `json:"enabled"`
	MaxWords int    `json:"maxWords"`
	StyleID  string `json:"styleId"`
}

type EditorOutputConfigResponse struct {
	AspectRatio string `json:"aspectRatio"`
}

type EditorDecisionResponse struct {
	ID              string  `json:"id"`
	Type            string  `json:"type"`
	Label           *string `json:"label"`
	SourceStartMs   int64   `json:"sourceStartMs"`
	SourceEndMs     int64   `json:"sourceEndMs"`
	AutomaticAction *string `json:"automaticAction"`
	EffectiveAction string  `json:"effectiveAction"`
	ModifiedByUser  bool    `json:"modifiedByUser"`
}

type EditorSubtitlesResponse struct {
	Words []EditorWordResponse `json:"words"`
}

type EditorWordResponse struct {
	ID            string   `json:"id"`
	Text          string   `json:"text"`
	SourceStartMs int64    `json:"sourceStartMs"`
	SourceEndMs   int64    `json:"sourceEndMs"`
	Confidence    *float64 `json:"confidence"`
}

type EditorViralCandidateResponse struct {
	ID            string  `json:"id"`
	SourceStartMs int64   `json:"sourceStartMs"`
	SourceEndMs   int64   `json:"sourceEndMs"`
	Score         float64 `json:"score"`
	Title         string  `json:"title"`
	Hook          string  `json:"hook"`
	Selected      bool    `json:"selected"`
}

func NewEditorStateResponse(view querymediafile.EditorStateView) EditorStateResponse {
	segments := make([]EditorTimelineSegmentResponse, 0, len(view.Timeline.Segments))
	for _, s := range view.Timeline.Segments {
		segments = append(segments, EditorTimelineSegmentResponse{
			ID:            s.ID.String(),
			Index:         s.Index,
			SourceStartMs: s.SourceStartMs,
			SourceEndMs:   s.SourceEndMs,
			OutputStartMs: s.OutputStartMs,
			OutputEndMs:   s.OutputEndMs,
		})
	}
	decisions := make([]EditorDecisionResponse, 0, len(view.Decisions))
	for _, d := range view.Decisions {
		decisions = append(decisions, EditorDecisionResponse{
			ID:              d.ID.String(),
			Type:            d.Type,
			Label:           d.Label,
			SourceStartMs:   d.SourceStartMs,
			SourceEndMs:     d.SourceEndMs,
			AutomaticAction: d.AutomaticAction,
			EffectiveAction: d.EffectiveAction,
			ModifiedByUser:  d.ModifiedByUser,
		})
	}
	words := make([]EditorWordResponse, 0, len(view.Subtitles.Words))
	for _, w := range view.Subtitles.Words {
		words = append(words, EditorWordResponse{
			ID:            w.ID.String(),
			Text:          w.Text,
			SourceStartMs: w.SourceStartMs,
			SourceEndMs:   w.SourceEndMs,
			Confidence:    w.Confidence,
		})
	}
	virals := make([]EditorViralCandidateResponse, 0, len(view.ViralCandidates))
	for _, v := range view.ViralCandidates {
		virals = append(virals, EditorViralCandidateResponse{
			ID:            v.ID.String(),
			SourceStartMs: v.SourceStartMs,
			SourceEndMs:   v.SourceEndMs,
			Score:         v.Score,
			Title:         v.Title,
			Hook:          v.Hook,
			Selected:      v.Selected,
		})
	}
	cfg := view.Configuration
	return EditorStateResponse{
		Media: EditorMediaResponse{
			ID:         view.Media.ID.String(),
			Name:       view.Media.Name,
			URL:        view.Media.URL,
			DurationMs: view.Media.DurationMs,
			Width:      view.Media.Width,
			Height:     view.Media.Height,
		},
		Timeline: EditorTimelineResponse{
			ID:         view.Timeline.ID.String(),
			Version:    view.Timeline.Version,
			DurationMs: view.Timeline.DurationMs,
			Segments:   segments,
		},
		Configuration: EditorConfigurationResponse{
			Silence: EditorSilenceConfigResponse{
				Enabled:               cfg.Silence.Enabled,
				ThresholdMode:         cfg.Silence.ThresholdMode,
				DetectionLevel:        cfg.Silence.DetectionLevel,
				MinDurationMs:         cfg.Silence.MinDurationMs,
				PaddingBeforeMs:       cfg.Silence.PaddingBeforeMs,
				PaddingAfterMs:        cfg.Silence.PaddingAfterMs,
				ThresholdDB:           cfg.Silence.ThresholdDB,
				NoiseFloorDB:          cfg.Silence.NoiseFloorDB,
				CalculatedThresholdDB: cfg.Silence.CalculatedThresholdDB,
			},
			Filler:     EditorToggleConfigResponse{Enabled: cfg.Filler.Enabled},
			Repetition: EditorToggleConfigResponse{Enabled: cfg.Repetition.Enabled},
			Subtitles: EditorSubtitlesConfigResponse{
				Enabled:  cfg.Subtitles.Enabled,
				MaxWords: cfg.Subtitles.MaxWords,
				StyleID:  cfg.Subtitles.StyleID,
			},
			Output: EditorOutputConfigResponse{AspectRatio: cfg.Output.AspectRatio},
		},
		Decisions:       decisions,
		Subtitles:       EditorSubtitlesResponse{Words: words},
		ViralCandidates: virals,
	}
}

type FinalizeAcceptedResponse struct {
	Status                  string `json:"status"`
	JobID                   string `json:"jobId"`
	TimelineID              string `json:"timelineId"`
	PreviousTimelineVersion int    `json:"previousTimelineVersion"`
}

func NewFinalizeAcceptedResponse(jobID, timelineID string, previousVersion int) FinalizeAcceptedResponse {
	return FinalizeAcceptedResponse{
		Status:                  "accepted",
		JobID:                   jobID,
		TimelineID:              timelineID,
		PreviousTimelineVersion: previousVersion,
	}
}

// EditorRebuildAcceptedResponse is returned when an editor action enqueues a timeline rebuild.
// Frontend should wait for realtime media_file.timeline_updated then GET /editor.
type EditorRebuildAcceptedResponse struct {
	Status                  string `json:"status"`
	JobID                   string `json:"jobId"`
	PreviousTimelineVersion int    `json:"previousTimelineVersion"`
}

func NewEditorRebuildAcceptedResponse(jobID string, previousVersion int) EditorRebuildAcceptedResponse {
	return EditorRebuildAcceptedResponse{
		Status:                  "accepted",
		JobID:                   jobID,
		PreviousTimelineVersion: previousVersion,
	}
}
