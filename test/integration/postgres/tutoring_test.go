//go:build integration

package postgres_test

import (
	"errors"
	identitypg "github.com/youenchene/EnglishCraft/internal/identity/adapters/postgres"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	tutorpg "github.com/youenchene/EnglishCraft/internal/tutoring/adapters/postgres"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTutoringBudgetConcurrentReservationsAndDeletion(t *testing.T) {
	pool := isolatedPool(t)
	identities := identitypg.New(pool)
	if err := identities.Provision(t.Context(), identity.Family{
		Parent: identity.Account{ID: "parent", Username: "parent", PasswordHash: "test", Role: identity.RoleParent},
		Child:  identity.Account{ID: "child", Username: "child", PasswordHash: "test", Role: identity.RoleChild},
	}); err != nil {
		t.Fatal(err)
	}
	store := tutorpg.New(pool)
	if err := store.SetEnabled(t.Context(), "child", true); err != nil {
		t.Fatal(err)
	}
	var accepted atomic.Int32
	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			_, err := store.Reserve(t.Context(), tutor.Allowance{Student: "child", Provider: "openai", At: time.Now(), ReserveMicros: 100, DailyMicros: 500, StudentDailyMicros: 400})
			if err == nil {
				accepted.Add(1)
				return
			}
			if !errors.Is(err, tutor.ErrBudget) {
				t.Error(err)
			}
		})
	}
	group.Wait()
	if accepted.Load() != 4 {
		t.Fatalf("accepted %d, want 4", accepted.Load())
	}
	values, err := store.Spending(t.Context(), "child", time.Now())
	if err != nil || len(values) != 1 || values[0].ReservedMicros != 400 {
		t.Fatalf("%+v %v", values, err)
	}
	if err := store.Record(t.Context(), "missing", tutor.Usage{}, "ok"); !errors.Is(err, tutor.ErrInvalid) {
		t.Fatalf("missing audit accepted: %v", err)
	}
	if err := identities.DeleteChild(t.Context(), "someone-else", "child"); !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Fatal(err)
	}
	if err := identities.DeleteChild(t.Context(), "parent", "child"); err != nil {
		t.Fatal(err)
	}
	values, err = store.Spending(t.Context(), "child", time.Now())
	if err != nil || len(values) != 0 {
		t.Fatal("personal usage not deleted")
	}
}
func TestVoiceOwnershipDeadlinesAndImportApproval(t *testing.T) {
	pool := isolatedPool(t)
	identities := identitypg.New(pool)
	if err := identities.Provision(t.Context(), identity.Family{
		Parent: identity.Account{ID: "parent", Username: "parent", PasswordHash: "test", Role: identity.RoleParent},
		Child:  identity.Account{ID: "child", Username: "child", PasswordHash: "test", Role: identity.RoleChild},
	}); err != nil {
		t.Fatal(err)
	}
	store := tutorpg.New(pool)
	deadline := time.Now().Add(-time.Second)
	if err := store.Save(t.Context(), tutor.VoiceRecord{Student: "child", Mission: "diamond", Session: tutor.SpeechSession{ID: "rtc_test", Expires: deadline}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Find(t.Context(), "other", "rtc_test"); !errors.Is(err, tutor.ErrInvalid) {
		t.Fatal("foreign voice accessible")
	}
	admission, err := store.Lookup(t.Context(), "rtc_test")
	if err != nil || admission.Student != "child" || admission.Mission != "diamond" {
		t.Fatal("callback admission lost server identity")
	}
	if _, err := store.Lookup(t.Context(), "unknown"); !errors.Is(err, tutor.ErrInvalid) {
		t.Fatal("unknown callback admission accepted")
	}
	expired, err := store.Expired(t.Context(), time.Now())
	if err != nil || len(expired) != 1 {
		t.Fatalf("expired=%+v err=%v", expired, err)
	}
	if err := store.PutImport(t.Context(), "child", tutor.SchoolImport{ID: "lesson", Lesson: tutor.LessonSuggestion{Title: "School"}}); err != nil {
		t.Fatal(err)
	}
	lessons, err := store.Imports(t.Context(), "child", true)
	if err != nil || len(lessons) != 0 {
		t.Fatal("draft published")
	}
	if err := store.ApproveImport(t.Context(), "other", "lesson"); !errors.Is(err, tutor.ErrInvalid) {
		t.Fatal("foreign approval allowed")
	}
	if err := store.ApproveImport(t.Context(), "child", "lesson"); err != nil {
		t.Fatal(err)
	}
	lessons, err = store.Imports(t.Context(), "child", true)
	if err != nil || len(lessons) != 1 || !lessons[0].Approved {
		t.Fatal("approval not persisted")
	}
}
