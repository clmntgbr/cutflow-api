package presenter

import (
	"time"

	cmdproject "go-api/internal/application/command/project"
	domainproject "go-api/internal/domain/project"
)

type RequestUploadURLResponse struct {
	ProjectID   string    `json:"projectId"`
	MediaFileID string    `json:"mediaFileId"`
	UploadURL   string    `json:"uploadUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func NewRequestUploadURLResponse(result *cmdproject.RequestUploadURLResult) RequestUploadURLResponse {
	return RequestUploadURLResponse{
		ProjectID:   result.ProjectID,
		MediaFileID: result.MediaFileID,
		UploadURL:   result.UploadURL,
		ExpiresAt:   result.ExpiresAt,
	}
}

// ProjectListItemResponse is intentionally lean for short list calls.
type ProjectListItemResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Status       string               `json:"status"`
	CreatedAt    time.Time            `json:"createdAt"`
	ThumbnailURL *string              `json:"thumbnailUrl"`
	Jobs         []ProjectJobResponse `json:"jobs"`
}

func NewProjectListResponseFromViews(views []domainproject.ProjectListView) []ProjectListItemResponse {
	out := make([]ProjectListItemResponse, 0, len(views))
	for _, view := range views {
		out = append(out, ProjectListItemResponse{
			ID:           view.ID.String(),
			Name:         view.Name,
			Status:       view.Status,
			CreatedAt:    view.CreatedAt,
			ThumbnailURL: optionalNonEmptyString(mediaFileThumbnailURL(view.MediaFileID.String(), view.ThumbnailKey)),
			Jobs:         newProjectJobResponses(view.Jobs),
		})
	}
	return out
}

type ProjectDetailResponse struct {
	ID             string                              `json:"id"`
	Name           string                              `json:"name"`
	Status         string                              `json:"status"`
	CreatedAt      time.Time                           `json:"createdAt"`
	UpdatedAt      time.Time                           `json:"updatedAt"`
	MediaFiles     []ProjectMediaFileResponse          `json:"mediaFiles"`
	Jobs           []ProjectJobResponse                `json:"jobs"`
	Timelines      []ProjectTimelineResponse           `json:"timelines"`
	Configurations []ProjectMediaConfigurationResponse `json:"configurations"`
}

type ProjectMediaFileResponse struct {
	ID               string    `json:"id"`
	OriginalFilename *string   `json:"originalFilename"`
	MimeType         *string   `json:"mimeType"`
	SizeBytes        int64     `json:"sizeBytes"`
	DurationMs       int64     `json:"durationMs"`
	Width            *int      `json:"width"`
	Height           *int      `json:"height"`
	FPS              *float64  `json:"fps"`
	VideoCodec       *string   `json:"videoCodec"`
	AudioCodec       *string   `json:"audioCodec"`
	AudioSampleRate  *int      `json:"audioSampleRate"`
	AudioChannels    *int      `json:"audioChannels"`
	OriginalURL      *string   `json:"originalUrl"`
	ThumbnailURL     *string   `json:"thumbnailUrl"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
}

