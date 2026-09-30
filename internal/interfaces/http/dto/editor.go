package dto

const (
	EditorActionUpdateConfiguration   = "update_configuration"
	EditorActionOverrideDecision      = "override_decision"
	EditorActionClearDecisionOverride = "clear_decision_override"
	EditorActionCreateManualCut       = "create_manual_cut"
)

// UpdateEditorRequest is the unified PATCH /editor body (one action per request).
type UpdateEditorRequest struct {
	Type            string                            `json:"type" validate:"required"`
	TimelineVersion *int                              `json:"timelineVersion"`
	Configuration   *EditorConfigurationPatchRequest  `json:"configuration"`
	DecisionID      *string                           `json:"decisionId"`
	Action          *string                           `json:"action"`
	SourceStartMs   *int64                            `json:"sourceStartMs"`
	SourceEndMs     *int64                            `json:"sourceEndMs"`
}

type EditorConfigurationPatchRequest struct {
	Silence    *PatchSilenceConfigRequest   `json:"silence"`
	Filler     *PatchToggleConfigRequest    `json:"filler"`
	Repetition *PatchToggleConfigRequest    `json:"repetition"`
	Subtitles  *PatchSubtitlesConfigRequest `json:"subtitles"`
}

type PatchSilenceConfigRequest struct {
	Enabled         *bool    `json:"enabled"`
	ThresholdMode   *string  `json:"thresholdMode"`
	DetectionLevel  *string  `json:"detectionLevel"`
	MinDurationMs   *int     `json:"minDurationMs"`
	PaddingBeforeMs *int     `json:"paddingBeforeMs"`
	PaddingAfterMs  *int     `json:"paddingAfterMs"`
	ThresholdDB     *float64 `json:"thresholdDb"`
}

type PatchToggleConfigRequest struct {
	Enabled *bool `json:"enabled"`
}

type PatchSubtitlesConfigRequest struct {
	Enabled  *bool `json:"enabled"`
	MaxWords *int  `json:"maxWords"`
}

type FinalizeEditorRequest struct {
	TimelineID      string                 `json:"timelineId" validate:"required,uuid"`
	TimelineVersion int                    `json:"timelineVersion" validate:"required,min=1"`
	Output          *FinalizeOutputRequest `json:"output"`
}

type FinalizeOutputRequest struct {
	AspectRatio *string `json:"aspectRatio"`
	Resolution  *string `json:"resolution"`
}
