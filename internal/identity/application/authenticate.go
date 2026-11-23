package application

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Role string

const (
	RoleParent Role = "parent"
	RoleChild  Role = "child"
)

type Principal struct {
	SubjectID string
	Role      Role
}

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (Principal, error)
}

type Authenticate struct{ verifier TokenVerifier }

func NewAuthenticate(verifier TokenVerifier) *Authenticate {
	return &Authenticate{verifier: verifier}
}

func (a *Authenticate) Execute(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrInvalidCredentials
	}
	return a.verifier.Verify(ctx, token)
}
