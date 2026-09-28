package projecttest

import (
	"context"
	"errors"
	"time"

	cmdproject "go-api/internal/application/command/project"
	queryproject "go-api/internal/application/query/project"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
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

type mockListProjectsHandler struct {
	called bool
	query  queryproject.ListProjectsQuery
	views  []domainproject.ProjectListView
	total  int64
	err    error
}

func (m *mockListProjectsHandler) Handle(
	_ context.Context,
	q queryproject.ListProjectsQuery,
) ([]domainproject.ProjectListView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockGetProjectByIDHandler struct {
	called bool
	query  queryproject.GetProjectByIDQuery
	view   *domainproject.ProjectDetailView
	err    error
}

func (m *mockGetProjectByIDHandler) Handle(
	_ context.Context,
	q queryproject.GetProjectByIDQuery,
) (*domainproject.ProjectDetailView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

func newProjectHandler(
	upload *mockRequestUploadURLHandler,
	list *mockListProjectsHandler,
	getByID *mockGetProjectByIDHandler,
) *handler.ProjectHandler {
	if upload == nil {
		upload = &mockRequestUploadURLHandler{}
	}
	if list == nil {
		list = &mockListProjectsHandler{}
	}
	if getByID == nil {
		getByID = &mockGetProjectByIDHandler{}
	}
	return handler.NewProjectHandler(upload, list, getByID)
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

func sampleProjectListViews() []domainproject.ProjectListView {
	return []domainproject.ProjectListView{
		{
			ID:        testutil.TestProjectID,
			Name:      "Demo",
			Status:    domainproject.StatusProcessing,
			CreatedAt: time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC),
		},
	}
}

func sampleProjectDetailView() *domainproject.ProjectDetailView {
	return &domainproject.ProjectDetailView{
		ID:        testutil.TestProjectID,
		Name:      "Demo",
		Status:    domainproject.StatusProcessing,
		CreatedAt: time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 9, 28, 11, 0, 0, 0, time.UTC),
		MediaFiles: []domainproject.ProjectMediaFileView{
			{
				ID:               testutil.TestMediaFileID,
				OriginalFilename: "demo.mp4",
				MimeType:         "video/mp4",
				SizeBytes:        1024,
				DurationMs:       5000,
				OriginalURL:      "https://cdn.example/original.mp4",
				ThumbnailURL:     "https://cdn.example/thumb.jpg",
				Status:           "uploaded",
				CreatedAt:        time.Date(2026, 9, 28, 10, 5, 0, 0, time.UTC),
			},
		},
	}
}

func unusedUUID() uuid.UUID {
	return uuid.MustParse("01960000-0000-7000-8000-000000000099")
}

var errUnexpected = errors.New("unexpected handler call")
