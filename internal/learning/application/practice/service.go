package practice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

var (
	ErrVerbNotFound     = errors.New("irregular verb not found")
	ErrProgressNotFound = errors.New("skill progress not found")
	ErrConcurrentUpdate = errors.New("skill progress was updated concurrently")
	ErrInvalidCommand   = errors.New("invalid answer command")
)

// VerbCatalog is an outbound port owned by the use case that consumes it.
type VerbCatalog interface {
	NextForStudent(ctx context.Context, studentID string) (curriculum.IrregularVerb, error)
	ByID(ctx context.Context, id string) (curriculum.IrregularVerb, error)
}

type ProgressRepository interface {
	ByStudentAndSkill(ctx context.Context, studentID string, skillID shared.SkillID) (ProgressSnapshot, error)
}

// AttemptRecorder persists the attempt and updated progress atomically.
type AttemptRecorder interface {
	Record(ctx context.Context, attempt learning.PracticeAttempt, progress learning.SkillProgress, expectedVersion int64) error
}

type ProgressSnapshot struct {
	Progress learning.SkillProgress
	Version  int64
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() string
}

type Service struct {
	verbs    VerbCatalog
	progress ProgressRepository
	recorder AttemptRecorder
	clock    Clock
	ids      IDGenerator
}

func NewService(verbs VerbCatalog, progress ProgressRepository, recorder AttemptRecorder, clock Clock, ids IDGenerator) *Service {
	return &Service{verbs: verbs, progress: progress, recorder: recorder, clock: clock, ids: ids}
}

type Exercise struct {
	VerbID      string
	BaseForm    string
	Translation string
	SkillID     string
}

func (s *Service) NextExercise(ctx context.Context, studentID string) (Exercise, error) {
	if studentID == "" {
		return Exercise{}, ErrInvalidCommand
	}
	verb, err := s.verbs.NextForStudent(ctx, studentID)
	if err != nil {
		return Exercise{}, fmt.Errorf("get next irregular verb: %w", err)
	}
	return Exercise{
		VerbID: verb.ID(), BaseForm: verb.BaseForm(),
		Translation: verb.Translation(), SkillID: verb.SkillID().String(),
	}, nil
}

type SubmitAnswerCommand struct {
	StudentID      string
	VerbID         string
	PastSimple     string
	PastParticiple string
}

type AnswerResult struct {
	Correct                bool
	PastSimpleCorrect      bool
	PastParticipleCorrect  bool
	ExpectedPastSimple     string
	ExpectedPastParticiple string
	Feedback               string
	ValidationLevel        learning.ValidationLevel
	CorrectAnswers         int
	TotalAnswers           int
	NextReviewAt           time.Time
}

func (s *Service) SubmitAnswer(ctx context.Context, command SubmitAnswerCommand) (AnswerResult, error) {
	if command.StudentID == "" || command.VerbID == "" || !validAnswer(command.PastSimple) || !validAnswer(command.PastParticiple) {
		return AnswerResult{}, ErrInvalidCommand
	}

	verb, err := s.verbs.ByID(ctx, command.VerbID)
	if err != nil {
		return AnswerResult{}, fmt.Errorf("find irregular verb: %w", err)
	}
	correction := verb.Check(command.PastSimple, command.PastParticiple)
	for range 3 {
		snapshot, err := s.progress.ByStudentAndSkill(ctx, command.StudentID, verb.SkillID())
		if errors.Is(err, ErrProgressNotFound) {
			progress, createErr := learning.NewSkillProgress(command.StudentID, verb.SkillID())
			if createErr != nil {
				return AnswerResult{}, fmt.Errorf("create skill progress: %w", createErr)
			}
			snapshot = ProgressSnapshot{Progress: progress}
			err = nil
		}
		if err != nil {
			return AnswerResult{}, fmt.Errorf("load skill progress: %w", err)
		}

		now := s.clock.Now()
		snapshot.Progress.RecordRecall(correction.Correct(), now)
		attempt, err := learning.NewPracticeAttempt(
			s.ids.NewID(), command.StudentID, verb.ID(), verb.SkillID(),
			command.PastSimple, command.PastParticiple, correction.Correct(), now,
		)
		if err != nil {
			return AnswerResult{}, fmt.Errorf("create practice attempt: %w", err)
		}
		if err := s.recorder.Record(ctx, attempt, snapshot.Progress, snapshot.Version); errors.Is(err, ErrConcurrentUpdate) {
			continue
		} else if err != nil {
			return AnswerResult{}, fmt.Errorf("record answer: %w", err)
		}
		return answerResult(correction, snapshot.Progress), nil
	}
	return AnswerResult{}, fmt.Errorf("record answer after retries: %w", ErrConcurrentUpdate)
}

func answerResult(correction curriculum.VerbCorrection, progress learning.SkillProgress) AnswerResult {
	return AnswerResult{
		Correct:                correction.Correct(),
		PastSimpleCorrect:      correction.PastSimpleCorrect,
		PastParticipleCorrect:  correction.PastParticipleCorrect,
		ExpectedPastSimple:     correction.ExpectedPastSimple,
		ExpectedPastParticiple: correction.ExpectedParticiple,
		Feedback:               feedback(correction.Correct()),
		ValidationLevel:        progress.Level(),
		CorrectAnswers:         progress.CorrectAnswers(),
		TotalAnswers:           progress.TotalAnswers(),
		NextReviewAt:           progress.NextReviewAt(),
	}
}

func validAnswer(value string) bool {
	length := utf8.RuneCountInString(strings.TrimSpace(value))
	return length > 0 && length <= 100
}

func feedback(correct bool) string {
	if correct {
		return "Excellent! Both forms are correct. Keep going!"
	}
	return "Good try! Look at the corrected forms, then try them again soon."
}
