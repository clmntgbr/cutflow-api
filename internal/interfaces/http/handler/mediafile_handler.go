package handler

import (
	"errors"
	"io"
	"log"
	"strconv"

	cmdmediafile "go-api/internal/application/command/mediafile"
	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
	domaintimeline "go-api/internal/domain/timeline"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MediaFileHandler struct {
	getOwnedHandler            mediaFileGetOwnedHandler
	getEditorHandler           mediaFileGetEditorHandler
	updateConfigurationHandler mediaFileUpdateConfigurationHandler
	decisionMutationHandler    mediaFileDecisionMutationHandler
	finalizeHandler            mediaFileFinalizeHandler
	storage                    mediaFileThumbnailStorage
}

func NewMediaFileHandler(
	getOwnedHandler mediaFileGetOwnedHandler,
	getEditorHandler mediaFileGetEditorHandler,
	updateConfigurationHandler mediaFileUpdateConfigurationHandler,
	decisionMutationHandler mediaFileDecisionMutationHandler,
	finalizeHandler mediaFileFinalizeHandler,
	storage mediaFileThumbnailStorage,
) *MediaFileHandler {
	return &MediaFileHandler{
		getOwnedHandler:            getOwnedHandler,
		getEditorHandler:           getEditorHandler,
		updateConfigurationHandler: updateConfigurationHandler,
		decisionMutationHandler:    decisionMutationHandler,
		finalizeHandler:            finalizeHandler,
		storage:                    storage,
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

func (h *MediaFileHandler) GetEditor(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	mediaFileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid media file id",
		})
	}

	view, err := h.getEditorHandler.Handle(c.Context(), querymediafile.GetEditorStateQuery{
		MediaFileID: mediaFileID,
		UserID:      user.ID,
	})
	if err != nil {
		if errors.Is(err, domainmediafile.ErrMediaNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Media file not found",
			})
		}
		if errors.Is(err, querymediafile.ErrEditorNotReady) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Editor is not ready",
				"code":    "EDITOR_NOT_READY",
				"status":  "ANALYZING",
			})
		}
		log.Printf("failed to get editor state: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to load editor state",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewEditorStateResponse(*view))
}

func (h *MediaFileHandler) UpdateEditor(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	mediaFileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid media file id",
		})
	}

	var req dto.UpdateEditorRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}

	switch req.Type {
	case dto.EditorActionUpdateConfiguration:
		return h.updateConfiguration(c, mediaFileID, user.ID, req)
	case dto.EditorActionUpdateSilenceConfiguration:
		return h.updateSilenceConfiguration(c, mediaFileID, user.ID, req)
	case dto.EditorActionOverrideDecision:
		return h.overrideDecision(c, mediaFileID, user.ID, req)
	case dto.EditorActionClearDecisionOverride:
		return h.clearDecisionOverride(c, mediaFileID, user.ID, req)
	case dto.EditorActionCreateManualCut:
		return h.createManualCut(c, mediaFileID, user.ID, req)
	default:
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors": fiber.Map{
				"type": "is invalid",
			},
		})
		return validation.ErrValidationFailed
	}
}

func (h *MediaFileHandler) Finalize(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	mediaFileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid media file id",
		})
	}

	var req dto.FinalizeEditorRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}
	timelineID := uuid.MustParse(req.TimelineID)

	result, err := h.finalizeHandler.Handle(c.Context(), cmdmediafile.FinalizeEditorCommand{
		MediaFileID:     mediaFileID,
		UserID:          user.ID,
		TimelineID:      timelineID,
		TimelineVersion: req.TimelineVersion,
	})
	if err != nil {
		return h.mapEditorMutationError(c, err)
	}

	return c.Status(fiber.StatusAccepted).JSON(presenter.NewFinalizeAcceptedResponse(
		result.JobID.String(),
		result.TimelineID.String(),
		result.PreviousTimelineVersion,
	))
}

