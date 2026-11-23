package application

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Policy struct {
	Provider                                                               string
	VoiceProvider                                                          string
	DailyMicros, StudentDailyMicros, TextReserveMicros, VoiceReserveMicros int64
	VoiceMaxSeconds, VoiceDailySeconds                                     int
}
type Service struct {
	learning Learning
	tutor    TutorConversation
	speech   SpeechSessionGateway
	budget   Budget
	voices   VoiceRepository
	clock    Clock
	policy   Policy
}
type Dependencies struct {
	Learning Learning
	Tutor    TutorConversation
	Speech   SpeechSessionGateway
	Budget   Budget
	Voices   VoiceRepository
	Clock    Clock
	Policy   Policy
}

func New(d Dependencies) *Service {
	return &Service{learning: d.Learning, tutor: d.Tutor, speech: d.Speech, budget: d.Budget,
		voices: d.Voices, clock: d.Clock, policy: d.Policy}
}
func (s *Service) Context(ctx context.Context, student, mission string) (TeachingContext, error) {
	step, err := s.learning.Current(ctx, student, mission)
	if err != nil {
		return TeachingContext{}, err
	}
	summary, err := s.learning.Progress(ctx, student)
	if err != nil {
		return TeachingContext{}, err
	}
	recurring := []string{}
	for skill, count := range summary.Errors {
		if count >= 2 && summary.Reviews[skill].Stage != "mastery" {
			recurring = append(recurring, skill)
		}
	}
	sort.Strings(recurring)
	level := "A1"
	// A local scaffolding estimate, never a durable CEFR certification.
	if len(summary.History) >= 5 {
		correct := 0
		for _, a := range summary.History[:5] {
			if a.Status == "correct" {
				correct++
			}
		}
		if correct >= 4 {
			level = "A2"
		}
	}
	skills, vocabulary := []string{}, []string{}
	if step.Exercise != nil {
		skills = append(skills, step.Exercise.Skill)
		if step.Exercise.Verb != "" {
			vocabulary = append(vocabulary, step.Exercise.Verb)
		}
	}
	return TeachingContext{SchoolYear: "collège · 13 ans", Level: level, Objective: step.Title,
		Skills: skills, Vocabulary: vocabulary, RecurringErrors: recurring, ReviewItems: summary.Due,
		Step: step, Feedback: "Hint before the answer; short encouraging English; French only when explicitly requested."}, nil
}

type EventType string

const (
	CharacterSpeak   EventType = "CHARACTER_SPEAK"
	ShowHint         EventType = "SHOW_HINT"
	OpenDoor         EventType = "OPEN_DOOR"
	GiveDiamond      EventType = "GIVE_DIAMOND"
	ScoreGoal        EventType = "SCORE_GOAL"
	AwardXP          EventType = "AWARD_XP"
	MissionCompleted EventType = "MISSION_COMPLETED"
)

type Event struct {
	Version int       `json:"version"`
	Type    EventType `json:"type"`
	Text    string    `json:"text,omitempty"`
	Amount  int       `json:"amount,omitempty"`
}
type ConversationResult struct {
	TutorReply
	Mode   string  `json:"mode"`
	Events []Event `json:"events"`
}

func (s *Service) Converse(ctx context.Context, student, mission, message string) (ConversationResult, error) {
	return s.converse(ctx, student, mission, message, nil)
}
func (s *Service) converse(ctx context.Context, student, mission, message string, history []Turn) (ConversationResult, error) {
	if !validText(message, 1000) {
		return ConversationResult{}, ErrInvalid
	}
	teaching, err := s.Context(ctx, student, mission)
	if err != nil {
		return ConversationResult{}, err
	}
	_, question := adventureStory(teaching)
	fallback := ConversationResult{TutorReply: TutorReply{Text: "The live tutor is unavailable right now. " + question,
		Mood: "encouraging", Hint: "Try a short English sentence, or use local revisions while you wait."},
		Mode: "local", Events: []Event{}}
	if s.tutor == nil {
		return fallback, nil
	}
	enabled, err := s.budget.Enabled(ctx, student)
	if err != nil {
		return ConversationResult{}, err
	}
	if !enabled {
		return fallback, nil
	}
	id, err := s.reserve(ctx, student, s.policy.Provider, s.policy.TextReserveMicros, 0)
	if err != nil {
		if errors.Is(err, ErrBudget) {
			return fallback, nil
		}
		return ConversationResult{}, err
	}
	reply, usage, callErr := s.tutor.Reply(ctx, ConversationRequest{Context: teaching, History: history, Message: message})
	status := "ok"
	if callErr != nil {
		status = "failed"
	}
	// Uncertain requests retain their reservation; they may have been billed.
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := s.budget.Record(auditCtx, id, usage, status); err != nil {
		return ConversationResult{}, err
	}
	if callErr != nil {
		if ctx.Err() != nil {
			return ConversationResult{}, ctx.Err()
		}
		return fallback, nil
	}
	if !validText(reply.Text, 2000) {
		return fallback, nil
	}
	return ConversationResult{TutorReply: reply, Mode: "ai", Events: []Event{
		{Version: 1, Type: "CHARACTER_SPEAK", Text: reply.Text},
	}}, nil
}
func (s *Service) reserve(ctx context.Context, student, provider string, micros int64, seconds int) (string, error) {
	return s.budget.Reserve(ctx, Allowance{Student: student, Provider: provider, At: s.clock.Now(),
		ReserveMicros: micros, DailyMicros: s.policy.DailyMicros, StudentDailyMicros: s.policy.StudentDailyMicros,
		VoiceSeconds: seconds, DailyVoiceSeconds: s.policy.VoiceDailySeconds})
}
func validText(text string, max int) bool {
	return utf8.ValidString(text) && strings.TrimSpace(text) != "" && len(text) <= max && !strings.ContainsRune(text, 0)
}

type ToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type ToolResult struct {
	Value  any     `json:"value"`
	Events []Event `json:"events"`
}

func (s *Service) Tool(ctx context.Context, student, mission string, request ToolRequest) (ToolResult, error) {
	result := ToolResult{Events: []Event{}}
	if len(request.Arguments) > 2048 {
		return result, ErrInvalid
	}
	if request.Name == "submit_student_answer" {
		var args struct {
			Step       *int   `json:"step"`
			ExerciseID string `json:"exerciseId"`
			Answer     string `json:"answer"`
			Key        string `json:"idempotencyKey"`
		}
		if err := StrictJSON(request.Arguments, &args); err != nil || args.Step == nil {
			return result, ErrInvalid
		}
		outcome, err := s.learning.Submit(ctx, core.Submit{Student: student, MissionID: mission,
			Step: *args.Step, ExerciseID: args.ExerciseID, Answer: args.Answer, Key: args.Key})
		if err != nil {
			return result, err
		}
		result.Value = outcome
		if outcome.Status == "correct" {
			result.Events = append(result.Events, Event{Version: 1, Type: "OPEN_DOOR"})
			if outcome.XP > 0 {
				result.Events = append(result.Events, Event{Version: 1, Type: "AWARD_XP", Amount: outcome.XP})
			}
			if outcome.Completed {
				result.Events = append(result.Events, Event{Version: 1, Type: "MISSION_COMPLETED"})
			}
		}
		return result, nil
	}
	var empty struct{}
	if err := StrictJSON(request.Arguments, &empty); err != nil {
		return result, ErrInvalid
	}
	switch request.Name {
	case "get_current_objective", "get_current_challenge", "complete_mission_step":
		// This reads the step already advanced by Submit; it cannot bypass evidence.
		value, err := s.learning.Current(ctx, student, mission)
		result.Value = value
		return result, err
	case "get_review_items":
		value, err := s.learning.Training(ctx, student)
		result.Value = value
		return result, err
	case "request_hint":
		value, err := s.learning.Current(ctx, student, mission)
		if err != nil {
			return result, err
		}
		hint := "Repère le temps demandé. Dis d'abord une courte phrase en anglais."
		if value.Exercise != nil && value.Exercise.Verb != "" {
			hint = "Cherche la forme du verbe « " + value.Exercise.Verb + " » au temps demandé."
		}
		result.Value = hint
		result.Events = append(result.Events, Event{Version: 1, Type: "SHOW_HINT", Text: hint})
		return result, nil
	default:
		return result, ErrInvalid
	}
}

// VoiceReady checks provider, parental consent and authorised teaching context
// before the browser asks for microphone permission. StartVoice checks again.
func (s *Service) VoiceReady(ctx context.Context, student, mission string) error {
	if s.speech == nil {
		return ErrVoiceDisabled
	}
	enabled, err := s.budget.Enabled(ctx, student)
	if err != nil {
		return err
	}
	if !enabled {
		return ErrAIConsent
	}
	_, err = s.Context(ctx, student, mission)
	return err
}
func (s *Service) StartVoice(ctx context.Context, student, mission, sdp string, pushToTalk bool) (SpeechSession, error) {
	if (s.VoiceProvider() == "elevenlabs" && sdp != "") || (s.VoiceProvider() != "elevenlabs" && (!strings.HasPrefix(sdp, "v=0") || len(sdp) > 64000)) {
		return SpeechSession{}, ErrInvalid
	}
	if err := s.VoiceReady(ctx, student, mission); err != nil {
		return SpeechSession{}, err
	}
	teaching, err := s.Context(ctx, student, mission)
	if err != nil {
		return SpeechSession{}, err
	}
	if _, err := s.reserve(ctx, student, s.VoiceProvider(), s.policy.VoiceReserveMicros, s.policy.VoiceMaxSeconds); err != nil {
		return SpeechSession{}, err
	}
	session, err := s.speech.Create(ctx, SpeechRequest{Context: teaching, SDP: sdp, PushToTalk: pushToTalk, MaxSeconds: s.policy.VoiceMaxSeconds})
	if err != nil {
		return SpeechSession{}, err
	}
	session.Expires = s.clock.Now().Add(time.Duration(s.policy.VoiceMaxSeconds) * time.Second)
	session.Provider = s.VoiceProvider()
	session.Greeting = opening(teaching)
	stored := SpeechSession{ID: session.ID, Expires: session.Expires}
	if err := s.voices.Save(ctx, VoiceRecord{Student: student, Mission: mission, Session: stored}); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.speech.Hangup(cleanup, session.ID)
		return SpeechSession{}, err
	}
	return session, nil
}

