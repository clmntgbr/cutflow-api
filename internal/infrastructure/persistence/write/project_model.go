package write

import (
	"time"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ProjectModel struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id"`
	Name      string    `gorm:"column:name"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ProjectModel) TableName() string {
	return "project"
}

func projectModelFromDomain(p *domainproject.Project) *ProjectModel {
	return &ProjectModel{
		ID:        p.ID,
		UserID:    p.UserID,
		Name:      p.Name,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func projectDomainFromModel(m *ProjectModel) *domainproject.Project {
	return &domainproject.Project{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
