package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

type SchoolImport struct {
	ID        string               `json:"id"`
	Lesson    LessonSuggestion     `json:"lesson"`
	Exercises []ExerciseSuggestion `json:"exercises"`
	Approved  bool                 `json:"approved"`
	Source    string               `json:"source"`
}
type ImportRepository interface {
	PutImport(context.Context, string, SchoolImport) error
	Imports(context.Context, string, bool) ([]SchoolImport, error)
	ApproveImport(context.Context, string, string) error
}
type Imports struct {
	repo     ImportRepository
	learning Learning
}

func NewImports(repo ImportRepository, learning Learning) *Imports {
	return &Imports{repo: repo, learning: learning}
}

// ImportText deliberately extracts only known catalogue items. No document is
// sent to a provider, executed, or promoted to a teaching truth automatically.
func (s *Imports) ImportText(ctx context.Context, student, text string) (SchoolImport, error) {
	if !validText(text, 20000) {
		return SchoolImport{}, ErrInvalid
	}
	result := SchoolImport{Lesson: LessonSuggestion{Title: "Cours du collège", Explanation: "Extraction locale à vérifier avec le parent.",
		Vocabulary: []string{}, Grammar: []string{}, Skills: []string{}, Verbs: []string{}, Missions: []string{}},
		Exercises: []ExerciseSuggestion{}, Source: "local"}
	words := map[string]bool{}
	for _, word := range regexp.MustCompile(`[A-Za-z]+`).FindAllString(strings.ToLower(text), -1) {
		words[word] = true
	}
	for _, v := range s.learning.Verbs() {
		if words[v.Base] || words[v.Past] || words[v.Participle] {
			skill := "english.irregular." + v.Base
			result.Lesson.Verbs = append(result.Lesson.Verbs, v.Base)
			result.Lesson.Skills = append(result.Lesson.Skills, skill)
			result.Lesson.Vocabulary = append(result.Lesson.Vocabulary, v.Base+" : "+v.Translation)
			result.Exercises = append(result.Exercises, ExerciseSuggestion{Prompt: "Prétérit de " + v.Base, Answer: v.Past, Skill: skill})
		}
	}
	for _, rule := range []string{"past simple", "present perfect", "past participle", "present simple"} {
		if strings.Contains(strings.ToLower(text), rule) {
			result.Lesson.Grammar = append(result.Lesson.Grammar, rule)
		}
	}
	if len(result.Lesson.Verbs) > 0 {
		result.Lesson.Missions = append(result.Lesson.Missions, "Réviser les verbes importés, puis reprendre The Lost Diamond.")
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return SchoolImport{}, err
	}
	result.ID = hex.EncodeToString(raw)
	return result, s.repo.PutImport(ctx, student, result)
}
func (s *Imports) List(ctx context.Context, student string, approvedOnly bool) ([]SchoolImport, error) {
	return s.repo.Imports(ctx, student, approvedOnly)
}
func (s *Imports) Approve(ctx context.Context, student, id string) error {
	return s.repo.ApproveImport(ctx, student, id)
}
