package mediafiletest

import (
	"bytes"
	"net/http"
	"testing"

	cmdmediafile "go-api/internal/application/command/mediafile"
	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestMediaFileHandler_GetEditor_Success(t *testing.T) {
	editor := &mockGetEditorHandler{view: sampleEditorStateView()}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetEditor)

	req, err := testutil.JSONRequest(http.MethodGet, editorPath(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !editor.called {
		t.Fatal("expected editor handler call")
	}
	var out presenter.EditorStateResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.Media.ID != testutil.TestMediaFileID.String() {
		t.Fatalf("media id: got %s", out.Media.ID)
	}
	if out.Timeline.Version != 1 || len(out.Timeline.Segments) != 1 {
		t.Fatalf("timeline: %#v", out.Timeline)
	}
	if len(out.Decisions) != 1 || out.Decisions[0].Type != "silence" {
		t.Fatalf("decisions: %#v", out.Decisions)
	}
}

func TestMediaFileHandler_GetEditor_Unauthorized(t *testing.T) {
	editor := &mockGetEditorHandler{}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", h.GetEditor)
	req, _ := testutil.JSONRequest(http.MethodGet, editorPath(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if editor.called {
		t.Fatal("must not call editor without auth")
	}
}

func TestMediaFileHandler_GetEditor_InvalidInput(t *testing.T) {
	editor := &mockGetEditorHandler{}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetEditor)
	req, _ := testutil.JSONRequest(http.MethodGet, "/media-files/not-a-uuid/editor", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_GetEditor_NotFound(t *testing.T) {
	editor := &mockGetEditorHandler{err: domainmediafile.ErrMediaNotFound}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetEditor)
	req, _ := testutil.JSONRequest(http.MethodGet, editorPath(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_GetEditor_NotReady(t *testing.T) {
	editor := &mockGetEditorHandler{err: querymediafile.ErrEditorNotReady}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetEditor)
	req, _ := testutil.JSONRequest(http.MethodGet, editorPath(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "EDITOR_NOT_READY" {
		t.Fatalf("code: got %v", body["code"])
	}
}

func TestMediaFileHandler_GetEditor_Internal(t *testing.T) {
	editor := &mockGetEditorHandler{err: errUnexpected}
	h := newMediaFileHandlerFull(nil, editor, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetEditor)
	req, _ := testutil.JSONRequest(http.MethodGet, editorPath(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_Success(t *testing.T) {
	update := &mockUpdateConfigurationHandler{}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	enabled := false
	version := 4
	req, err := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionUpdateConfiguration,
		"timelineVersion": version,
		"configuration": map[string]any{
			"filler": map[string]any{"enabled": enabled},
		},
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !update.called || update.cmd.FillerEnabled == nil || *update.cmd.FillerEnabled != false {
		t.Fatalf("cmd: %#v", update.cmd)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateSilenceConfiguration_Success(t *testing.T) {
	update := &mockUpdateConfigurationHandler{}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	minMs := 800
	version := 4
	req, err := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionUpdateSilenceConfiguration,
		"timelineVersion": version,
		"configuration": map[string]any{
			"silence": map[string]any{"minDurationMs": minMs},
		},
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !update.called {
		t.Fatalf("cmd: %#v", update.cmd)
	}
	if update.cmd.Silence == nil || update.cmd.Silence.MinDurationMs == nil || *update.cmd.Silence.MinDurationMs != 800 {
		t.Fatalf("silence: %#v", update.cmd.Silence)
	}
	var out presenter.EditorRebuildAcceptedResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.Status != "accepted" || out.JobID == "" || out.PreviousTimelineVersion != 4 {
		t.Fatalf("response: %#v", out)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_AcceptsSilence(t *testing.T) {
	update := &mockUpdateConfigurationHandler{}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	level := "low"
	enabled := false
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"silence": map[string]any{
				"detectionLevel":  level,
				"minDurationMs":   1200,
				"paddingBeforeMs": 240,
				"paddingAfterMs":  280,
			},
			"filler": map[string]any{"enabled": enabled},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !update.called || update.cmd.Silence == nil || update.cmd.Silence.DetectionLevel == nil || *update.cmd.Silence.DetectionLevel != "low" {
		t.Fatalf("cmd: %#v", update.cmd)
	}
	if update.cmd.FillerEnabled == nil || *update.cmd.FillerEnabled != false {
		t.Fatalf("filler: %#v", update.cmd.FillerEnabled)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateSilenceConfiguration_MissingSilence(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionUpdateSilenceConfiguration,
		"configuration": map[string]any{},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_HandlerError(t *testing.T) {
	update := &mockUpdateConfigurationHandler{err: errUnexpected}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"filler": map[string]any{"enabled": true},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_WithToggles(t *testing.T) {
	update := &mockUpdateConfigurationHandler{}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	enabled := false
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"filler":     map[string]any{"enabled": enabled},
			"repetition": map[string]any{"enabled": enabled},
			"subtitles":  map[string]any{"enabled": enabled, "maxWords": 5},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if update.cmd.FillerEnabled == nil || *update.cmd.FillerEnabled != false {
		t.Fatalf("filler: %#v", update.cmd.FillerEnabled)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_SubtitlesOnly_NoRebuild(t *testing.T) {
	update := &mockUpdateConfigurationHandler{}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	enabled := false
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"subtitles": map[string]any{"enabled": enabled, "maxWords": 5},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if update.cmd.Silence != nil || update.cmd.FillerEnabled != nil || update.cmd.RepetitionEnabled != nil {
		t.Fatalf("unexpected timeline fields: %#v", update.cmd)
	}
}

func TestMediaFileHandler_UpdateEditor_OverrideDecision_Success(t *testing.T) {
	decisions := &mockDecisionMutationHandler{}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	decisionID := "01960000-0000-7000-8000-000000000011"
	req, err := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionOverrideDecision,
		"decisionId":      decisionID,
		"action":          "keep",
		"timelineVersion": 4,
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !decisions.overrideCalled || decisions.overrideCmd.Action != "keep" {
		t.Fatalf("override: %#v", decisions.overrideCmd)
	}
}

func TestMediaFileHandler_UpdateEditor_ClearDecisionOverride_Success(t *testing.T) {
	decisions := &mockDecisionMutationHandler{}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	decisionID := "01960000-0000-7000-8000-000000000011"
	req, err := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionClearDecisionOverride,
		"decisionId":      decisionID,
		"timelineVersion": 5,
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !decisions.clearCalled {
		t.Fatal("expected clear call")
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_Success(t *testing.T) {
	decisions := &mockDecisionMutationHandler{}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, err := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionCreateManualCut,
		"sourceStartMs":   120500,
		"sourceEndMs":     126200,
		"timelineVersion": 5,
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !decisions.createCalled || decisions.createCmd.Action != "remove" {
		t.Fatalf("create: %#v", decisions.createCmd)
	}
}

func TestMediaFileHandler_UpdateEditor_Unauthorized(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"silence": map[string]any{"minDurationMs": 800},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_InvalidID(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, "/media-files/bad/editor", map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
		"configuration": map[string]any{
			"silence": map[string]any{"minDurationMs": 800},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UnsupportedType(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": "select_viral_candidate",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_MissingType(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"configuration": map[string]any{},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_InvalidBody(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, err := http.NewRequest(http.MethodPatch, editorPath(), bytes.NewBufferString("{"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_MissingConfiguration(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionUpdateConfiguration,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_UpdateConfiguration_StaleTimeline(t *testing.T) {
	update := &mockUpdateConfigurationHandler{err: &cmdmediafile.StaleTimelineError{CurrentVersion: 6}}
	h := newMediaFileHandlerFull(nil, nil, update, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":            dto.EditorActionUpdateSilenceConfiguration,
		"timelineVersion": 5,
		"configuration": map[string]any{
			"silence": map[string]any{"minDurationMs": 800},
		},
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "STALE_TIMELINE" {
		t.Fatalf("code: got %v", body["code"])
	}
}

func TestMediaFileHandler_UpdateEditor_OverrideDecision_NotFound(t *testing.T) {
	decisions := &mockDecisionMutationHandler{err: cmdmediafile.ErrDecisionNotFound}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":       dto.EditorActionOverrideDecision,
		"decisionId": "01960000-0000-7000-8000-000000000011",
		"action":     "keep",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_OverrideDecision_InvalidFields(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":   dto.EditorActionOverrideDecision,
		"action": "nope",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_OverrideDecision_InvalidDecisionID(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":       dto.EditorActionOverrideDecision,
		"decisionId": "not-a-uuid",
		"action":     "keep",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_ClearDecisionOverride_InvalidDecisionID(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":       dto.EditorActionClearDecisionOverride,
		"decisionId": "bad",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_ClearDecisionOverride_MissingDecisionID(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionClearDecisionOverride,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_ClearDecisionOverride_MediaNotFound(t *testing.T) {
	decisions := &mockDecisionMutationHandler{err: domainmediafile.ErrMediaNotFound}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":       dto.EditorActionClearDecisionOverride,
		"decisionId": "01960000-0000-7000-8000-000000000011",
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_InvalidRange(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionCreateManualCut,
		"sourceStartMs": 2000,
		"sourceEndMs":   1000,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_MissingFields(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type": dto.EditorActionCreateManualCut,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_NegativeStart(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionCreateManualCut,
		"sourceStartMs": -1,
		"sourceEndMs":   100,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_EditorNotReady(t *testing.T) {
	decisions := &mockDecisionMutationHandler{err: querymediafile.ErrEditorNotReady}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionCreateManualCut,
		"sourceStartMs": 1000,
		"sourceEndMs":   2000,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_InvalidManualRange(t *testing.T) {
	decisions := &mockDecisionMutationHandler{err: cmdmediafile.ErrInvalidManualRange}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionCreateManualCut,
		"sourceStartMs": 1000,
		"sourceEndMs":   999999,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_UpdateEditor_CreateManualCut_Internal(t *testing.T) {
	decisions := &mockDecisionMutationHandler{err: errUnexpected}
	h := newMediaFileHandlerFull(nil, nil, nil, decisions, nil, nil)
	app := testutil.NewTestApp()
	app.Patch("/media-files/:id/editor", testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdateEditor)
	req, _ := testutil.JSONRequest(http.MethodPatch, editorPath(), map[string]any{
		"type":          dto.EditorActionCreateManualCut,
		"sourceStartMs": 1000,
		"sourceEndMs":   2000,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_Finalize_Success(t *testing.T) {
	finalize := &mockFinalizeHandler{}
	h := newMediaFileHandlerFull(nil, nil, nil, nil, finalize, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", testutil.WithUserWithoutProject(testutil.TestUserID), h.Finalize)
	req, err := testutil.JSONRequest(http.MethodPost, finalizePath(), map[string]any{
		"timelineId":      testutil.TestTimelineID.String(),
		"timelineVersion": 8,
		"output": map[string]any{
			"aspectRatio": "9:16",
			"resolution":  "1080x1920",
		},
	})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	if !finalize.called || finalize.cmd.TimelineVersion != 8 {
		t.Fatalf("cmd: %#v", finalize.cmd)
	}
	var out presenter.FinalizeAcceptedResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.Status != "accepted" || out.PreviousTimelineVersion != 8 || out.JobID == "" {
		t.Fatalf("response: %#v", out)
	}
}

func TestMediaFileHandler_Finalize_Unauthorized(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", h.Finalize)
	req, _ := testutil.JSONRequest(http.MethodPost, finalizePath(), map[string]any{
		"timelineId":      testutil.TestTimelineID.String(),
		"timelineVersion": 1,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_Finalize_InvalidID(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", testutil.WithUserWithoutProject(testutil.TestUserID), h.Finalize)
	req, _ := testutil.JSONRequest(http.MethodPost, "/media-files/bad/finalize", map[string]any{
		"timelineId":      testutil.TestTimelineID.String(),
		"timelineVersion": 1,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_Finalize_InvalidBody(t *testing.T) {
	h := newMediaFileHandlerFull(nil, nil, nil, nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", testutil.WithUserWithoutProject(testutil.TestUserID), h.Finalize)
	req, _ := testutil.JSONRequest(http.MethodPost, finalizePath(), map[string]any{
		"timelineVersion": 1,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}

func TestMediaFileHandler_Finalize_StaleTimeline(t *testing.T) {
	finalize := &mockFinalizeHandler{err: &cmdmediafile.StaleTimelineError{CurrentVersion: 9}}
	h := newMediaFileHandlerFull(nil, nil, nil, nil, finalize, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", testutil.WithUserWithoutProject(testutil.TestUserID), h.Finalize)
	req, _ := testutil.JSONRequest(http.MethodPost, finalizePath(), map[string]any{
		"timelineId":      testutil.TestTimelineID.String(),
		"timelineVersion": 8,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "STALE_TIMELINE" {
		t.Fatalf("code: got %v", body["code"])
	}
}

func TestMediaFileHandler_Finalize_NotFound(t *testing.T) {
	finalize := &mockFinalizeHandler{err: domainmediafile.ErrMediaNotFound}
	h := newMediaFileHandlerFull(nil, nil, nil, nil, finalize, nil)
	app := testutil.NewTestApp()
	app.Post("/media-files/:id/finalize", testutil.WithUserWithoutProject(testutil.TestUserID), h.Finalize)
	req, _ := testutil.JSONRequest(http.MethodPost, finalizePath(), map[string]any{
		"timelineId":      testutil.TestTimelineID.String(),
		"timelineVersion": 1,
	})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
}
