package mediafile

import (
	"context"
	"log"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
	domainmediaconfig "go-api/internal/domain/mediaconfig"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domaintranscript "go-api/internal/domain/transcript"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

type AnalyzeTranscriptCommand struct {
	MediaFileID  uuid.UUID
	TranscriptID uuid.UUID
	ProjectID    uuid.UUID
	UserID       uuid.UUID
	JobID        uuid.UUID
}

type AnalyzeTranscriptHandler struct {
	transcriptRepo domaintranscript.TranscriptWriteRepository
	issueRepo      domaintranscriptissue.IssueWriteRepository
	configRepo     domainmediaconfig.MediaConfigurationWriteRepository
	jobRepo        domainjob.JobWriteRepository
	outbox         port.OutboxRepository
}

func NewAnalyzeTranscriptHandler(
	transcriptRepo domaintranscript.TranscriptWriteRepository,
	issueRepo domaintranscriptissue.IssueWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
) *AnalyzeTranscriptHandler {
	return &AnalyzeTranscriptHandler{
		transcriptRepo: transcriptRepo,
		issueRepo:      issueRepo,
		configRepo:     configRepo,
		jobRepo:        jobRepo,
		outbox:         outbox,
	}
}

func (h *AnalyzeTranscriptHandler) Handle(ctx context.Context, cmd AnalyzeTranscriptCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	existing, err := h.issueRepo.CountByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing > 0 {
		return h.markJobSuccess(ctx, cmd.JobID)
	}

	transcript, err := h.transcriptRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if transcript == nil || transcript.Status != domaintranscript.StatusCompleted {
		err := domainmediafile.ErrMediaNotFound
		if failErr := h.fail(ctx, cmd.JobID, "transcript not ready"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	words, err := h.transcriptRepo.ListWords(ctx, transcript.ID)
	if err != nil {
		return messaging.Retryable(err)
	}

	cfg, err := h.configRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	detectFillers := true
	detectRepetitions := true
	detectFalseStarts := true
	if cfg != nil {
		detectFillers = cfg.FillerRemovalEnabled
		detectRepetitions = cfg.RepetitionRemovalEnabled
		detectFalseStarts = cfg.RepetitionRemovalEnabled
	}

	language := transcript.Language
	issues := domaintranscriptissue.Detect(words, domaintranscriptissue.DetectOptions{
		DetectFillers:     detectFillers,
		DetectRepetitions: detectRepetitions,
		DetectFalseStarts: detectFalseStarts,
		Language:          language,
	})

	for _, issue := range issues {
		issue.ID = uuid.New()
		issue.MediaFileID = cmd.MediaFileID
		issue.TranscriptID = transcript.ID
		issue.CreatedAt = time.Now().UTC()
	}

	updatedWords := domaintranscriptissue.ApplyKinds(words, issues)

	log.Printf(
		"transcript analysis mediaFileId=%s fillers=%d repetitions=%d falseStarts=%d words=%d",
		cmd.MediaFileID,
		countType(issues, domaintranscriptissue.TypeFiller),
		countType(issues, domaintranscriptissue.TypeRepetition),
		countType(issues, domaintranscriptissue.TypeFalseStart),
		len(words),
	)

	return h.persist(ctx, cmd, transcript.ID, issues, updatedWords)
}

func countType(issues []*domaintranscriptissue.Issue, typ string) int {
	n := 0
	for _, issue := range issues {
		if issue.Type == typ {
			n++
		}
	}
	return n
}

func (h *AnalyzeTranscriptHandler) persist(
	ctx context.Context,
	cmd AnalyzeTranscriptCommand,
	transcriptID uuid.UUID,
	issues []*domaintranscriptissue.Issue,
	words []domaintranscript.Word,
) error {
	return h.issueRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		count, err := h.issueRepo.CountByMediaFileID(txCtx, cmd.MediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if count > 0 {
			return h.markJobSuccessInTx(txCtx, cmd.JobID)
		}

		if err := h.issueRepo.ReplaceForMediaFile(txCtx, cmd.MediaFileID, issues); err != nil {
			return messaging.Retryable(err)
		}
		if err := h.transcriptRepo.ReplaceWords(txCtx, transcriptID, words); err != nil {
			return messaging.Retryable(err)
		}

		events := []event.DomainEvent{
			domaintranscriptissue.TranscriptAnalysisReady{
				ID:              uuid.New().String(),
				MediaFileID:     cmd.MediaFileID.String(),
				ProjectID:       cmd.ProjectID.String(),
				UserID:          cmd.UserID.String(),
				TranscriptID:    transcriptID.String(),
				FillerCount:     countType(issues, domaintranscriptissue.TypeFiller),
				RepetitionCount: countType(issues, domaintranscriptissue.TypeRepetition),
				FalseStartCount: countType(issues, domaintranscriptissue.TypeFalseStart),
				Timestamp:       time.Now().UTC(),
			},
		}
		if cmd.JobID != uuid.Nil {
			job, err := h.jobRepo.GetByID(txCtx, cmd.JobID)
			if err != nil {
				return messaging.Retryable(err)
			}
			if job != nil && job.Status != domainjob.StatusSuccess {
				job.MarkSuccess()
				if err := h.jobRepo.Update(txCtx, job); err != nil {
					return messaging.Retryable(err)
				}
				events = append(events, job.PullEvents()...)
			}
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *AnalyzeTranscriptHandler) markProcessing(ctx context.Context, jobID uuid.UUID) error {
	if jobID == uuid.Nil {
		return nil
	}
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil || job.Status == domainjob.StatusSuccess || job.Status == domainjob.StatusProcessing {
			return nil
		}
		job.MarkProcessing()
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}

func (h *AnalyzeTranscriptHandler) markJobSuccess(ctx context.Context, jobID uuid.UUID) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		return h.markJobSuccessInTx(txCtx, jobID)
	})
}

func (h *AnalyzeTranscriptHandler) markJobSuccessInTx(ctx context.Context, jobID uuid.UUID) error {
	if jobID == uuid.Nil {
		return nil
	}
	job, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if job == nil || job.Status == domainjob.StatusSuccess {
		return nil
	}
	job.MarkSuccess()
	if err := h.jobRepo.Update(ctx, job); err != nil {
		return messaging.Retryable(err)
	}
	return h.outbox.StoreEvents(ctx, job.PullEvents())
}

func (h *AnalyzeTranscriptHandler) fail(ctx context.Context, jobID uuid.UUID, reason string) error {
	if jobID == uuid.Nil {
		return nil
	}
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil || job.Status == domainjob.StatusFailed {
			return nil
		}
		job.MarkFailed(reason)
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}
