package domain

import (
	"testing"

	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

func TestIrregularVerbCheck(t *testing.T) {
	skillID, err := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	if err != nil {
		t.Fatal(err)
	}
	verb, err := NewIrregularVerb("verb-go", "go", "went", "gone", "aller", skillID)
	if err != nil {
		t.Fatal(err)
	}

	correction := verb.Check(" Went ", "GONE")
	if !correction.Correct() {
		t.Fatalf("expected normalized answer to be correct: %#v", correction)
	}
}

func TestIrregularVerbCheckIdentifiesEachMistake(t *testing.T) {
	skillID, _ := shared.NewSkillID("english.grammar.past_simple.irregular_verbs")
	verb, _ := NewIrregularVerb("verb-go", "go", "went", "gone", "aller", skillID)

	correction := verb.Check("goed", "gone")
	if correction.PastSimpleCorrect || !correction.PastParticipleCorrect || correction.Correct() {
		t.Fatalf("unexpected correction: %#v", correction)
	}
}
