package project

import (
	"context"
	"fmt"

	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ListProjectsQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
}

type ListProjectsHandler struct {
	readRepo domainproject.ProjectReadRepository
}

func NewListProjectsHandler(readRepo domainproject.ProjectReadRepository) *ListProjectsHandler {
	return &ListProjectsHandler{readRepo: readRepo}
}

func (h *ListProjectsHandler) Handle(
	ctx context.Context,
	q ListProjectsQuery,
) ([]domainproject.ProjectListView, int64, error) {
	views, total, err := h.readRepo.List(ctx, q.UserID, q.Query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	return views, total, nil
}
