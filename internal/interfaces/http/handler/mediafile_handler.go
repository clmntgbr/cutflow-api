package handler

import (
	"errors"
	"io"
	"strconv"

	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
	httpctx "go-api/internal/interfaces/http/context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MediaFileHandler struct {
	getOwnedHandler mediaFileGetOwnedHandler
	storage         mediaFileThumbnailStorage
}

func NewMediaFileHandler(
	getOwnedHandler mediaFileGetOwnedHandler,
	storage mediaFileThumbnailStorage,
) *MediaFileHandler {
	return &MediaFileHandler{
		getOwnedHandler: getOwnedHandler,
		storage:         storage,
	}
}

func (h *MediaFileHandler) GetThumbnail(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	mediaFileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	view, err := h.getOwnedHandler.Handle(c.Context(), querymediafile.GetOwnedMediaFileQuery{
		ID:     mediaFileID,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, domainmediafile.ErrMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if view.ThumbnailKey == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	reader, err := h.storage.Get(c.Context(), view.ThumbnailKey)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Content-Type", "image/jpeg")
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("Content-Length", strconv.Itoa(len(body)))
	return c.Send(body)
}