type ProjectJobResponse struct {
	ID           string     `json:"id"`
	MediaFileID  string     `json:"mediaFileId"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"errorMessage"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	StartedAt    *time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
}

type ProjectTimelineResponse struct {
	ID            string                           `json:"id"`
	MediaFileID   string                           `json:"mediaFileId"`
	Version       int                              `json:"version"`
	DurationMs    int64                            `json:"durationMs"`
	Fingerprint   string                           `json:"fingerprint"`
	EngineVersion string                           `json:"engineVersion"`
	IsActive      bool                             `json:"isActive"`
	Segments      []ProjectTimelineSegmentResponse `json:"segments"`
	Decisions     []ProjectEditDecisionResponse    `json:"decisions"`
	CreatedAt     time.Time                        `json:"createdAt"`
	UpdatedAt     time.Time                        `json:"updatedAt"`
}

type ProjectTimelineSegmentResponse struct {
	Index         int    `json:"index"`
	MediaFileID   string `json:"mediaFileId"`
	SourceStartMs int64  `json:"sourceStartMs"`
	SourceEndMs   int64  `json:"sourceEndMs"`
	OutputStartMs int64  `json:"outputStartMs"`
	OutputEndMs   int64  `json:"outputEndMs"`
}

type ProjectEditDecisionResponse struct {
	ID            string   `json:"id"`
	MediaFileID   string   `json:"mediaFileId"`
	Type          string   `json:"type"`
	SourceStartMs int64    `json:"sourceStartMs"`
	SourceEndMs   int64    `json:"sourceEndMs"`
	Action        string   `json:"action"`
	Source        string   `json:"source"`
	Confidence    *float64 `json:"confidence"`
	Reasons       []string `json:"reasons"`
}

type ProjectMediaConfigurationResponse struct {
	ID                           string   `json:"id"`
	MediaFileID                  string   `json:"mediaFileId"`
	SilenceRemovalEnabled        bool     `json:"silenceRemovalEnabled"`
	SilenceThresholdMode         string   `json:"silenceThresholdMode"`
	SilenceThresholdDB           *float64 `json:"silenceThresholdDb"`
	NoiseFloorDB                 *float64 `json:"noiseFloorDb"`
	CalculatedSilenceThresholdDB *float64 `json:"calculatedSilenceThresholdDb"`
	SilenceDetectionLevel        string   `json:"silenceDetectionLevel"`
	SilencePaddingBeforeMs       int      `json:"silencePaddingBeforeMs"`
	SilencePaddingAfterMs        int      `json:"silencePaddingAfterMs"`
	SilenceMinDurationMs         int      `json:"silenceMinDurationMs"`
	SpeechMinDurationMs          int      `json:"speechMinDurationMs"`
	FillerRemovalEnabled         bool     `json:"fillerRemovalEnabled"`
	RepetitionRemovalEnabled     bool     `json:"repetitionRemovalEnabled"`
	SubtitlesEnabled             bool     `json:"subtitlesEnabled"`
	SubtitleMaxWords             int      `json:"subtitleMaxWords"`
	ViralDetectionEnabled        bool     `json:"viralDetectionEnabled"`
	ViralClipMinDurationMs       int      `json:"viralClipMinDurationMs"`
	ViralClipMaxDurationMs       int      `json:"viralClipMaxDurationMs"`
	ViralMaxCandidates           int      `json:"viralMaxCandidates"`
	ViralMinScore                float64  `json:"viralMinScore"`
}

func NewProjectDetailResponseFromView(view domainproject.ProjectDetailView) ProjectDetailResponse {
	mediaFiles := make([]ProjectMediaFileResponse, 0, len(view.MediaFiles))
	for _, media := range view.MediaFiles {
		mediaFiles = append(mediaFiles, ProjectMediaFileResponse{
			ID:               media.ID.String(),
			OriginalFilename: optionalNonEmptyString(media.OriginalFilename),
			MimeType:         optionalNonEmptyString(media.MimeType),
			SizeBytes:        media.SizeBytes,
			DurationMs:       media.DurationMs,
			Width:            media.Width,
			Height:           media.Height,
			FPS:              media.FPS,
			VideoCodec:       optionalNonEmptyString(media.VideoCodec),
			AudioCodec:       optionalNonEmptyString(media.AudioCodec),
			AudioSampleRate:  media.AudioSampleRate,
			AudioChannels:    media.AudioChannels,
			OriginalURL:      optionalNonEmptyString(media.OriginalURL),
			ThumbnailURL:     optionalNonEmptyString(mediaFileThumbnailURL(media.ID.String(), media.ThumbnailKey)),
			Status:           media.Status,
			CreatedAt:        media.CreatedAt,
		})
	}
	return ProjectDetailResponse{
		ID:             view.ID.String(),
		Name:           view.Name,
		Status:         view.Status,
		CreatedAt:      view.CreatedAt,
		UpdatedAt:      view.UpdatedAt,
		MediaFiles:     mediaFiles,
		Jobs:           newProjectJobResponses(view.Jobs),
		Timelines:      newProjectTimelineResponses(view.Timelines),
		Configurations: newProjectConfigurationResponses(view.Configurations),
	}
}

func newProjectJobResponses(jobs []domainproject.ProjectJobView) []ProjectJobResponse {
	out := make([]ProjectJobResponse, 0, len(jobs))
	for _, job := range jobs {
		out = append(out, ProjectJobResponse{
			ID:           job.ID.String(),
			MediaFileID:  job.MediaFileID.String(),
			Name:         job.Name,
			Status:       job.Status,
			ErrorMessage: optionalNonEmptyString(job.ErrorMessage),
			CreatedAt:    job.CreatedAt,
			UpdatedAt:    job.UpdatedAt,
			StartedAt:    job.StartedAt,
			CompletedAt:  job.CompletedAt,
		})
	}
	return out
}

func newProjectTimelineResponses(timelines []domainproject.ProjectTimelineView) []ProjectTimelineResponse {
	out := make([]ProjectTimelineResponse, 0, len(timelines))
	for _, tl := range timelines {
		segments := make([]ProjectTimelineSegmentResponse, 0, len(tl.Segments))
		for _, seg := range tl.Segments {
			segments = append(segments, ProjectTimelineSegmentResponse{
				Index:         seg.Index,
				MediaFileID:   seg.MediaFileID.String(),
				SourceStartMs: seg.SourceStartMs,
				SourceEndMs:   seg.SourceEndMs,
				OutputStartMs: seg.OutputStartMs,
				OutputEndMs:   seg.OutputEndMs,
			})
		}
		decisions := make([]ProjectEditDecisionResponse, 0, len(tl.Decisions))
		for _, dec := range tl.Decisions {
			reasons := dec.Reasons
			if reasons == nil {
				reasons = []string{}
			}
			decisions = append(decisions, ProjectEditDecisionResponse{
				ID:            dec.ID.String(),
				MediaFileID:   dec.MediaFileID.String(),
				Type:          dec.Type,
				SourceStartMs: dec.SourceStartMs,
				SourceEndMs:   dec.SourceEndMs,
				Action:        dec.Action,
				Source:        dec.Source,
				Confidence:    dec.Confidence,
				Reasons:       reasons,
			})
		}
		out = append(out, ProjectTimelineResponse{
			ID:            tl.ID.String(),
			MediaFileID:   tl.MediaFileID.String(),
			Version:       tl.Version,
			DurationMs:    tl.DurationMs,
			Fingerprint:   tl.Fingerprint,
			EngineVersion: tl.EngineVersion,
			IsActive:      tl.IsActive,
			Segments:      segments,
			Decisions:     decisions,
			CreatedAt:     tl.CreatedAt,
			UpdatedAt:     tl.UpdatedAt,
		})
	}
	return out
}

func newProjectConfigurationResponses(
	configurations []domainproject.ProjectMediaConfigurationView,
) []ProjectMediaConfigurationResponse {
	out := make([]ProjectMediaConfigurationResponse, 0, len(configurations))
	for _, cfg := range configurations {
		out = append(out, ProjectMediaConfigurationResponse{
			ID:                           cfg.ID.String(),
			MediaFileID:                  cfg.MediaFileID.String(),
			SilenceRemovalEnabled:        cfg.SilenceRemovalEnabled,
			SilenceThresholdMode:         cfg.SilenceThresholdMode,
			SilenceThresholdDB:           cfg.SilenceThresholdDB,
			NoiseFloorDB:                 cfg.NoiseFloorDB,
			CalculatedSilenceThresholdDB: cfg.CalculatedSilenceThresholdDB,
			SilenceDetectionLevel:        cfg.SilenceDetectionLevel,
			SilencePaddingBeforeMs:       cfg.SilencePaddingBeforeMs,
			SilencePaddingAfterMs:        cfg.SilencePaddingAfterMs,
			SilenceMinDurationMs:         cfg.SilenceMinDurationMs,
			SpeechMinDurationMs:          cfg.SpeechMinDurationMs,
			FillerRemovalEnabled:         cfg.FillerRemovalEnabled,
			RepetitionRemovalEnabled:     cfg.RepetitionRemovalEnabled,
			SubtitlesEnabled:             cfg.SubtitlesEnabled,
			SubtitleMaxWords:             cfg.SubtitleMaxWords,
			ViralDetectionEnabled:        cfg.ViralDetectionEnabled,
			ViralClipMinDurationMs:       cfg.ViralClipMinDurationMs,
			ViralClipMaxDurationMs:       cfg.ViralClipMaxDurationMs,
			ViralMaxCandidates:           cfg.ViralMaxCandidates,
			ViralMinScore:                cfg.ViralMinScore,
		})
	}
	return out
}

func mediaFileThumbnailURL(mediaFileID, thumbnailKey string) string {
	if thumbnailKey == "" {
		return ""
	}
	return "/api/media-files/" + mediaFileID + "/thumbnail"
}
