package handler

import (
	"errors"
	"log"

	cmdproject "go-api/internal/application/command/project"
	domainmediafile "go-api/internal/domain/mediafile"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	requestUploadURLHandler projectRequestUploadURLHandler
}

func NewProjectHandler(requestUploadURLHandler projectRequestUploadURLHandler) *ProjectHandler {
	return &ProjectHandler{requestUploadURLHandler: requestUploadURLHandler}
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
