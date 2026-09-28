package projecttest

import (
	"net/http"
	"testing"

	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestProjectHandler_RequestUploadURL_Success(t *testing.T) {
	upload := &mockRequestUploadURLHandler{result: sampleUploadResult()}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !upload.called {
		t.Fatal("expected upload handler to be called")
	}
	if upload.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", upload.cmd.UserID)
	}
	if upload.cmd.Filename != "demo.mp4" {
		t.Fatalf("filename: got %q", upload.cmd.Filename)
	}

	var out presenter.RequestUploadURLResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ProjectID != testutil.TestProjectID.String() {
		t.Fatalf("project id: got %s", out.ProjectID)
	}
	if out.MediaFileID != testutil.TestMediaFileID.String() {
		t.Fatalf("media file id: got %s", out.MediaFileID)
	}
	if out.UploadURL == "" {
		t.Fatal("expected upload url")
	}
}

func TestProjectHandler_RequestUploadURL_Unauthorized(t *testing.T) {
	upload := &mockRequestUploadURLHandler{}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if upload.called {
		t.Fatal("upload handler must not be called without auth")
	}
}

func TestProjectHandler_RequestUploadURL_InvalidInput(t *testing.T) {
	h := newProjectHandler(nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", map[string]any{"filename": ""})
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestProjectHandler_RequestUploadURL_HandlerError_UnsupportedType(t *testing.T) {
	upload := &mockRequestUploadURLHandler{err: domainmediafile.ErrUnsupportedType}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Unsupported video type" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestProjectHandler_RequestUploadURL_HandlerError_InvalidFilename(t *testing.T) {
	upload := &mockRequestUploadURLHandler{err: domainmediafile.ErrInvalidFilename}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestProjectHandler_RequestUploadURL_HandlerError_TooLarge(t *testing.T) {
	upload := &mockRequestUploadURLHandler{err: domainmediafile.ErrMediaTooLarge}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "File exceeds maximum size" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestProjectHandler_RequestUploadURL_HandlerError_Internal(t *testing.T) {
	upload := &mockRequestUploadURLHandler{err: errUnexpected}
	h := newProjectHandler(upload, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects/upload-url", testutil.WithUserWithoutProject(testutil.TestUserID), h.RequestUploadURL)

	req, err := testutil.JSONRequest(http.MethodPost, "/projects/upload-url", validUploadBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to generate upload url" {
		t.Fatalf("message: got %v", body["message"])
	}
}