func (s *Service) VoiceProvider() string {
	if s.policy.VoiceProvider == "" {
		return "openai"
	}
	return s.policy.VoiceProvider
}

// AdmitVoice is only consumed by the authenticated provider callback, never an
// identity-bearing browser DTO. The persisted admission supplies student/mission.
func (s *Service) AdmitVoice(ctx context.Context, id string) (VoiceRecord, error) {
	if s.VoiceProvider() != "elevenlabs" || s.speech == nil {
		return VoiceRecord{}, ErrVoiceDisabled
	}
	record, err := s.voices.Lookup(ctx, id)
	if err != nil {
		return VoiceRecord{}, err
	}
	if !s.clock.Now().Before(record.Session.Expires) {
		return VoiceRecord{}, ErrInvalid
	}
	if err := s.VoiceReady(ctx, record.Student, record.Mission); err != nil {
		return VoiceRecord{}, err
	}
	return record, nil
}

// VoiceReply uses the same consent, budget, teaching context and provider audit
// as text tutoring. Transcripts do not submit answers or award learning evidence.
func (s *Service) VoiceReply(ctx context.Context, id string, turns []Turn) (string, error) {
	record, err := s.AdmitVoice(ctx, id)
	if err != nil {
		return "", err
	}
	if len(turns) == 0 || len(turns) > 12 || turns[len(turns)-1].Role != "user" {
		return "", ErrInvalid
	}
	for _, turn := range turns {
		if (turn.Role != "user" && turn.Role != "assistant") || !validText(turn.Text, 1000) {
			return "", ErrInvalid
		}
	}
	result, err := s.converse(ctx, record.Student, record.Mission, turns[len(turns)-1].Text, turns[:len(turns)-1])
	if err != nil {
		return "", err
	}
	if result.Mode != "ai" {
		return "", ErrUnavailable
	}
	return result.Text, nil
}
func (s *Service) StopVoice(ctx context.Context, student, id string) error {
	if s.speech == nil {
		return ErrUnavailable
	}
	record, err := s.voices.Find(ctx, student, id)
	if err != nil {
		return err
	}
	return s.endVoice(ctx, record.Session.ID)
}

func (s *Service) endVoice(ctx context.Context, id string) error {
	// Revoke pending ElevenLabs tokens before closing the local callback. A
	// concurrent reconnect then cannot find an admission, even after restart.
	if s.VoiceProvider() == "elevenlabs" {
		if err := s.voices.Remove(ctx, id); err != nil {
			return err
		}
		return s.speech.Hangup(ctx, id)
	}
	if err := s.speech.Hangup(ctx, id); err != nil {
		return err
	}
	return s.voices.Remove(ctx, id)
}

// ReapVoice is process-driven. Stored deadlines survive process restarts.
func (s *Service) ReapVoice(ctx context.Context) error {
	if s.speech == nil {
		return nil
	}
	records, err := s.voices.Expired(ctx, s.clock.Now())
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := s.endVoice(ctx, record.Session.ID); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) Spending(ctx context.Context, student string) ([]Spending, error) {
	return s.budget.Spending(ctx, student, s.clock.Now())
}
func (s *Service) AIEnabled(ctx context.Context, student string) (bool, error) {
	return s.budget.Enabled(ctx, student)
}
func (s *Service) SetAIEnabled(ctx context.Context, student string, enabled bool) error {
	if err := s.budget.SetEnabled(ctx, student, enabled); err != nil {
		return err
	}
	if !enabled {
		return s.CloseStudentVoice(ctx, student)
	}
	return nil
}
func (s *Service) CloseStudentVoice(ctx context.Context, student string) error {
	if s.speech == nil {
		return nil
	}
	records, err := s.voices.Expired(ctx, s.clock.Now().Add(24*time.Hour))
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.Student != student {
			continue
		}
		if err := s.endVoice(ctx, record.Session.ID); err != nil {
			return err
		}
	}
	return nil
}
