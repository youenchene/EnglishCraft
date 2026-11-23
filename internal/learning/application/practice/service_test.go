package practice

import (
	"context"
	"errors"
	"testing"
	"time"

	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

type fakeCatalog struct{ verb curriculum.IrregularVerb }

func (f fakeCatalog) NextForStudent(context.Context, string) (curriculum.IrregularVerb, error) {
	return f.verb, nil
}
func (f fakeCatalog) ByID(context.Context, string) (curriculum.IrregularVerb, error) {
	return f.verb, nil
}

type fakeProgressRepository struct {
	snapshot ProgressSnapshot
	err      error
}

func (f fakeProgressRepository) ByStudentAndSkill(context.Context, string, shared.SkillID) (ProgressSnapshot, error) {
	return f.snapshot, f.err
}

type recordingRecorder struct {
	attempt  learning.PracticeAttempt
	progress learning.SkillProgress
}

func (r *recordingRecorder) Record(_ context.Context, attempt learning.PracticeAttempt, progress learning.SkillProgress, _ int64) error {
	r.attempt, r.progress = attempt, progress
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "attempt-1" }

type concurrentRecorder struct{ calls int }

func (r *concurrentRecorder) Record(context.Context, learning.PracticeAttempt, learning.SkillProgress, int64) error {
	r.calls++
	if r.calls == 1 {
		return ErrConcurrentUpdate
	}
	return nil
}

func TestSubmitAnswerCorrectsAndRecordsProgress(t *testing.T) {
	skillID, _ := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	verb, _ := curriculum.NewIrregularVerb("verb-go", "go", "went", "gone", "aller", skillID)
	recorder := &recordingRecorder{}
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	service := NewService(
		fakeCatalog{verb: verb},
		fakeProgressRepository{err: ErrProgressNotFound},
		recorder, fixedClock{now: now}, fixedIDs{},
	)

	result, err := service.SubmitAnswer(context.Background(), SubmitAnswerCommand{
		StudentID: "student-1", VerbID: "verb-go", PastSimple: "went", PastParticiple: "gone",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct || result.ValidationLevel != learning.ValidationRecognition {
		t.Fatalf("unexpected result: %#v", result)
	}
	if recorder.attempt.ID() != "attempt-1" || recorder.progress.TotalAnswers() != 1 {
		t.Fatalf("attempt was not recorded: %#v", recorder)
	}
}

func TestSubmitAnswerRejectsIncompleteCommand(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil)
	_, err := service.SubmitAnswer(context.Background(), SubmitAnswerCommand{})
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("got %v, want %v", err, ErrInvalidCommand)
	}
}

func TestSubmitAnswerRetriesConcurrentProgressUpdate(t *testing.T) {
	skillID, _ := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	verb, _ := curriculum.NewIrregularVerb("verb-go", "go", "went", "gone", "aller", skillID)
	recorder := &concurrentRecorder{}
	service := NewService(
		fakeCatalog{verb: verb}, fakeProgressRepository{err: ErrProgressNotFound}, recorder,
		fixedClock{now: time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)}, fixedIDs{},
	)

	_, err := service.SubmitAnswer(context.Background(), SubmitAnswerCommand{
		StudentID: "student-1", VerbID: "verb-go", PastSimple: "went", PastParticiple: "gone",
	})
	if err != nil {
		t.Fatal(err)
	}
	if recorder.calls != 2 {
		t.Fatalf("got %d record attempts, want 2", recorder.calls)
	}
}
