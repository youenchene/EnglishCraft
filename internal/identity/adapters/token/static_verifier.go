package token

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"

	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

type credential struct {
	digest    [sha256.Size]byte
	principal identity.Principal
}

// StaticVerifier supports pre-provisioned MVP accounts without public signup.
// Only SHA-256 token digests are retained in memory.
type StaticVerifier struct{ credentials []credential }

func NewStaticVerifier(parentToken, childToken string) *StaticVerifier {
	return &StaticVerifier{credentials: []credential{
		{digest: sha256.Sum256([]byte(parentToken)), principal: identity.Principal{SubjectID: "parent-1", Role: identity.RoleParent}},
		{digest: sha256.Sum256([]byte(childToken)), principal: identity.Principal{SubjectID: "student-1", Role: identity.RoleChild}},
	}}
}

func (v *StaticVerifier) Verify(_ context.Context, rawToken string) (identity.Principal, error) {
	digest := sha256.Sum256([]byte(rawToken))
	for _, candidate := range v.credentials {
		if subtle.ConstantTimeCompare(digest[:], candidate.digest[:]) == 1 {
			return candidate.principal, nil
		}
	}
	return identity.Principal{}, identity.ErrInvalidCredentials
}
