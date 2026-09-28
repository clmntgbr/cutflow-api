package projecttest

import (
	"net/http"
	"testing"

	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestProjectHandler_List_Success(t *testing.T) {
	list := &mockListProjectsHandler{
		views: sampleProjectListViews(),
		total: 1,
	}
	h := newProjectHandler(nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithUserWithoutProject(testutil.TestUserID), h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects", nil)
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
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if list.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", list.query.UserID)
	}
	if list.query.Query.SortBy != "created_at" {
		t.Fatalf("sortBy: got %q", list.query.Query.SortBy)
	}
	if list.query.Query.OrderBy != paginate.OrderByDesc {
		t.Fatalf("orderBy: got %q", list.query.Query.OrderBy)
	}

	var out paginate.PaginateResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.Total != 1 {
		t.Fatalf("total: got %d", out.Total)
	}
	members, ok := out.Members.([]any)
	if !ok || len(members) != 1 {
		t.Fatalf("members: got %#v", out.Members)
	}
	item, ok := members[0].(map[string]any)
	if !ok {
		t.Fatalf("member item: got %#v", members[0])
	}
	if item["id"] != testutil.TestProjectID.String() {
		t.Fatalf("id: got %v", item["id"])
	}
	if item["name"] != "Demo" {
		t.Fatalf("name: got %v", item["name"])
	}
	if _, hasUpdatedAt := item["updatedAt"]; hasUpdatedAt {
		t.Fatal("list item must not include updatedAt")
	}
	if _, hasMediaFiles := item["mediaFiles"]; hasMediaFiles {
		t.Fatal("list item must not include mediaFiles")
	}
}

func TestProjectHandler_List_Unauthorized(t *testing.T) {
	list := &mockListProjectsHandler{}
	h := newProjectHandler(nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects", nil)
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
	if list.called {
		t.Fatal("list handler must not be called without auth")
	}
}

func TestProjectHandler_List_InvalidInput(t *testing.T) {
	list := &mockListProjectsHandler{}
	h := newProjectHandler(nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithUserWithoutProject(testutil.TestUserID), h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects?orderBy=sideways", nil)
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
	if list.called {
		t.Fatal("list handler must not be called with invalid query")
	}
}

func TestProjectHandler_List_HandlerError_Internal(t *testing.T) {
	list := &mockListProjectsHandler{err: errUnexpected}
	h := newProjectHandler(nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithUserWithoutProject(testutil.TestUserID), h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects", nil)
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
	if body["message"] != "Failed to list projects" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestProjectHandler_GetByID_Success(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{view: sampleProjectDetailView()}
	h := newProjectHandler(nil, nil, getByID)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil)
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
	if !getByID.called {
		t.Fatal("expected get handler to be called")
	}
	if getByID.query.ID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", getByID.query.ID)
	}
	if getByID.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", getByID.query.UserID)
	}

	var out presenter.ProjectDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestProjectID.String() {
		t.Fatalf("id: got %s", out.ID)
	}
	if out.Name != "Demo" {
		t.Fatalf("name: got %s", out.Name)
	}
	if len(out.MediaFiles) != 1 {
		t.Fatalf("mediaFiles: got %d", len(out.MediaFiles))
	}
	if out.MediaFiles[0].OriginalURL == nil || *out.MediaFiles[0].OriginalURL == "" {
		t.Fatal("expected original url")
	}
}

func TestProjectHandler_GetByID_Unauthorized(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{}
	h := newProjectHandler(nil, nil, getByID)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil)
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
	if getByID.called {
		t.Fatal("get handler must not be called without auth")
	}
}

func TestProjectHandler_GetByID_InvalidInput(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{}
	h := newProjectHandler(nil, nil, getByID)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects/not-a-uuid", nil)
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
	if getByID.called {
		t.Fatal("get handler must not be called with invalid id")
	}
}

func TestProjectHandler_GetByID_HandlerError_NotFound(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{err: domainproject.ErrProjectNotFound}
	h := newProjectHandler(nil, nil, getByID)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects/"+unusedUUID().String(), nil)
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

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Project not found" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestProjectHandler_GetByID_HandlerError_Internal(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{err: errUnexpected}
	h := newProjectHandler(nil, nil, getByID)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil)
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
	if body["message"] != "Failed to get project" {
		t.Fatalf("message: got %v", body["message"])
	}
}
