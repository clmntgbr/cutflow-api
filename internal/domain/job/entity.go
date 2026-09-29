package job

import (
	"context"
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusSuccess    = "success"
	StatusFailed     = "failed"

	NameExtractAudio      = "extract_audio"
	NameDetectSilence     = "detect_silence"
	NameTranscribeAudio   = "transcribe_audio"
	NameAnalyzeTranscript = "analyze_transcript"
	NameAnalyzeViral      = "analyze_viral"
	NameRebuildTimeline   = "rebuild_timeline"
)

type Job struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	MediaFileID   uuid.UUID
	UserID        uuid.UUID
	Name          string
	Status        string
	ErrorMessage  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time

	events []event.DomainEvent
}

func New(projectID, mediaFileID, userID uuid.UUID, name string) *Job {
	now := time.Now().UTC()
	j := &Job{
		ID:          uuid.New(),
		ProjectID:   projectID,
		MediaFileID: mediaFileID,
		UserID:      userID,
		Name:        name,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	j.recordEvent(JobCreated{
		ID:          uuid.New().String(),
		JobID:       j.ID.String(),
		ProjectID:   j.ProjectID.String(),
		MediaFileID: j.MediaFileID.String(),
		UserID:      j.UserID.String(),
		Name:        j.Name,
		Status:      j.Status,
		Timestamp:   now,
	})
	return j
}

func (j *Job) PullEvents() []event.DomainEvent {
	events := j.events
	j.events = nil
	return events
}

func (j *Job) recordEvent(e event.DomainEvent) {
	j.events = append(j.events, e)
}

func (j *Job) recordStatusChanged() {
	j.recordEvent(JobStatusChanged{
		ID:           uuid.New().String(),
		JobID:        j.ID.String(),
		ProjectID:    j.ProjectID.String(),
		MediaFileID:  j.MediaFileID.String(),
		UserID:       j.UserID.String(),
		Name:         j.Name,
		Status:       j.Status,
		ErrorMessage: j.ErrorMessage,
		Timestamp:    j.UpdatedAt,
	})
}

func (j *Job) MarkProcessing() {
	if j.Status == StatusProcessing || j.Status == StatusSuccess {
		return
	}
	now := time.Now().UTC()
	j.Status = StatusProcessing
	j.ErrorMessage = ""
	j.StartedAt = &now
	j.UpdatedAt = now
	j.recordStatusChanged()
}

func (j *Job) MarkSuccess() {
	if j.Status == StatusSuccess {
		return
	}
	now := time.Now().UTC()
	j.Status = StatusSuccess
	j.ErrorMessage = ""
	j.CompletedAt = &now
	j.UpdatedAt = now
	j.recordStatusChanged()
}

func (j *Job) MarkFailed(reason string) {
	if j.Status == StatusFailed {
		return
	}
	now := time.Now().UTC()
	j.Status = StatusFailed
	j.ErrorMessage = reason
	j.CompletedAt = &now
	j.UpdatedAt = now
	j.recordStatusChanged()
}

type JobWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, job *Job) error
	Update(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
}
