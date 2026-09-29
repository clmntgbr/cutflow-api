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
	startedAt := time.Date(2026, 9, 28, 10, 6, 0, 0, time.UTC)
	return []domainproject.ProjectListView{
		{
			ID:           testutil.TestProjectID,
			Name:         "Demo",
			Status:       domainproject.StatusProcessing,
			CreatedAt:    time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC),
			MediaFileID:  testutil.TestMediaFileID,
			ThumbnailKey: "videos/" + testutil.TestMediaFileID.String() + "/thumbnail.jpg",
			Jobs: []domainproject.ProjectJobView{
				{
					ID:          testutil.TestJobID,
					MediaFileID: testutil.TestMediaFileID,
					Name:        "extract_audio",
					Status:      "processing",
					CreatedAt:   time.Date(2026, 9, 28, 10, 5, 30, 0, time.UTC),
					UpdatedAt:   startedAt,
					StartedAt:   &startedAt,
				},
			},
		},
	}
}

func sampleProjectDetailView() *domainproject.ProjectDetailView {
	startedAt := time.Date(2026, 9, 28, 10, 6, 0, 0, time.UTC)
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
				ThumbnailKey:     "videos/" + testutil.TestMediaFileID.String() + "/thumbnail.jpg",
				Status:           "uploaded",
				CreatedAt:        time.Date(2026, 9, 28, 10, 5, 0, 0, time.UTC),
			},
		},
		Jobs: []domainproject.ProjectJobView{
			{
				ID:          testutil.TestJobID,
				MediaFileID: testutil.TestMediaFileID,
				Name:        "extract_audio",
				Status:      "processing",
				CreatedAt:   time.Date(2026, 9, 28, 10, 5, 30, 0, time.UTC),
				UpdatedAt:   startedAt,
				StartedAt:   &startedAt,
			},
		},
		Timelines: []domainproject.ProjectTimelineView{
			{
				ID:            testutil.TestTimelineID,
				MediaFileID:   testutil.TestMediaFileID,
				Version:       1,
				DurationMs:    4200,
				Fingerprint:   "abc123",
				EngineVersion: "timeline-engine-v1",
				IsActive:      true,
				Segments: []domainproject.ProjectTimelineSegmentView{
					{
						Index:         0,
						MediaFileID:   testutil.TestMediaFileID,
						SourceStartMs: 0,
						SourceEndMs:   2000,
						OutputStartMs: 0,
						OutputEndMs:   2000,
					},
					{
						Index:         1,
						MediaFileID:   testutil.TestMediaFileID,
						SourceStartMs: 2800,
						SourceEndMs:   5000,
						OutputStartMs: 2000,
						OutputEndMs:   4200,
					},
				},
				Decisions: []domainproject.ProjectEditDecisionView{
					{
						ID:            uuid.MustParse("01960000-0000-7000-8000-000000000006"),
						MediaFileID:   testutil.TestMediaFileID,
						Type:          "silence",
						SourceStartMs: 2000,
						SourceEndMs:   2800,
						Action:        "remove",
						Source:        "automatic",
						Reasons:       []string{"silence"},
					},
				},
				CreatedAt: time.Date(2026, 9, 28, 10, 10, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 9, 28, 10, 10, 0, 0, time.UTC),
			},
		},
	}
}

func unusedUUID() uuid.UUID {
	return uuid.MustParse("01960000-0000-7000-8000-000000000099")
}

var errUnexpected = errors.New("unexpected handler call")
