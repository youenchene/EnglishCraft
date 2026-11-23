package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type ExerciseKind string

const (
	PastSimple     ExerciseKind = "past_simple"
	PastParticiple ExerciseKind = "past_participle"
	FillBlank      ExerciseKind = "fill_blank"
	IncorrectForm  ExerciseKind = "incorrect_form"
	BuildSentence  ExerciseKind = "build_sentence"
)

type StudentAnswer struct{ text string }

func NewStudentAnswer(text string) (StudentAnswer, error) {
	text = strings.TrimSpace(text)
	if text == "" || utf8.RuneCountInString(text) > 500 {
		return StudentAnswer{}, errors.New("invalid student answer")
	}
	return StudentAnswer{text: text}, nil
}

func (a StudentAnswer) Text() string { return a.text }

type Assessment struct {
	Status   string
	Expected string
}

func (a Assessment) Correct() bool { return a.Status == "correct" }

type Exercise struct {
	id       string
	kind     ExerciseKind
	prompt   string
	skill    string
	verb     string
	accepted []string
}

type ExerciseSpec struct {
	ID       string
	Kind     ExerciseKind
	Prompt   string
	Skill    string
	Verb     string
	Accepted []string
}

func NewExercise(spec ExerciseSpec) (Exercise, error) {
	validKind := spec.Kind == PastSimple || spec.Kind == PastParticiple || spec.Kind == FillBlank ||
		spec.Kind == IncorrectForm || spec.Kind == BuildSentence
	if spec.ID == "" || spec.Prompt == "" || !strings.HasPrefix(spec.Skill, "english.") || !validKind {
		return Exercise{}, errors.New("invalid exercise")
	}
	if spec.Kind != BuildSentence && len(spec.Accepted) == 0 {
		return Exercise{}, errors.New("closed exercise needs an answer")
	}
	for _, answer := range spec.Accepted {
		if strings.TrimSpace(answer) == "" {
			return Exercise{}, errors.New("empty accepted answer")
		}
	}
	return Exercise{id: spec.ID, kind: spec.Kind, prompt: spec.Prompt, skill: spec.Skill,
		verb: spec.Verb, accepted: append([]string{}, spec.Accepted...)}, nil
}

func (e Exercise) ID() string         { return e.id }
func (e Exercise) Kind() ExerciseKind { return e.kind }
func (e Exercise) Prompt() string     { return e.prompt }
func (e Exercise) Skill() string      { return e.skill }
func (e Exercise) Verb() string       { return e.verb }
func (e Exercise) Assess(a StudentAnswer) Assessment {
	if e.kind == BuildSentence {
		return Assessment{Status: "needs_evaluation"}
	}
	for _, accepted := range e.accepted {
		if Normalize(a.Text()) == Normalize(accepted) {
			return Assessment{Status: "correct", Expected: strings.Join(e.accepted, " / ")}
		}
	}
	return Assessment{Status: "incorrect", Expected: strings.Join(e.accepted, " / ")}
}

// Normalize changes case and whitespace only: no fuzzy matching, punctuation
// deletion or accent stripping that could turn a wrong answer into a right one.
func Normalize(text string) string { return strings.ToLower(strings.Join(strings.Fields(text), " ")) }

type Student struct{ id, name string }

func NewStudent(id, name string) (Student, error) {
	if id == "" || strings.TrimSpace(name) == "" {
		return Student{}, errors.New("invalid student")
	}
	return Student{id: id, name: name}, nil
}
func (s Student) ID() string   { return s.id }
func (s Student) Name() string { return s.name }

type LearningSession struct {
	student Student
	id      string
}

func NewLearningSession(id string, student Student) (LearningSession, error) {
	if id == "" || student.ID() == "" {
		return LearningSession{}, errors.New("invalid learning session")
	}
	return LearningSession{id: id, student: student}, nil
}
func (s LearningSession) ID() string       { return s.id }
func (s LearningSession) Student() Student { return s.student }
