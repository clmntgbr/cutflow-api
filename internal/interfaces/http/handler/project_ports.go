package handler

import (
	"context"

	cmdproject "go-api/internal/application/command/project"
	queryproject "go-api/internal/application/query/project"
	domainproject "go-api/internal/domain/project"
)

type projectRequestUploadURLHandler interface {
	Handle(ctx context.Context, cmd cmdproject.RequestUploadURLCommand) (*cmdproject.RequestUploadURLResult, error)
}

type projectListHandler interface {
	Handle(ctx context.Context, q queryproject.ListProjectsQuery) ([]domainproject.ProjectListView, int64, error)
}

type projectGetByIDHandler interface {
	Handle(ctx context.Context, q queryproject.GetProjectByIDQuery) (*domainproject.ProjectDetailView, error)
}
