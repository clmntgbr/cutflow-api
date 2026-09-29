package project

import (
	"testing"

	"github.com/google/uuid"
)

func TestProject_MarkReady_FromProcessing(t *testing.T) {
	p, err := NewProject(testUserID(), "Demo")
	if err != nil {
		t.Fatalf("new project: %v", err)
	}
	_ = p.PullEvents()
	if err := p.MarkProcessing(); err != nil {
		t.Fatalf("mark processing: %v", err)
	}
	_ = p.PullEvents()

	if err := p.MarkReady(); err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	if p.Status != StatusReady {
		t.Fatalf("status: got %s want %s", p.Status, StatusReady)
	}
	events := p.PullEvents()
	if len(events) != 1 {
		t.Fatalf("events: got %d", len(events))
	}
	updated, ok := events[0].(ProjectUpdated)
	if !ok {
		t.Fatalf("event type: got %T", events[0])
	}
	if updated.Status != StatusReady {
		t.Fatalf("event status: got %s", updated.Status)
	}
}

func TestProject_MarkReady_Idempotent(t *testing.T) {
	p, err := NewProject(testUserID(), "Demo")
	if err != nil {
		t.Fatalf("new project: %v", err)
	}
	_ = p.PullEvents()
	_ = p.MarkProcessing()
	_ = p.PullEvents()
	_ = p.MarkReady()
	_ = p.PullEvents()

	if err := p.MarkReady(); err != nil {
		t.Fatalf("second mark ready: %v", err)
	}
	if len(p.PullEvents()) != 0 {
		t.Fatal("expected no events on idempotent MarkReady")
	}
}

func testUserID() uuid.UUID {
	return uuid.MustParse("01960000-0000-7000-8000-000000000001")
}
