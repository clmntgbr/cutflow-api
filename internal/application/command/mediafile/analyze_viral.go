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
	domainviral "go-api/internal/domain/viral"

	"github.com/google/uuid"
)

type AnalyzeViralCommand struct {
	MediaFileID  uuid.UUID
	TranscriptID uuid.UUID
	ProjectID    uuid.UUID
	UserID       uuid.UUID
	JobID        uuid.UUID
}

type AnalyzeViralHandler struct {
	transcriptRepo domaintranscript.TranscriptWriteRepository
	mediaRepo      domainmediafile.MediaFileWriteRepository
	configRepo     domainmediaconfig.MediaConfigurationWriteRepository
	candidateRepo  domainviral.CandidateWriteRepository
	jobRepo        domainjob.JobWriteRepository
	analyzer       port.ViralAnalyzer
	outbox         port.OutboxRepository
	chunkDuration  int64
	chunkOverlap   int64
}

func NewAnalyzeViralHandler(
	transcriptRepo domaintranscript.TranscriptWriteRepository,
	mediaRepo domainmediafile.MediaFileWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	candidateRepo domainviral.CandidateWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	analyzer port.ViralAnalyzer,
	outbox port.OutboxRepository,
	chunkDurationMs, chunkOverlapMs int64,
) *AnalyzeViralHandler {
	return &AnalyzeViralHandler{
		transcriptRepo: transcriptRepo,
		mediaRepo:      mediaRepo,
		configRepo:     configRepo,
		candidateRepo:  candidateRepo,
		jobRepo:        jobRepo,
		analyzer:       analyzer,
		outbox:         outbox,
		chunkDuration:  chunkDurationMs,
		chunkOverlap:   chunkOverlapMs,
	}
}

func (h *AnalyzeViralHandler) Handle(ctx context.Context, cmd AnalyzeViralCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	existing, err := h.candidateRepo.CountByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing > 0 {
		return h.markJobSuccess(ctx, cmd.JobID)
	}

	cfg, err := h.configRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if cfg == nil {
		cfg = domainmediaconfig.NewDefault(cmd.MediaFileID)
	}
	if !cfg.ViralDetectionEnabled {
		log.Printf("viral analysis skipped mediaFileId=%s: disabled", cmd.MediaFileID)
		return h.persist(ctx, cmd, nil)
	}

	transcript, err := h.transcriptRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if transcript == nil || transcript.Status != domaintranscript.StatusCompleted {
		if failErr := h.fail(ctx, cmd.JobID, "transcript not ready"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}

	words, err := h.transcriptRepo.ListWords(ctx, transcript.ID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if len(words) == 0 {
		return h.persist(ctx, cmd, nil)
	}

	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	mediaDuration := int64(0)
	if media != nil {
		mediaDuration = media.DurationMs
	}

	maxCandidates := cfg.ViralMaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = domainmediaconfig.DefaultViralMaxCandidates
	}
	minScore := cfg.ViralMinScore
	if minScore <= 0 {
		minScore = domainmediaconfig.DefaultViralMinScore
	}

	chunks := domainviral.ChunkWords(words, h.chunkDuration, h.chunkOverlap)
	raw := make([]*domainviral.Candidate, 0)
	for _, chunk := range chunks {
		text := domainviral.FormatTimestamped(chunk.Words, 5000)
		proposals, analyzeErr := h.analyzer.Analyze(ctx, port.ViralAnalyzeInput{
			TranscriptText: text,
			Language:       transcript.Language,
			MinDurationMs:  int64(cfg.ViralClipMinDurationMs),
			MaxDurationMs:  int64(cfg.ViralClipMaxDurationMs),
			MaxCandidates:  maxCandidates,
		})
		if analyzeErr != nil {
			log.Printf("viral llm failed mediaFileId=%s: %v", cmd.MediaFileID, analyzeErr)
			if failErr := h.fail(ctx, cmd.JobID, "viral analysis failed"); failErr != nil {
				return failErr
			}
			return messaging.NonRetryable(analyzeErr)
		}
		for _, p := range proposals {
			start, end := domainviral.ResolveBoundaries(words, p.StartMs, p.EndMs)
			raw = append(raw, &domainviral.Candidate{
				ID:              uuid.New(),
				MediaFileID:     cmd.MediaFileID,
				ProjectID:       cmd.ProjectID,
				SourceStartMs:   start,
				SourceEndMs:     end,
				Score:           p.Score,
				HookScore:       p.HookScore,
				StandaloneScore: p.StandaloneScore,
				PayoffScore:     p.PayoffScore,
				InterestScore:   p.InterestScore,
				Title:           p.Title,
				Hook:            p.Hook,
				Reason:          p.Reason,
				Provider:        h.analyzer.Provider(),
				Model:           h.analyzer.Model(),
				CreatedAt:       time.Now().UTC(),
			})
		}
	}

	validated := domainviral.FilterValid(raw, domainviral.ValidateConfig{
		MediaDurationMs: mediaDuration,
		MinDurationMs:   int64(cfg.ViralClipMinDurationMs),
		MaxDurationMs:   int64(cfg.ViralClipMaxDurationMs),
		MinScore:        minScore,
	})
	ranked := domainviral.Rank(validated, maxCandidates)

	log.Printf(
		"viral analysis mediaFileId=%s chunks=%d raw=%d kept=%d provider=%s model=%s",
		cmd.MediaFileID, len(chunks), len(raw), len(ranked), h.analyzer.Provider(), h.analyzer.Model(),
	)
	return h.persist(ctx, cmd, ranked)
}

func (h *AnalyzeViralHandler) persist(
	ctx context.Context,
	cmd AnalyzeViralCommand,
	candidates []*domainviral.Candidate,
) error {
	return h.candidateRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		count, err := h.candidateRepo.CountByMediaFileID(txCtx, cmd.MediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if count > 0 {
			return h.markJobSuccessInTx(txCtx, cmd.JobID)
		}
		if err := h.candidateRepo.ReplaceForMediaFile(txCtx, cmd.MediaFileID, candidates); err != nil {
			return messaging.Retryable(err)
		}
		events := []event.DomainEvent{
			domainviral.ViralReady{
				ID:             uuid.New().String(),
				MediaFileID:    cmd.MediaFileID.String(),
				ProjectID:      cmd.ProjectID.String(),
				UserID:         cmd.UserID.String(),
				CandidateCount: len(candidates),
				Timestamp:      time.Now().UTC(),
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

func (h *AnalyzeViralHandler) markProcessing(ctx context.Context, jobID uuid.UUID) error {
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

func (h *AnalyzeViralHandler) markJobSuccess(ctx context.Context, jobID uuid.UUID) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		return h.markJobSuccessInTx(txCtx, jobID)
	})
}

func (h *AnalyzeViralHandler) markJobSuccessInTx(ctx context.Context, jobID uuid.UUID) error {
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

func (h *AnalyzeViralHandler) fail(ctx context.Context, jobID uuid.UUID, reason string) error {
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
