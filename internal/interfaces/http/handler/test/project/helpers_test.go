package projecttest

import (
	"context"
	"errors"
	"time"

	cmdproject "go-api/internal/application/command/project"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"
)

type mockRequestUploadURLHandler struct {
	called bool
	cmd    cmdproject.RequestUploadURLCommand
	result *cmdproject.RequestUploadURLResult
	err    error
}

func (m *mockRequestUploadURLHandler) Handle(
	_ context.Context,
	cmd cmdproject.RequestUploadURLCommand,
) (*cmdproject.RequestUploadURLResult, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

func newProjectHandler(upload *mockRequestUploadURLHandler) *handler.ProjectHandler {
	if upload == nil {
		upload = &mockRequestUploadURLHandler{}
	}
	return handler.NewProjectHandler(upload)
}

func validUploadBody() map[string]any {
	return map[string]any{
		"filename":    "demo.mp4",
		"contentType": "video/mp4",
		"sizeBytes":   1024,
	}
}

func sampleUploadResult() *cmdproject.RequestUploadURLResult {
	return &cmdproject.RequestUploadURLResult{
		ProjectID:   testutil.TestProjectID.String(),
		MediaFileID: testutil.TestMediaFileID.String(),
		UploadURL:   "http://localhost:9000/media/videos/" + testutil.TestMediaFileID.String() + "/original.mp4",
		ExpiresAt:   time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC),
	}
}

var errUnexpected = errors.New("unexpected handler call")
