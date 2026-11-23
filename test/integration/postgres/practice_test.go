//go:build integration

package postgres_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	learningpostgres "github.com/youenchene/EnglishCraft/internal/learning/adapters/postgres"
	practice "github.com/youenchene/EnglishCraft/internal/learning/application/practice"
	platformpostgres "github.com/youenchene/EnglishCraft/internal/platform/postgres"
	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

type ids struct{}

func (ids) NewID() string { return uuid.NewString() }

func TestPracticeRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := platformpostgres.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	store := learningpostgres.NewStore(pool)
	service := practice.NewService(store, store, store, fixedClock{now: time.Now().UTC()}, ids{})
	studentID := "integration-" + uuid.NewString()
	exercise, err := service.NextExercise(ctx, studentID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.SubmitAnswer(ctx, practice.SubmitAnswerCommand{
		StudentID: studentID, VerbID: exercise.VerbID,
		PastSimple: "wrong", PastParticiple: "wrong",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Correct || result.TotalAnswers != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestConcurrentAnswersDoNotLoseProgress(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := platformpostgres.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	store := learningpostgres.NewStore(pool)
	service := practice.NewService(store, store, store, fixedClock{now: time.Now().UTC()}, ids{})
	studentID := "concurrent-" + uuid.NewString()
	exercise, err := service.NextExercise(ctx, studentID)
	if err != nil {
		t.Fatal(err)
	}

	errorsChannel := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, submitErr := service.SubmitAnswer(ctx, practice.SubmitAnswerCommand{
				StudentID: studentID, VerbID: exercise.VerbID,
				PastSimple: "wrong", PastParticiple: "wrong",
			})
			errorsChannel <- submitErr
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for submitErr := range errorsChannel {
		if submitErr != nil {
			t.Fatal(submitErr)
		}
	}

	skillID, _ := shared.NewSkillID(exercise.SkillID)
	snapshot, err := store.ByStudentAndSkill(ctx, studentID, skillID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Progress.TotalAnswers() != 2 {
		t.Fatalf("got %d answers, want 2", snapshot.Progress.TotalAnswers())
	}
}
