package domain

import "testing"

func TestClosedAnswersAreStrictAndOpenAnswersNeedEvaluation(t *testing.T) {
	for _, tc := range []struct {
		name, answer, status string
		kind                 ExerciseKind
	}{
		{name: "case and whitespace", answer: "  WENT  ", status: "correct", kind: PastSimple},
		{name: "wrong morphology", answer: "goed", status: "incorrect", kind: PastSimple},
		{name: "extra word", answer: "I went", status: "incorrect", kind: PastSimple},
		{name: "punctuation not ignored", answer: "went!", status: "incorrect", kind: PastSimple},
		{name: "open answer", answer: "I went to the cave.", status: "needs_evaluation", kind: BuildSentence},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := NewExercise(ExerciseSpec{ID: "go", Kind: tc.kind, Prompt: "Go", Skill: "english.go", Accepted: []string{"went"}})
			if err != nil {
				t.Fatal(err)
			}
			a, err := NewStudentAnswer(tc.answer)
			if err != nil {
				t.Fatal(err)
			}
			if got := e.Assess(a).Status; got != tc.status {
				t.Fatalf("got %s want %s", got, tc.status)
			}
		})
	}
}
func TestLearningInvariants(t *testing.T) {
	if _, err := NewStudentAnswer("  "); err == nil {
		t.Fatal("empty answer accepted")
	}
	if _, err := NewExercise(ExerciseSpec{ID: "x", Kind: PastSimple, Prompt: "x", Skill: "english.x"}); err == nil {
		t.Fatal("closed exercise without correction accepted")
	}
	if _, err := NewStudent("", "Alex"); err == nil {
		t.Fatal("empty student identity")
	}
	if _, err := NewLearningSession("session", Student{}); err == nil {
		t.Fatal("session without student")
	}
}
