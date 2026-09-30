package mediafiletest

import (
	"bytes"
	"context"
	"errors"
	"io"

	cmdmediafile "go-api/internal/application/command/mediafile"
	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

type mockGetOwnedMediaFileHandler struct {
	called bool
	query  querymediafile.GetOwnedMediaFileQuery
	view   *domainmediafile.MediaFileThumbnailView
	err    error
}

func (m *mockGetOwnedMediaFileHandler) Handle(
	_ context.Context,
	q querymediafile.GetOwnedMediaFileQuery,
) (*domainmediafile.MediaFileThumbnailView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockGetEditorHandler struct {
	called bool
	query  querymediafile.GetEditorStateQuery
	view   *querymediafile.EditorStateView
	err    error
}

func (m *mockGetEditorHandler) Handle(
	_ context.Context,
	q querymediafile.GetEditorStateQuery,
) (*querymediafile.EditorStateView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockUpdateConfigurationHandler struct {
	called bool
	cmd    cmdmediafile.UpdateEditorConfigurationCommand
	result *cmdmediafile.EditorMutationResult
	err    error
}

func (m *mockUpdateConfigurationHandler) Handle(
	_ context.Context,
	cmd cmdmediafile.UpdateEditorConfigurationCommand,
) (*cmdmediafile.EditorMutationResult, error) {
	m.called = true
	m.cmd = cmd
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	if cmd.RebuildTimeline {
		return &cmdmediafile.EditorMutationResult{
			JobID:                   uuid.MustParse("01960000-0000-7000-8000-000000000020"),
			PreviousTimelineVersion: 4,
		}, nil
	}
	return nil, nil
}

type mockDecisionMutationHandler struct {
	overrideCalled bool
	clearCalled    bool
	createCalled   bool
	overrideCmd    cmdmediafile.OverrideDecisionCommand
	clearCmd       cmdmediafile.ClearDecisionOverrideCommand
	createCmd      cmdmediafile.CreateManualDecisionCommand
	err            error
}

func (m *mockDecisionMutationHandler) Override(
	_ context.Context,
	cmd cmdmediafile.OverrideDecisionCommand,
) error {
	m.overrideCalled = true
	m.overrideCmd = cmd
	return m.err
}

func (m *mockDecisionMutationHandler) ClearOverride(
	_ context.Context,
	cmd cmdmediafile.ClearDecisionOverrideCommand,
) error {
	m.clearCalled = true
	m.clearCmd = cmd
	return m.err
}

func (m *mockDecisionMutationHandler) CreateManual(
	_ context.Context,
	cmd cmdmediafile.CreateManualDecisionCommand,
) error {
	m.createCalled = true
	m.createCmd = cmd
	return m.err
}

type mockFinalizeHandler struct {
	called bool
	cmd    cmdmediafile.FinalizeEditorCommand
	result *cmdmediafile.FinalizeEditorResult
	err    error
}

func (m *mockFinalizeHandler) Handle(
	_ context.Context,
	cmd cmdmediafile.FinalizeEditorCommand,
) (*cmdmediafile.FinalizeEditorResult, error) {
	m.called = true
	m.cmd = cmd
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &cmdmediafile.FinalizeEditorResult{
		Status:                  "accepted",
		JobID:                   uuid.MustParse("01960000-0000-7000-8000-000000000020"),
		TimelineID:              testutil.TestTimelineID,
		PreviousTimelineVersion: cmd.TimelineVersion,
	}, nil
}

type mockThumbnailStorage struct {
	called bool
	key    string
	body   []byte
	err    error
}

func (m *mockThumbnailStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	m.called = true
	m.key = key
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.body)), nil
}

type readFailCloser struct{}

func (readFailCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (readFailCloser) Close() error             { return nil }

type mockFailingReadStorage struct{}

func (mockFailingReadStorage) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return readFailCloser{}, nil
}

type thumbnailStorage interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

func newMediaFileHandler(
	owned *mockGetOwnedMediaFileHandler,
	storage thumbnailStorage,
) *handler.MediaFileHandler {
	return newMediaFileHandlerFull(owned, nil, nil, nil, nil, storage)
}

func newMediaFileHandlerFull(
	owned *mockGetOwnedMediaFileHandler,
	editor *mockGetEditorHandler,
	updateCfg *mockUpdateConfigurationHandler,
	decisions *mockDecisionMutationHandler,
	finalize *mockFinalizeHandler,
	storage thumbnailStorage,
) *handler.MediaFileHandler {
	if owned == nil {
		owned = &mockGetOwnedMediaFileHandler{}
	}
	if editor == nil {
		editor = &mockGetEditorHandler{}
	}
	if updateCfg == nil {
		updateCfg = &mockUpdateConfigurationHandler{}
	}
	if decisions == nil {
		decisions = &mockDecisionMutationHandler{}
	}
	if finalize == nil {
		finalize = &mockFinalizeHandler{}
	}
	if storage == nil {
		storage = &mockThumbnailStorage{body: []byte{0xff, 0xd8, 0xff}}
	}
	return handler.NewMediaFileHandler(owned, editor, updateCfg, decisions, finalize, storage)
}

func sampleThumbnailView() *domainmediafile.MediaFileThumbnailView {
	return &domainmediafile.MediaFileThumbnailView{
		ID:           testutil.TestMediaFileID,
		ThumbnailKey: "videos/" + testutil.TestMediaFileID.String() + "/thumbnail.jpg",
	}
}

func sampleEditorStateView() *querymediafile.EditorStateView {
	w, h := 1920, 1080
	remove := "remove"
	return &querymediafile.EditorStateView{
		Media: querymediafile.EditorMediaView{
			ID:         testutil.TestMediaFileID,
			Name:       "demo.mp4",
			URL:        "https://cdn.example/original.mp4",
			DurationMs: 5000,
			Width:      &w,
			Height:     &h,
		},
		Timeline: querymediafile.EditorTimelineView{
			ID:         testutil.TestTimelineID,
			Version:    1,
			DurationMs: 4200,
			Segments: []querymediafile.EditorTimelineSegmentView{
				{
					ID:            uuid.MustParse("01960000-0000-7000-8000-000000000010"),
					Index:         0,
					SourceStartMs: 0,
					SourceEndMs:   2000,
					OutputStartMs: 0,
					OutputEndMs:   2000,
				},
			},
		},
		Configuration: querymediafile.EditorConfigView{
			Silence: querymediafile.EditorSilenceConfigView{
				Enabled:        true,
				ThresholdMode:  "auto",
				DetectionLevel: "aggressive",
				MinDurationMs:  500,
			},
			Filler:     querymediafile.EditorToggleConfigView{Enabled: true},
			Repetition: querymediafile.EditorToggleConfigView{Enabled: true},
			Subtitles:  querymediafile.EditorSubtitlesConfigView{Enabled: true, MaxWords: 3, StyleID: "default"},
			Output:     querymediafile.EditorOutputConfigView{AspectRatio: "original"},
		},
		Decisions: []querymediafile.EditorDecisionView{
			{
				ID:              uuid.MustParse("01960000-0000-7000-8000-000000000011"),
				Type:            "silence",
				SourceStartMs:   2000,
				SourceEndMs:     2800,
				AutomaticAction: &remove,
				EffectiveAction: "remove",
			},
		},
		Subtitles: querymediafile.EditorSubtitlesView{
			Words: []querymediafile.EditorWordView{
				{ID: uuid.MustParse("01960000-0000-7000-8000-000000000012"), Text: "Bonjour", SourceStartMs: 100, SourceEndMs: 400},
			},
		},
		ViralCandidates: []querymediafile.EditorViralCandidateView{},
	}
}

func unusedUUID() uuid.UUID {
	return uuid.MustParse("01960000-0000-7000-8000-000000000099")
}

func editorPath() string {
	return "/media-files/" + testutil.TestMediaFileID.String() + "/editor"
}

func finalizePath() string {
	return "/media-files/" + testutil.TestMediaFileID.String() + "/finalize"
}

var errUnexpected = errors.New("unexpected handler call")
