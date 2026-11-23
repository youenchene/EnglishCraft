package application

import (
	"strings"
	"testing"

	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
)

func TestOpeningStoriesAndRules(t *testing.T) {
	for _, test := range []struct{ mission, story string }{
		{"lost-diamond", "diamond"}, {"mysterious-village", "library"}, {"post-match", "reporter"},
	} {
		text := opening(TeachingContext{Step: core.StepView{MissionID: test.mission, Dialogue: "Alex : Réponds en français à cet exercice."}})
		if !strings.Contains(text, test.story) || !strings.Contains(text, "Speak English") || !strings.Contains(text, "one question") || !strings.Contains(text, "'help'") || !strings.Contains(text, "'repeat'") {
			t.Fatalf("missing setup for %s", test.mission)
		}
		if strings.Contains(text, "français") || strings.Contains(text, "Alex :") || strings.Count(text, "?") != 1 {
			t.Fatal("opening must be English and ask just one question")
		}
	}
}

func TestIntroPreviewDoesNotReserveOrCallModel(t *testing.T) {
	budget := &fakeBudget{}
	model := &fakeTutor{}
	service := New(Dependencies{Learning: &fakeLearning{}, Budget: budget, Tutor: model})
	text, err := service.Introduction(t.Context(), "child", "diamond")
	if err != nil || !strings.Contains(text, "Speak English") || budget.reserved != 0 || model.calls != 0 {
		t.Fatal("intro preview spent budget or called an LLM")
	}
}
