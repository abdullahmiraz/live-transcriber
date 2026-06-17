package meeting

import (
	"context"
	"testing"
)

type fakeRepo struct {
	created   *Meeting
	bySlug    map[string]*Meeting
	createErr error
}

func (f *fakeRepo) Create(_ context.Context, m *Meeting) error {
	if f.createErr != nil {
		return f.createErr
	}
	m.ID = "fake-id"
	f.created = m
	if f.bySlug == nil {
		f.bySlug = map[string]*Meeting{}
	}
	f.bySlug[m.Slug] = m
	return nil
}

func (f *fakeRepo) GetBySlug(_ context.Context, slug string) (*Meeting, error) {
	if m, ok := f.bySlug[slug]; ok {
		return m, nil
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) End(_ context.Context, slug string) (*Meeting, error) {
	m, ok := f.bySlug[slug]
	if !ok {
		return nil, ErrNotFound
	}
	m.Status = StatusEnded
	return m, nil
}

func (f *fakeRepo) DeleteBySlug(_ context.Context, slug string) error {
	if f.bySlug == nil {
		return ErrNotFound
	}
	if _, ok := f.bySlug[slug]; !ok {
		return ErrNotFound
	}
	delete(f.bySlug, slug)
	return nil
}

func TestCreateAssignsSlugAndActiveStatus(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	m, err := svc.Create(context.Background(), CreateInput{Title: " Standup ", HostName: " Alice "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Slug == "" {
		t.Error("expected a generated slug")
	}
	if m.Status != StatusActive {
		t.Errorf("expected status %q, got %q", StatusActive, m.Status)
	}
	if m.Title != "Standup" || m.HostName != "Alice" {
		t.Errorf("expected trimmed fields, got title=%q host=%q", m.Title, m.HostName)
	}
}

func TestGetBySlugNotFound(t *testing.T) {
	svc := NewService(&fakeRepo{})
	if _, err := svc.GetBySlug(context.Background(), "missing"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEndMarksEnded(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	created, _ := svc.Create(context.Background(), CreateInput{Title: "x"})

	ended, err := svc.End(context.Background(), created.Slug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ended.Status != StatusEnded {
		t.Errorf("expected ended, got %q", ended.Status)
	}
}
