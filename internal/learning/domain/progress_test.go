package domain

import (
	"testing"
	"time"

	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

func TestRecallRequiresRepeatedEvidence(t *testing.T) {
	skillID, _ := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	progress, err := NewSkillProgress("student-1", skillID)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)

	for range 4 {
		progress.RecordRecall(true, now)
	}
	if progress.Level() != ValidationRecognition {
		t.Fatalf("got %s before enough evidence", progress.Level())
	}
	progress.RecordRecall(true, now)
	if progress.Level() != ValidationRecall {
		t.Fatalf("got %s, want %s", progress.Level(), ValidationRecall)
	}
	if got, want := progress.NextReviewAt(), now.Add(7*24*time.Hour); !got.Equal(want) {
		t.Fatalf("next review = %s, want %s", got, want)
	}
}

func TestIncorrectFirstAnswerSchedulesNearReview(t *testing.T) {
	skillID, _ := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	progress, _ := NewSkillProgress("student-1", skillID)
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)

	progress.RecordRecall(false, now)

	if progress.Level() != ValidationNotStarted {
		t.Fatalf("got %s", progress.Level())
	}
	if !progress.NextReviewAt().Equal(now.Add(4 * time.Hour)) {
		t.Fatalf("unexpected review date: %s", progress.NextReviewAt())
	}
}
