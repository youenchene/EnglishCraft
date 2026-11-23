package content

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
)

func TestVersionedLibraryAndMissionsValidate(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Verbs()) < 30 || len(c.Missions()) != 3 {
		t.Fatal("incomplete initial content")
	}
	for _, m := range c.Missions() {
		for i := range m.Length() {
			step, err := m.Step(i)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := c.Exercise(step.Challenge().ExerciseID()); !ok {
				t.Fatal("dangling challenge")
			}
		}
	}
	if _, ok := c.Exercise("train-go-build_sentence"); !ok {
		t.Fatal("open exercises missing")
	}
}

type journeyRepository struct{ state core.StudentState }

func (r *journeyRepository) Read(context.Context, string) (core.StudentState, error) {
	return r.state, nil
}
func (r *journeyRepository) Transact(_ context.Context, _ string, mutate func(*core.StudentState) error) error {
	return mutate(&r.state)
}

type journeyClock struct{}

func (journeyClock) Now() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }

func TestAllInitialMissionsCanBeCompletedWithoutAI(t *testing.T) {
	catalog, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id      string
		answers []string
		xp      int
	}{
		{id: "lost-diamond", answers: []string{"went", "found", "builded", "taken", "I found a diamond."}, xp: 65},
		{id: "mysterious-village", answers: []string{"behind", "Did", "the library", "next to"}, xp: 60},
		{id: "post-match", answers: []string{"won", "caught", "Did", "goalkeeper", "We won the match."}, xp: 70},
	} {
		t.Run(tc.id, func(t *testing.T) {
			ctx := context.Background()
			repo := &journeyRepository{state: core.NewState()}
			service := core.New(catalog, repo, journeyClock{})
			view, err := service.Start(ctx, "child", tc.id)
			if err != nil {
				t.Fatal(err)
			}
			for i, answer := range tc.answers {
				if view.Exercise == nil {
					t.Fatal("mission completed before all objectives")
				}
				result, err := service.Submit(ctx, core.Submit{Student: "child", MissionID: tc.id, Step: i, ExerciseID: view.Exercise.ID,
					Answer: answer, Key: fmt.Sprintf("journey-%s-%02d", tc.id, i)})
				if err != nil || result.Status != "correct" {
					t.Fatalf("step %d: %+v %v", i, result, err)
				}
				view, err = service.Current(ctx, "child", tc.id)
				if err != nil {
					t.Fatal(err)
				}
			}
			progress, err := service.Progress(ctx, "child")
			if err != nil || !view.Completed || progress.XP != tc.xp {
				t.Fatalf("wrong final mission state: %+v %v", progress, err)
			}
		})
	}
}
