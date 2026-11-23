package domain

import (
	"errors"
	"strings"
)

var ErrInvalidSkillID = errors.New("invalid skill ID")

// SkillID is the small shared-kernel value used to correlate curriculum,
// learning, missions, and tutoring without sharing their domain entities.
type SkillID string

func NewSkillID(value string) (SkillID, error) {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "english.") {
		return "", ErrInvalidSkillID
	}
	return SkillID(value), nil
}

func (id SkillID) String() string { return string(id) }
