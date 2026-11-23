package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
)

var ErrInvalidChild = errors.New("invalid child credentials")
var ErrUsernameTaken = errors.New("username already exists")

type ChildProfile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
type ChildrenRepository interface {
	Children(context.Context, string) ([]ChildProfile, error)
	AddChild(context.Context, string, Account) error
}

func (s *Sessions) Children(ctx context.Context, parent Principal) ([]ChildProfile, error) {
	if parent.Role != RoleParent {
		return nil, ErrInvalidCredentials
	}
	repo, ok := s.repo.(ChildrenRepository)
	if !ok {
		return nil, errors.New("children repository unavailable")
	}
	return repo.Children(ctx, parent.SubjectID)
}
func (s *Sessions) SelectChild(ctx context.Context, parent Principal, id string) (string, error) {
	if parent.Role != RoleParent {
		return "", ErrInvalidCredentials
	}
	if id == "" {
		return s.Child(ctx, parent.SubjectID)
	}
	children, err := s.Children(ctx, parent)
	if err != nil {
		return "", err
	}
	for _, child := range children {
		if child.ID == id {
			return id, nil
		}
	}
	return "", ErrInvalidCredentials
}
func (s *Sessions) AddChild(ctx context.Context, parent Principal, username, password string) (ChildProfile, error) {
	if parent.Role != RoleParent {
		return ChildProfile{}, ErrInvalidCredentials
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]{3,80}$`).MatchString(username) || len(password) < 12 || len(password) > 72 {
		return ChildProfile{}, ErrInvalidChild
	}
	repo, ok := s.repo.(ChildrenRepository)
	if !ok {
		return ChildProfile{}, errors.New("children repository unavailable")
	}
	hash, err := s.passwords.Hash(password)
	if err != nil {
		return ChildProfile{}, err
	}
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ChildProfile{}, err
	}
	child := Account{ID: "child-" + hex.EncodeToString(bytes), Username: username, PasswordHash: hash, Role: RoleChild}
	if err := repo.AddChild(ctx, parent.SubjectID, child); err != nil {
		return ChildProfile{}, err
	}
	return ChildProfile{ID: child.ID, Username: username}, nil
}
