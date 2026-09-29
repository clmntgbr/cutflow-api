package write

import (
	"time"

	domainviral "go-api/internal/domain/viral"

	"github.com/google/uuid"
)

type ViralCandidateModel struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID     uuid.UUID `gorm:"column:media_file_id"`
	ProjectID       uuid.UUID `gorm:"column:project_id"`
	SourceStartMs   int64     `gorm:"column:source_start_ms"`
	SourceEndMs     int64     `gorm:"column:source_end_ms"`
	Score           float64   `gorm:"column:score"`
	HookScore       *float64  `gorm:"column:hook_score"`
	StandaloneScore *float64  `gorm:"column:standalone_score"`
	PayoffScore     *float64  `gorm:"column:payoff_score"`
	InterestScore   *float64  `gorm:"column:interest_score"`
	Title           string    `gorm:"column:title"`
	Hook            string    `gorm:"column:hook"`
	Reason          string    `gorm:"column:reason"`
	Provider        string    `gorm:"column:provider"`
	Model           string    `gorm:"column:model"`
	Selected        bool      `gorm:"column:selected"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (ViralCandidateModel) TableName() string { return "viral_candidate" }

func viralCandidateModelFromDomain(c *domainviral.Candidate) *ViralCandidateModel {
	return &ViralCandidateModel{
		ID:              c.ID,
		MediaFileID:     c.MediaFileID,
		ProjectID:       c.ProjectID,
		SourceStartMs:   c.SourceStartMs,
		SourceEndMs:     c.SourceEndMs,
		Score:           c.Score,
		HookScore:       c.HookScore,
		StandaloneScore: c.StandaloneScore,
		PayoffScore:     c.PayoffScore,
		InterestScore:   c.InterestScore,
		Title:           c.Title,
		Hook:            c.Hook,
		Reason:          c.Reason,
		Provider:        c.Provider,
		Model:           c.Model,
		Selected:        c.Selected,
		CreatedAt:       c.CreatedAt,
	}
}
