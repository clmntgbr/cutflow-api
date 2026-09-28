package handler

import (
	"errors"
	"log"

	cmdproject "go-api/internal/application/command/project"
	queryproject "go-api/internal/application/query/project"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	requestUploadURLHandler projectRequestUploadURLHandler
	listHandler             projectListHandler
	getByIDHandler          projectGetByIDHandler
}

func NewProjectHandler(
	requestUploadURLHandler projectRequestUploadURLHandler,
	listHandler projectListHandler,
	getByIDHandler projectGetByIDHandler,
) *ProjectHandler {
	return &ProjectHandler{
		requestUploadURLHandler: requestUploadURLHandler,
		listHandler:             listHandler,
		getByIDHandler:          getByIDHandler,
	}
}

func (h *ProjectHandler) List(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var listQuery paginate.PaginateQuery
	if err := validation.BindAndValidateQuery(c, &listQuery); err != nil {
		return err
	}

	sortBy, orderBy := listQuery.SortBy, listQuery.OrderBy
	listQuery.Normalize()
	if sortBy == "" {
		listQuery.SortBy = "created_at"
	}
	if orderBy == "" {
		listQuery.OrderBy = paginate.OrderByDesc
	}

	views, total, err := h.listHandler.Handle(c.Context(), queryproject.ListProjectsQuery{
		UserID: user.ID,
		Query:  listQuery,
	})
	if err != nil {
		log.Printf("failed to list projects: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list projects",
		})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewProjectListResponseFromViews(views),
		int(total),
		listQuery,
	))
}

func (h *ProjectHandler) GetByID(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid project id",
		})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryproject.GetProjectByIDQuery{
		ID:     projectID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, domainproject.ErrProjectNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Project not found",
			})
		}
		log.Printf("failed to get project: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get project",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewProjectDetailResponseFromView(*view))
}

func (h *ProjectHandler) RequestUploadURL(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var req dto.RequestUploadURLRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}

	result, err := h.requestUploadURLHandler.Handle(c.Context(), cmdproject.RequestUploadURLCommand{
		UserID:      user.ID,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		if errors.Is(err, domainmediafile.ErrInvalidFilename) || errors.Is(err, domainmediafile.ErrUnsupportedType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unsupported video type",
			})
		}
		if errors.Is(err, domainmediafile.ErrMediaTooLarge) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "File exceeds maximum size",
			})
		}
		log.Printf("failed to request upload url: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate upload url",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewRequestUploadURLResponse(result))
}
