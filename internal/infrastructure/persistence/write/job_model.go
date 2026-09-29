package write

import (
	"time"

	domainjob "go-api/internal/domain/job"

	"github.com/google/uuid"
)

type JobModel struct {
	ID           uuid.UUID  `gorm:"column:id;primaryKey"`
	ProjectID    uuid.UUID  `gorm:"column:project_id"`
	MediaFileID  uuid.UUID  `gorm:"column:media_file_id"`
	UserID       uuid.UUID  `gorm:"column:user_id"`
	Name         string     `gorm:"column:name"`
	Status       string     `gorm:"column:status"`
	ErrorMessage *string    `gorm:"column:error_message"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	StartedAt    *time.Time `gorm:"column:started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
}

func (JobModel) TableName() string { return "job" }

func jobModelFromDomain(j *domainjob.Job) *JobModel {
	model := &JobModel{
		ID:          j.ID,
		ProjectID:   j.ProjectID,
		MediaFileID: j.MediaFileID,
		UserID:      j.UserID,
		Name:        j.Name,
		Status:      j.Status,
		CreatedAt:   j.CreatedAt,
		UpdatedAt:   j.UpdatedAt,
		StartedAt:   j.StartedAt,
		CompletedAt: j.CompletedAt,
	}
	if j.ErrorMessage != "" {
		msg := j.ErrorMessage
		model.ErrorMessage = &msg
	}
	return model
}

func jobDomainFromModel(m *JobModel) *domainjob.Job {
	j := &domainjob.Job{
		ID:          m.ID,
		ProjectID:   m.ProjectID,
		MediaFileID: m.MediaFileID,
		UserID:      m.UserID,
		Name:        m.Name,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,
	}
	if m.ErrorMessage != nil {
		j.ErrorMessage = *m.ErrorMessage
	}
	return j
}