func (h *MediaFileHandler) updateConfiguration(
	c fiber.Ctx,
	mediaFileID, userID uuid.UUID,
	req dto.UpdateEditorRequest,
) error {
	if req.Configuration == nil {
		return h.editorValidationError(c, fiber.Map{"configuration": "is required"})
	}
	if req.Configuration.Silence != nil {
		return h.editorValidationError(c, fiber.Map{
			"configuration.silence": "use update_silence_configuration",
		})
	}
	cmd := cmdmediafile.UpdateEditorConfigurationCommand{
		MediaFileID:     mediaFileID,
		UserID:          userID,
		TimelineVersion: req.TimelineVersion,
	}
	cfg := req.Configuration
	rebuild := false
	if cfg.Filler != nil {
		cmd.FillerEnabled = cfg.Filler.Enabled
		rebuild = true
	}
	if cfg.Repetition != nil {
		cmd.RepetitionEnabled = cfg.Repetition.Enabled
		rebuild = true
	}
	if cfg.Subtitles != nil {
		cmd.SubtitlesEnabled = cfg.Subtitles.Enabled
		cmd.SubtitleMaxWords = cfg.Subtitles.MaxWords
	}
	cmd.RebuildTimeline = rebuild

	result, err := h.updateConfigurationHandler.Handle(c.Context(), cmd)
	if err != nil {
		return h.mapEditorMutationError(c, err)
	}
	if rebuild && result != nil {
		return c.Status(fiber.StatusAccepted).JSON(presenter.NewEditorRebuildAcceptedResponse(
			result.JobID.String(),
			result.PreviousTimelineVersion,
		))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaFileHandler) updateSilenceConfiguration(
	c fiber.Ctx,
	mediaFileID, userID uuid.UUID,
	req dto.UpdateEditorRequest,
) error {
	if req.Configuration == nil || req.Configuration.Silence == nil {
		return h.editorValidationError(c, fiber.Map{"configuration.silence": "is required"})
	}
	s := req.Configuration.Silence
	cmd := cmdmediafile.UpdateEditorConfigurationCommand{
		MediaFileID:     mediaFileID,
		UserID:          userID,
		TimelineVersion: req.TimelineVersion,
		Silence: &cmdmediafile.SilenceConfigPatch{
			Enabled:         s.Enabled,
			ThresholdMode:   s.ThresholdMode,
			DetectionLevel:  s.DetectionLevel,
			MinDurationMs:   s.MinDurationMs,
			PaddingBeforeMs: s.PaddingBeforeMs,
			PaddingAfterMs:  s.PaddingAfterMs,
			ThresholdDB:     s.ThresholdDB,
		},
		RebuildTimeline: true,
	}

	result, err := h.updateConfigurationHandler.Handle(c.Context(), cmd)
	if err != nil {
		return h.mapEditorMutationError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(presenter.NewEditorRebuildAcceptedResponse(
		result.JobID.String(),
		result.PreviousTimelineVersion,
	))
}

func (h *MediaFileHandler) overrideDecision(
	c fiber.Ctx,
	mediaFileID, userID uuid.UUID,
	req dto.UpdateEditorRequest,
) error {
	fieldErrors := fiber.Map{}
	if req.DecisionID == nil || *req.DecisionID == "" {
		fieldErrors["decisionId"] = "is required"
	}
	if req.Action == nil || (*req.Action != domaintimeline.ActionKeep && *req.Action != domaintimeline.ActionRemove) {
		fieldErrors["action"] = "is invalid"
	}
	if len(fieldErrors) > 0 {
		return h.editorValidationError(c, fieldErrors)
	}
	decisionID, err := uuid.Parse(*req.DecisionID)
	if err != nil {
		return h.editorValidationError(c, fiber.Map{"decisionId": "must be a valid UUID"})
	}

	if err := h.decisionMutationHandler.Override(c.Context(), cmdmediafile.OverrideDecisionCommand{
		MediaFileID:     mediaFileID,
		UserID:          userID,
		DecisionID:      decisionID,
		Action:          *req.Action,
		TimelineVersion: req.TimelineVersion,
	}); err != nil {
		return h.mapEditorMutationError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaFileHandler) clearDecisionOverride(
	c fiber.Ctx,
	mediaFileID, userID uuid.UUID,
	req dto.UpdateEditorRequest,
) error {
	if req.DecisionID == nil || *req.DecisionID == "" {
		return h.editorValidationError(c, fiber.Map{"decisionId": "is required"})
	}
	decisionID, err := uuid.Parse(*req.DecisionID)
	if err != nil {
		return h.editorValidationError(c, fiber.Map{"decisionId": "must be a valid UUID"})
	}

	if err := h.decisionMutationHandler.ClearOverride(c.Context(), cmdmediafile.ClearDecisionOverrideCommand{
		MediaFileID:     mediaFileID,
		UserID:          userID,
		DecisionID:      decisionID,
		TimelineVersion: req.TimelineVersion,
	}); err != nil {
		return h.mapEditorMutationError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaFileHandler) createManualCut(
	c fiber.Ctx,
	mediaFileID, userID uuid.UUID,
	req dto.UpdateEditorRequest,
) error {
	fieldErrors := fiber.Map{}
	if req.SourceStartMs == nil {
		fieldErrors["sourceStartMs"] = "is required"
	}
	if req.SourceEndMs == nil {
		fieldErrors["sourceEndMs"] = "is required"
	}
	if req.SourceStartMs != nil && req.SourceEndMs != nil {
		if *req.SourceStartMs < 0 {
			fieldErrors["sourceStartMs"] = "must be greater than or equal to 0"
		}
		if *req.SourceEndMs <= *req.SourceStartMs {
			fieldErrors["sourceEndMs"] = "must be greater than sourceStartMs"
		}
	}
	if len(fieldErrors) > 0 {
		return h.editorValidationError(c, fieldErrors)
	}

	if err := h.decisionMutationHandler.CreateManual(c.Context(), cmdmediafile.CreateManualDecisionCommand{
		MediaFileID:     mediaFileID,
		UserID:          userID,
		Action:          domaintimeline.ActionRemove,
		SourceStartMs:   *req.SourceStartMs,
		SourceEndMs:     *req.SourceEndMs,
		TimelineVersion: req.TimelineVersion,
	}); err != nil {
		return h.mapEditorMutationError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaFileHandler) editorValidationError(c fiber.Ctx, fieldErrors fiber.Map) error {
	_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "Validation failed",
		"errors":  fieldErrors,
	})
	return validation.ErrValidationFailed
}

func (h *MediaFileHandler) mapEditorMutationError(c fiber.Ctx, err error) error {
	if errors.Is(err, domainmediafile.ErrMediaNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Media file not found",
		})
	}
	if errors.Is(err, querymediafile.ErrEditorNotReady) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Editor is not ready",
			"code":    "EDITOR_NOT_READY",
		})
	}
	if errors.Is(err, cmdmediafile.ErrDecisionNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Decision not found",
		})
	}
	if errors.Is(err, cmdmediafile.ErrInvalidManualRange) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid manual cut range",
			"errors": fiber.Map{
				"sourceEndMs": "must be within media duration",
			},
		})
	}
	var stale *cmdmediafile.StaleTimelineError
	if errors.As(err, &stale) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message":        "Timeline is stale",
			"code":           "STALE_TIMELINE",
			"currentVersion": stale.CurrentVersion,
		})
	}
	log.Printf("editor mutation failed: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": "Failed to apply editor mutation",
	})
}
