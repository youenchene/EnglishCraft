package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Account struct {
	ID, Username, PasswordHash string
	Role                       Role
	Failed                     int
	LockedUntil                time.Time
}
type Session struct {
	Hash, AccountID string
	Expires         time.Time
}
type Family struct {
	Parent Account
	Child  Account
}
type SessionRepository interface {
	Login(context.Context, string, func(*Account) (Session, error)) error
	Resolve(context.Context, string, time.Time) (Principal, error)
	Revoke(context.Context, string) error
	ChildForParent(context.Context, string) (string, error)
	Provision(context.Context, Family) error
}
type Passwords interface {
	Hash(string) (string, error)
	Verify(string, string) bool
}
type SessionClock interface{ Now() time.Time }
type Sessions struct {
	repo      SessionRepository
	passwords Passwords
	clock     SessionClock
}

func NewSessions(repo SessionRepository, passwords Passwords, clock SessionClock) *Sessions {
	return &Sessions{repo: repo, passwords: passwords, clock: clock}
}
func TokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func (s *Sessions) Login(ctx context.Context, username, password string) (string, Principal, error) {
	if len(username) > 80 || len(password) > 72 || password == "" {
		return "", Principal{}, ErrInvalidCredentials
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", Principal{}, err
	}
	token := hex.EncodeToString(bytes)
	var principal Principal
	err := s.repo.Login(ctx, strings.TrimSpace(username), func(account *Account) (Session, error) {
		now := s.clock.Now()
		if account.LockedUntil.After(now) {
			return Session{}, ErrInvalidCredentials
		}
		verified := s.passwords.Verify(account.PasswordHash, password)
		if !verified || account.ID == "" {
			account.Failed++
			if account.Failed >= 5 {
				account.LockedUntil = now.Add(15 * time.Minute)
				account.Failed = 0
			}
			return Session{}, ErrInvalidCredentials
		}
		account.Failed, account.LockedUntil = 0, time.Time{}
		principal = Principal{SubjectID: account.ID, Role: account.Role}
		return Session{Hash: TokenHash(token), AccountID: account.ID, Expires: now.Add(12 * time.Hour)}, nil
	})
	return token, principal, err
}
func (s *Sessions) Resolve(ctx context.Context, token string) (Principal, error) {
	if len(token) != 64 {
		return Principal{}, ErrInvalidCredentials
	}
	return s.repo.Resolve(ctx, TokenHash(token), s.clock.Now())
}
func (s *Sessions) Logout(ctx context.Context, token string) error {
	return s.repo.Revoke(ctx, TokenHash(token))
}
func (s *Sessions) Child(ctx context.Context, parent string) (string, error) {
	return s.repo.ChildForParent(ctx, parent)
}
func (s *Sessions) Provision(ctx context.Context, parentName, parentPassword, childName, childPassword string) error {
	usernamePattern := regexp.MustCompile(`^[A-Za-z0-9_.-]{3,80}$`)
	validNames := usernamePattern.MatchString(parentName) && usernamePattern.MatchString(childName) && parentName != childName
	validPasswords := len(parentPassword) >= 12 && len(childPassword) >= 12 && parentPassword != childPassword &&
		len(parentPassword) <= 72 && len(childPassword) <= 72
	if !validNames || !validPasswords {
		return errors.New("distinct usernames and passwords of at least 12 characters required")
	}
	parentHash, err := s.passwords.Hash(parentPassword)
	if err != nil {
		return err
	}
	childHash, err := s.passwords.Hash(childPassword)
	if err != nil {
		return err
	}
	return s.repo.Provision(ctx, Family{
		Parent: Account{ID: "family-parent", Username: parentName, PasswordHash: parentHash, Role: RoleParent},
		Child:  Account{ID: "family-child", Username: childName, PasswordHash: childHash, Role: RoleChild},
	})
}
