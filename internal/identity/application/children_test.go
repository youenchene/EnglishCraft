package application

import (
	"context"
	"errors"
	"testing"
)

type childrenMemory struct {
	sessionMemory
	children []ChildProfile
	parent   string
	created  Account
}

func (r *childrenMemory) Children(context.Context, string) ([]ChildProfile, error) {
	return r.children, nil
}
func (r *childrenMemory) AddChild(_ context.Context, parent string, account Account) error {
	r.parent, r.created = parent, account
	return nil
}
func TestChildrenCreationAndOwnership(t *testing.T) {
	ctx := context.Background()
	repo := &childrenMemory{children: []ChildProfile{{ID: "owned", Username: "alex"}}}
	s := NewSessions(repo, passwordFake{}, &clockFake{})
	parent := Principal{SubjectID: "parent", Role: RoleParent}
	if _, err := s.AddChild(ctx, Principal{Role: RoleChild}, "alex", "long-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("child created an account")
	}
	if _, err := s.AddChild(ctx, parent, "alex", "short"); !errors.Is(err, ErrInvalidChild) {
		t.Fatal("weak password accepted")
	}
	child, err := s.AddChild(ctx, parent, "alex", "long-password")
	if err != nil || child.ID == "" || repo.parent != "parent" || repo.created.Role != RoleChild {
		t.Fatal("creation failed", err)
	}
	if _, err := s.SelectChild(ctx, parent, "another-family"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("foreign child accepted")
	}
	if id, err := s.SelectChild(ctx, parent, "owned"); err != nil || id != "owned" {
		t.Fatal("owned child rejected")
	}
}
