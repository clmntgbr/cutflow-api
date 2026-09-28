package project

import "time"

const EventTypeProjectCreated = "project.created.v1"

type ProjectCreated struct {
	ID        string    `json:"eventId"`
	ProjectID string    `json:"projectId"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ProjectCreated) EventID() string       { return e.ID }
func (e ProjectCreated) EventType() string     { return EventTypeProjectCreated }
func (e ProjectCreated) AggregateID() string   { return e.ProjectID }
func (e ProjectCreated) OccurredAt() time.Time { return e.Timestamp }
