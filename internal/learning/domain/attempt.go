package domain

import (
	"errors"
	"time"

	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

var ErrInvalidAttempt = errors.New("invalid practice attempt")

type PracticeAttempt struct {
	id             string
	studentID      string
	verbID         string
	skillID        shared.SkillID
	pastSimple     string
	pastParticiple string
	correct        bool
	answeredAt     time.Time
}

func NewPracticeAttempt(
	id, studentID, verbID string,
	skillID shared.SkillID,
	pastSimple, pastParticiple string,
	correct bool,
	answeredAt time.Time,
) (PracticeAttempt, error) {
	if id == "" || studentID == "" || verbID == "" || skillID == "" || answeredAt.IsZero() {
		return PracticeAttempt{}, ErrInvalidAttempt
	}
	return PracticeAttempt{
		id: id, studentID: studentID, verbID: verbID, skillID: skillID,
		pastSimple: pastSimple, pastParticiple: pastParticiple,
		correct: correct, answeredAt: answeredAt,
	}, nil
}

func (a PracticeAttempt) ID() string              { return a.id }
func (a PracticeAttempt) StudentID() string       { return a.studentID }
func (a PracticeAttempt) VerbID() string          { return a.verbID }
func (a PracticeAttempt) SkillID() shared.SkillID { return a.skillID }
func (a PracticeAttempt) PastSimple() string      { return a.pastSimple }
func (a PracticeAttempt) PastParticiple() string  { return a.pastParticiple }
func (a PracticeAttempt) Correct() bool           { return a.correct }
func (a PracticeAttempt) AnsweredAt() time.Time   { return a.answeredAt }
