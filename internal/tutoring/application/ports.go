package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	"time"
)

var (
	ErrInvalid       = errors.New("invalid tutoring input")
	ErrUnavailable   = errors.New("AI unavailable")
	ErrBudget        = errors.New("AI allowance exhausted")
	ErrVoiceDisabled = fmt.Errorf("voice provider disabled: %w", ErrUnavailable)
	ErrAIConsent     = fmt.Errorf("parental AI consent required: %w", ErrUnavailable)
)

// TeachingContext contains learning data, never prompts, API types or identities.
type TeachingContext struct {
	SchoolYear      string        `json:"schoolYear"`
	Level           string        `json:"level"`
	Objective       string        `json:"objective"`
	Skills          []string      `json:"skills"`
	Vocabulary      []string      `json:"vocabulary"`
	RecurringErrors []string      `json:"recurringErrors"`
	ReviewItems     []string      `json:"reviewItems"`
	Step            core.StepView `json:"step"`
	Feedback        string        `json:"feedback"`
}
type Turn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}
type ConversationRequest struct {
	Context TeachingContext `json:"context"`
	History []Turn          `json:"history"`
	Message string          `json:"message"`
}
type TutorReply struct {
	Text string `json:"text"`
	Mood string `json:"mood"`
	Hint string `json:"hint"`
}
type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}
type TutorConversation interface {
	Reply(context.Context, ConversationRequest) (TutorReply, Usage, error)
}
type LessonRequest struct {
	Text   string   `json:"text"`
	Skills []string `json:"skills"`
	Verbs  []string `json:"verbs"`
}
type LessonSuggestion struct {
	Title       string   `json:"title"`
	Explanation string   `json:"explanation"`
	Vocabulary  []string `json:"vocabulary"`
	Grammar     []string `json:"grammar"`
	Skills      []string `json:"skills"`
	Verbs       []string `json:"verbs"`
	Missions    []string `json:"missions"`
}
type ExerciseSuggestion struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
	Skill  string `json:"skill"`
}
type LessonGenerator interface {
	GenerateLesson(context.Context, LessonRequest) (LessonSuggestion, Usage, error)
}
type ExerciseGenerator interface {
	GenerateExercises(context.Context, LessonSuggestion) ([]ExerciseSuggestion, Usage, error)
}
type EvaluationRequest struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
}

// Evaluation is advice only, not an award of mastery or rewards.
type Evaluation struct {
	Feedback         string  `json:"feedback"`
	Confidence       float64 `json:"confidence"`
	NeedsHumanReview bool    `json:"needsHumanReview"`
}
type AnswerEvaluator interface {
	Evaluate(context.Context, EvaluationRequest) (Evaluation, Usage, error)
}
type SpeechRequest struct {
	Context    TeachingContext
	SDP        string
	PushToTalk bool
	MaxSeconds int
}
type SpeechSession struct {
	ID       string    `json:"id"`
	SDP      string    `json:"sdp"`
	Expires  time.Time `json:"expires"`
	Provider string    `json:"provider,omitempty"`
	Token    string    `json:"token,omitempty"`
	Greeting string    `json:"greeting,omitempty"`
}
type SpeechSessionGateway interface {
	Create(context.Context, SpeechRequest) (SpeechSession, error)
	Hangup(context.Context, string) error
}
type Learning interface {
	Current(context.Context, string, string) (core.StepView, error)
	Progress(context.Context, string) (core.Summary, error)
	Submit(context.Context, core.Submit) (core.Result, error)
	Training(context.Context, string) (*core.ExerciseView, error)
	Skills() []core.SkillView
	Verbs() []core.Verb
}
type Clock interface{ Now() time.Time }
type Allowance struct {
	Student, Provider                              string
	At                                             time.Time
	ReserveMicros, DailyMicros, StudentDailyMicros int64
	VoiceSeconds, DailyVoiceSeconds                int
}
type Spending struct {
	Provider       string `json:"provider"`
	ReservedMicros int64  `json:"reservedMicros"`
	Calls          int    `json:"calls"`
	VoiceSeconds   int    `json:"voiceSeconds"`
}
type Budget interface {
	Enabled(context.Context, string) (bool, error)
	SetEnabled(context.Context, string, bool) error
	Reserve(context.Context, Allowance) (string, error)
	Record(context.Context, string, Usage, string) error
	Spending(context.Context, string, time.Time) ([]Spending, error)
}
type VoiceRecord struct {
	Student, Mission string
	Session          SpeechSession
}
type VoiceRepository interface {
	Save(context.Context, VoiceRecord) error
	Find(context.Context, string, string) (VoiceRecord, error)
	Lookup(context.Context, string) (VoiceRecord, error)
	Expired(context.Context, time.Time) ([]VoiceRecord, error)
	Remove(context.Context, string) error
}
