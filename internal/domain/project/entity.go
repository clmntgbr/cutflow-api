package project

import (
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewProject(userID uuid.UUID, name string) (*Project, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = "Untitled"
	}
	if utf8.RuneCountInString(name) > 255 {
		name = string([]rune(name)[:255])
	}

	now := time.Now().UTC()
	p := &Project{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Status:    StatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	p.recordEvent(ProjectCreated{
		ID:        uuid.New().String(),
		ProjectID: p.ID.String(),
		UserID:    p.UserID.String(),
		Name:      p.Name,
		Status:    p.Status,
		Timestamp: now,
	})
	return p, nil
}

func NameFromFilename(filename string) string {
	base := filepath.Base(strings.TrimSpace(filename))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "Untitled"
	}
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	name = strings.TrimSpace(name)
	if name == "" {
		return "Untitled"
	}
	return name
}

func (p *Project) PullEvents() []event.DomainEvent {
	events := p.events
	p.events = nil
	return events
}

func (p *Project) recordEvent(e event.DomainEvent) {
	p.events = append(p.events, e)
}
