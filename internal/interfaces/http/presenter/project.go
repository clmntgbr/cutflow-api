package presenter

import (
	"time"

	cmdproject "go-api/internal/application/command/project"
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
