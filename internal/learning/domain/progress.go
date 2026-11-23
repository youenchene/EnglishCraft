package domain

import (
	"errors"
	"time"

	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

var ErrInvalidProgress = errors.New("invalid skill progress")

type ValidationLevel string

const (
	ValidationNotStarted  ValidationLevel = "not_started"
	ValidationRecognition ValidationLevel = "recognition"
	ValidationRecall      ValidationLevel = "recall"
	ValidationContext     ValidationLevel = "context"
)

type SkillProgress struct {
	studentID      string
	skillID        shared.SkillID
	correctAnswers int
	totalAnswers   int
	level          ValidationLevel
	nextReviewAt   time.Time
}

func NewSkillProgress(studentID string, skillID shared.SkillID) (SkillProgress, error) {
	if studentID == "" || skillID == "" {
		return SkillProgress{}, ErrInvalidProgress
	}
	return SkillProgress{studentID: studentID, skillID: skillID, level: ValidationNotStarted}, nil
}

func RestoreSkillProgress(studentID string, skillID shared.SkillID, correct, total int, level ValidationLevel, nextReviewAt time.Time) (SkillProgress, error) {
	if studentID == "" || skillID == "" || correct < 0 || total < 0 || correct > total || !level.valid() {
		return SkillProgress{}, ErrInvalidProgress
	}
	return SkillProgress{studentID: studentID, skillID: skillID, correctAnswers: correct, totalAnswers: total, level: level, nextReviewAt: nextReviewAt}, nil
}

func (p *SkillProgress) RecordRecall(correct bool, now time.Time) {
	p.totalAnswers++
	if correct {
		p.correctAnswers++
	}

	// Recall requires repeated evidence; context validation is handled by a
	// separate exercise type and can never be granted by this drill.
	switch {
	case p.totalAnswers >= 5 && p.correctAnswers*100/p.totalAnswers >= 80:
		p.level = ValidationRecall
		p.nextReviewAt = now.Add(7 * 24 * time.Hour)
	case p.correctAnswers >= 1:
		p.level = ValidationRecognition
		p.nextReviewAt = now.Add(24 * time.Hour)
	default:
		p.level = ValidationNotStarted
		p.nextReviewAt = now.Add(4 * time.Hour)
	}
}

func (p SkillProgress) StudentID() string       { return p.studentID }
func (p SkillProgress) SkillID() shared.SkillID { return p.skillID }
func (p SkillProgress) CorrectAnswers() int     { return p.correctAnswers }
func (p SkillProgress) TotalAnswers() int       { return p.totalAnswers }
func (p SkillProgress) Level() ValidationLevel  { return p.level }
func (p SkillProgress) NextReviewAt() time.Time { return p.nextReviewAt }

func (l ValidationLevel) valid() bool {
	return l == ValidationNotStarted || l == ValidationRecognition || l == ValidationRecall || l == ValidationContext
}
