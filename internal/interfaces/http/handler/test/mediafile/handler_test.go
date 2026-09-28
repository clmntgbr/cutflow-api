package mediafiletest

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/interfaces/http/testutil"
)

func TestMediaFileHandler_GetThumbnail_Success(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{view: sampleThumbnailView()}
	storage := &mockThumbnailStorage{body: []byte{0xff, 0xd8, 0xff, 0xd9}}
	h := newMediaFileHandler(owned, storage)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("content-type: got %q", ct)
	}
	if !owned.called {
		t.Fatal("expected owned handler to be called")
	}
	if owned.query.ID != testutil.TestMediaFileID {
		t.Fatalf("media id: got %s", owned.query.ID)
	}
	if !storage.called {
		t.Fatal("expected storage get to be called")
	}
	if storage.key != owned.view.ThumbnailKey {
		t.Fatalf("storage key: got %q", storage.key)
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !bytes.Equal(body, storage.body) {
		t.Fatalf("body: got %v want %v", body, storage.body)
	}
}

func TestMediaFileHandler_GetThumbnail_Unauthorized(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{}
	h := newMediaFileHandler(owned, nil)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
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
	if owned.called {
		t.Fatal("owned handler must not be called without auth")
	}
}

func TestMediaFileHandler_GetThumbnail_InvalidInput(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{}
	h := newMediaFileHandler(owned, nil)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/not-a-uuid/thumbnail", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if owned.called {
		t.Fatal("owned handler must not be called with invalid id")
	}
}

func TestMediaFileHandler_GetThumbnail_HandlerError_NotFound(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{err: domainmediafile.ErrMediaNotFound}
	h := newMediaFileHandler(owned, nil)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+unusedUUID().String()+"/thumbnail", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestMediaFileHandler_GetThumbnail_HandlerError_EmptyThumbnailKey(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{
		view: &domainmediafile.MediaFileThumbnailView{
			ID:           testutil.TestMediaFileID,
			ThumbnailKey: "",
		},
	}
	storage := &mockThumbnailStorage{body: []byte{1}}
	h := newMediaFileHandler(owned, storage)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if storage.called {
		t.Fatal("storage must not be called without thumbnail key")
	}
}

func TestMediaFileHandler_GetThumbnail_HandlerError_StorageMiss(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{view: sampleThumbnailView()}
	storage := &mockThumbnailStorage{err: errors.New("object missing")}
	h := newMediaFileHandler(owned, storage)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestMediaFileHandler_GetThumbnail_HandlerError_Internal(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{err: errUnexpected}
	h := newMediaFileHandler(owned, nil)

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
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
}

func TestMediaFileHandler_GetThumbnail_HandlerError_ReadBody(t *testing.T) {
	owned := &mockGetOwnedMediaFileHandler{view: sampleThumbnailView()}
	h := newMediaFileHandler(owned, mockFailingReadStorage{})

	app := testutil.NewTestApp()
	app.Get("/media-files/:id/thumbnail", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetThumbnail)

	req, err := testutil.JSONRequest(http.MethodGet, "/media-files/"+testutil.TestMediaFileID.String()+"/thumbnail", nil)
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
}
