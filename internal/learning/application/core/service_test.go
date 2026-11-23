package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	missions "github.com/youenchene/EnglishCraft/internal/missions/domain"
)

type memory struct {
	state StudentState
	fail  bool
}

func (m *memory) Read(context.Context, string) (StudentState, error) { return m.state, nil }
func (m *memory) Transact(_ context.Context, _ string, fn func(*StudentState) error) error {
	data, err := json.Marshal(m.state)
	if err != nil {
		return err
	}
	copy := NewState()
	if err := json.Unmarshal(data, &copy); err != nil {
		return err
	}
	if err := fn(&copy); err != nil {
		return err
	}
	if m.fail {
		return errors.New("write failure")
	}
	m.state = copy
	return nil
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type catalog struct {
	mission   missions.Mission
	exercises map[string]learning.Exercise
}

func (c catalog) Missions() []missions.Mission               { return []missions.Mission{c.mission} }
func (c catalog) Mission(id string) (missions.Mission, bool) { return c.mission, id == c.mission.ID() }
func (c catalog) Exercise(id string) (learning.Exercise, bool) {
	e, ok := c.exercises[id]
	return e, ok
}
func (c catalog) Skills() []curriculum.Skill { return []curriculum.Skill{} }
func (c catalog) Verbs() []Verb              { return []Verb{{Base: "go"}, {Base: "find"}} }
func fixture(t *testing.T) (*Service, *memory, *fakeClock) {
	t.Helper()
	exercises := map[string]learning.Exercise{}
	for _, spec := range []learning.ExerciseSpec{
		{ID: "go", Kind: learning.PastSimple, Prompt: "go", Skill: "english.irregular.go", Verb: "go", Accepted: []string{"went"}},
		{ID: "find", Kind: learning.PastParticiple, Prompt: "find", Skill: "english.irregular.find", Verb: "find", Accepted: []string{"found"}},
		{ID: "train-go-past_simple", Kind: learning.PastSimple, Prompt: "go", Skill: "english.irregular.go", Verb: "go", Accepted: []string{"went"}},
		{ID: "train-go-past_participle", Kind: learning.PastParticiple, Prompt: "go", Skill: "english.irregular.go", Verb: "go", Accepted: []string{"gone"}},
		{ID: "train-find-past_simple", Kind: learning.PastSimple, Prompt: "find", Skill: "english.irregular.find", Verb: "find", Accepted: []string{"found"}},
		{ID: "train-go-build_sentence", Kind: learning.BuildSentence, Prompt: "go", Skill: "english.irregular.go", Verb: "go"},
	} {
		e, err := learning.NewExercise(spec)
		if err != nil {
			t.Fatal(err)
		}
		exercises[e.ID()] = e
	}
	steps := []missions.MissionStep{}
	for _, id := range []string{"go", "find"} {
		ch, err := missions.NewChallenge(id)
		if err != nil {
			t.Fatal(err)
		}
		r, err := missions.NewReward(10)
		if err != nil {
			t.Fatal(err)
		}
		st, err := missions.NewMissionStep("Alex", ch, r)
		if err != nil {
			t.Fatal(err)
		}
		steps = append(steps, st)
	}
	m, err := missions.NewMission("diamond", "Diamond", "forest", 1, steps)
	if err != nil {
		t.Fatal(err)
	}
	repo := &memory{state: NewState()}
	clock := &fakeClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	return New(catalog{mission: m, exercises: exercises}, repo, clock), repo, clock
}
func TestMissionJourneyIdempotenceAndPersistence(t *testing.T) {
	s, repo, _ := fixture(t)
	ctx := context.Background()
	if _, err := s.Start(ctx, "child", "diamond"); err != nil {
		t.Fatal(err)
	}
	command := Submit{Student: "child", MissionID: "diamond", Step: 0, ExerciseID: "go", Answer: "goed", Key: "wrong-answer-0001"}
	wrong, err := s.Submit(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	if wrong.XP != 0 || wrong.Status != "incorrect" || wrong.Expected != "went" {
		t.Fatalf("bad correction %+v", wrong)
	}
	command.Answer, command.Key = " WENT ", "right-answer-0001"
	right, err := s.Submit(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.Submit(ctx, command)
	if err != nil || retry != right {
		t.Fatal("retry is not idempotent")
	}
	command.Answer = "wrong"
	if _, err := s.Submit(ctx, command); !errors.Is(err, ErrConflict) {
		t.Fatal("key reused with changed payload")
	}
	command.Answer, command.Key = "went", "stale-answer-0001"
	if _, err := s.Submit(ctx, command); !errors.Is(err, ErrConflict) {
		t.Fatal("stale step accepted")
	}
	result, err := s.Submit(ctx, Submit{Student: "child", MissionID: "diamond", Step: 1, ExerciseID: "find", Answer: "found", Key: "right-answer-0002"})
	if err != nil || !result.Completed {
		t.Fatalf("mission not complete: %+v %v", result, err)
	}
	if len(repo.state.Attempts) != 3 {
		t.Fatal("duplicate attempts persisted")
	}
	restarted := New(s.catalog, repo, s.clock)
	view, err := restarted.Start(ctx, "child", "diamond")
	if err != nil || !view.Completed {
		t.Fatal("completed mission reset")
	}
	progress, err := restarted.Progress(ctx, "child")
	if err != nil || progress.XP != 20 || len(progress.Errors) != 1 {
		t.Fatal("progress lost")
	}
}
func TestOpenAnswerRollbackAndReviewPriority(t *testing.T) {
	s, repo, clock := fixture(t)
	ctx := context.Background()
	command := Submit{Student: "child", ExerciseID: "train-go-build_sentence", Answer: "I went home.", Key: "open-answer-0001"}
	result, err := s.Submit(ctx, command)
	if err != nil || result.Status != "needs_evaluation" || len(repo.state.Reviews) != 0 {
		t.Fatal("open answer grants mastery")
	}
	repo.fail = true
	command.ExerciseID, command.Answer, command.Key = "train-go-past_simple", "goed", "closed-answer-0001"
	if _, err := s.Submit(ctx, command); err == nil || len(repo.state.Attempts) != 1 {
		t.Fatal("failed persistence changed state")
	}
	repo.fail = false
	if _, err := s.Submit(ctx, command); err != nil {
		t.Fatal(err)
	}
	view, err := s.Training(ctx, "child")
	if err != nil || view.Verb != "find" {
		t.Fatal("cooldown did not avoid repetition")
	}
	clock.now = clock.now.Add(4 * time.Hour)
	view, err = s.Training(ctx, "child")
	if err != nil || view.Verb != "go" {
		t.Fatal("forgotten verb not prioritised")
	}
}
