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
	ID         string                     `json:"id"`
	Name       string                     `json:"name"`
	Status     string                     `json:"status"`
	CreatedAt  time.Time                  `json:"createdAt"`
	UpdatedAt  time.Time                  `json:"updatedAt"`
	MediaFiles []ProjectMediaFileResponse `json:"mediaFiles"`
	Jobs       []ProjectJobResponse       `json:"jobs"`
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
		ID:         view.ID.String(),
		Name:       view.Name,
		Status:     view.Status,
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
		MediaFiles: mediaFiles,
		Jobs:       newProjectJobResponses(view.Jobs),
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

func mediaFileThumbnailURL(mediaFileID, thumbnailKey string) string {
	if thumbnailKey == "" {
		return ""
	}
	return "/api/media-files/" + mediaFileID + "/thumbnail"
}
