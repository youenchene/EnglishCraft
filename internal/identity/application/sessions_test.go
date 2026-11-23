package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

type sessionMemory struct {
	account Account
	session Session
}

func (r *sessionMemory) Login(_ context.Context, _ string, fn func(*Account) (Session, error)) error {
	s, err := fn(&r.account)
	if err == nil {
		r.session = s
	}
	return err
}
func (r *sessionMemory) Resolve(_ context.Context, hash string, now time.Time) (Principal, error) {
	if r.session.Hash != hash || !r.session.Expires.After(now) {
		return Principal{}, ErrInvalidCredentials
	}
	return Principal{SubjectID: r.account.ID, Role: r.account.Role}, nil
}
func (r *sessionMemory) Revoke(context.Context, string) error                   { r.session = Session{}; return nil }
func (r *sessionMemory) ChildForParent(context.Context, string) (string, error) { return "child", nil }
func (r *sessionMemory) Provision(context.Context, Family) error                { return nil }

type passwordFake struct{}

func (passwordFake) Hash(value string) (string, error) { return value, nil }
func (passwordFake) Verify(hash, value string) bool    { return hash == value }

type clockFake struct{ now time.Time }

func (c *clockFake) Now() time.Time { return c.now }
func TestLoginLockoutExpiryAndRevocation(t *testing.T) {
	ctx := context.Background()
	clock := &clockFake{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	repo := &sessionMemory{account: Account{ID: "child", Role: RoleChild, PasswordHash: "long-password"}}
	s := NewSessions(repo, passwordFake{}, clock)
	for range 5 {
		if _, _, err := s.Login(ctx, "alex", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatal("bad password accepted")
		}
	}
	if _, _, err := s.Login(ctx, "alex", "long-password"); err == nil {
		t.Fatal("lockout bypassed")
	}
	clock.now = clock.now.Add(15 * time.Minute)
	token, p, err := s.Login(ctx, "alex", "long-password")
	if err != nil || p.Role != RoleChild || len(token) != 64 {
		t.Fatal("login failed")
	}
	if repo.session.Hash == token {
		t.Fatal("raw token persisted")
	}
	if _, err := s.Resolve(ctx, token); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(12 * time.Hour)
	if _, err := s.Resolve(ctx, token); err == nil {
		t.Fatal("expired session accepted")
	}
	clock.now = clock.now.Add(-12 * time.Hour)
	if err := s.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(ctx, token); err == nil {
		t.Fatal("revoked session accepted")
	}
}
func TestProvisionRejectsWeakOrSharedPasswords(t *testing.T) {
	s := NewSessions(&sessionMemory{}, passwordFake{}, &clockFake{})
	if err := s.Provision(context.Background(), "parent", "1234", "alex", "1234"); err == nil {
		t.Fatal("PIN accepted")
	}
}
