package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	missions "github.com/youenchene/EnglishCraft/internal/missions/domain"
)

var (
	ErrNotFound = errors.New("content not found")
	ErrInvalid  = errors.New("invalid command")
	ErrConflict = errors.New("stale step or reused idempotency key")
)

type Verb struct {
	Base        string   `json:"base"`
	Past        string   `json:"past"`
	Participle  string   `json:"participle"`
	Translation string   `json:"translation"`
	Difficulty  int      `json:"difficulty"`
	Categories  []string `json:"categories"`
	Examples    []string `json:"examples"`
}
type Catalog interface {
	Missions() []missions.Mission
	Mission(string) (missions.Mission, bool)
	Exercise(string) (learning.Exercise, bool)
	Skills() []curriculum.Skill
	Verbs() []Verb
}
type Clock interface{ Now() time.Time }

// StudentState is a persistence snapshot. Only Service mutates it inside a
// repository transaction; adapters have no pedagogical decisions to perform.
type StudentState struct {
	Missions map[string]missions.ProgressSnapshot `json:"missions"`
	Reviews  map[string]learning.ReviewSnapshot   `json:"reviews"`
	Attempts []Attempt                            `json:"attempts"`
	Receipts map[string]Receipt                   `json:"receipts"`
}

func NewState() StudentState {
	return StudentState{Missions: map[string]missions.ProgressSnapshot{}, Reviews: map[string]learning.ReviewSnapshot{},
		Attempts: []Attempt{}, Receipts: map[string]Receipt{}}
}

type Repository interface {
	Read(context.Context, string) (StudentState, error)
	Transact(context.Context, string, func(*StudentState) error) error
}
type Attempt struct {
	ExerciseID string    `json:"exerciseId"`
	Skill      string    `json:"skill"`
	Answer     string    `json:"answer"`
	Status     string    `json:"status"`
	Expected   string    `json:"expected"`
	At         time.Time `json:"at"`
}
type Receipt struct {
	Fingerprint string `json:"fingerprint"`
	Result      Result `json:"result"`
}
type Result struct {
	Status    string `json:"status"`
	Expected  string `json:"expected"`
	XP        int    `json:"xp"`
	Completed bool   `json:"completed"`
	NextStep  int    `json:"nextStep"`
}
type ExerciseView struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Prompt string `json:"prompt"`
	Skill  string `json:"skill"`
	Verb   string `json:"verb"`
}

func exerciseView(e learning.Exercise) ExerciseView {
	return ExerciseView{ID: e.ID(), Kind: string(e.Kind()), Prompt: e.Prompt(), Skill: e.Skill(), Verb: e.Verb()}
}

type StepView struct {
	MissionID string        `json:"missionId"`
	Title     string        `json:"title"`
	Theme     string        `json:"theme"`
	Step      int           `json:"step"`
	Total     int           `json:"total"`
	Completed bool          `json:"completed"`
	Dialogue  string        `json:"dialogue"`
	Exercise  *ExerciseView `json:"exercise"`
}
type MissionView struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Theme   string `json:"theme"`
	Version int    `json:"version"`
	Steps   int    `json:"steps"`
}
type SkillView struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type Summary struct {
	XP       int                                  `json:"xp"`
	Missions map[string]missions.ProgressSnapshot `json:"missions"`
	Reviews  map[string]learning.ReviewSnapshot   `json:"reviews"`
	Due      []string                             `json:"due"`
	History  []Attempt                            `json:"history"`
	Errors   map[string]int                       `json:"errors"`
}
type Service struct {
	catalog Catalog
	repo    Repository
	clock   Clock
}

func New(catalog Catalog, repo Repository, clock Clock) *Service {
	return &Service{catalog: catalog, repo: repo, clock: clock}
}
func (s *Service) Missions() []MissionView {
	result := []MissionView{}
	for _, m := range s.catalog.Missions() {
		result = append(result, MissionView{ID: m.ID(), Title: m.Title(), Theme: m.Theme(), Version: m.Version(), Steps: m.Length()})
	}
	return result
}
func (s *Service) Verbs() []Verb { return s.catalog.Verbs() }
func (s *Service) Skills() []SkillView {
	result := []SkillView{}
	for _, skill := range s.catalog.Skills() {
		result = append(result, SkillView{ID: skill.ID(), Title: skill.Title()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
func (s *Service) Start(ctx context.Context, student, id string) (StepView, error) {
	m, ok := s.catalog.Mission(id)
	if !ok {
		return StepView{}, ErrNotFound
	}
	if student == "" {
		return StepView{}, ErrInvalid
	}
	err := s.repo.Transact(ctx, student, func(state *StudentState) error {
		if _, exists := state.Missions[id]; !exists {
			state.Missions[id] = missions.Start(m).Snapshot()
		}
		return nil
	})
	if err != nil {
		return StepView{}, err
	}
	return s.Current(ctx, student, id)
}
func (s *Service) Current(ctx context.Context, student, id string) (StepView, error) {
	m, ok := s.catalog.Mission(id)
	if !ok {
		return StepView{}, ErrNotFound
	}
	state, err := s.repo.Read(ctx, student)
	if err != nil {
		return StepView{}, err
	}
	snapshot, ok := state.Missions[id]
	if !ok {
		return StepView{}, ErrNotFound
	}
	progress, err := missions.RestoreProgress(m, snapshot)
	if err != nil {
		return StepView{}, err
	}
	view := StepView{MissionID: id, Title: m.Title(), Theme: m.Theme(), Step: snapshot.Step,
		Total: m.Length(), Completed: progress.Snapshot().Completed}
	if view.Completed {
		return view, nil
	}
	step, err := m.Step(snapshot.Step)
	if err != nil {
		return StepView{}, err
	}
	e, ok := s.catalog.Exercise(step.Challenge().ExerciseID())
	if !ok {
		return StepView{}, ErrNotFound
	}
	exercise := exerciseView(e)
	view.Dialogue, view.Exercise = step.Dialogue(), &exercise
	return view, nil
}

type Submit struct {
	Student    string
	MissionID  string
	Step       int
	ExerciseID string
	Answer     string
	Key        string
}

func (s *Service) Submit(ctx context.Context, command Submit) (Result, error) {
	validKey := len(command.Key) >= 16 && len(command.Key) <= 100
	if command.Student == "" || !validKey || command.Step < 0 {
		return Result{}, ErrInvalid
	}
	answer, err := learning.NewStudentAnswer(command.Answer)
	if err != nil {
		return Result{}, ErrInvalid
	}
	e, ok := s.catalog.Exercise(command.ExerciseID)
	if !ok {
		return Result{}, ErrNotFound
	}
	if command.MissionID == "" && !strings.HasPrefix(e.ID(), "train-") {
		return Result{}, ErrInvalid
	}
	hash := sha256.Sum256([]byte(command.MissionID + "\x00" + command.ExerciseID + "\x00" + command.Answer))
	fingerprint := hex.EncodeToString(hash[:]) + ":" + strconv.Itoa(command.Step)
	var result Result
	err = s.repo.Transact(ctx, command.Student, func(state *StudentState) error {
		if receipt, exists := state.Receipts[command.Key]; exists {
			if receipt.Fingerprint != fingerprint {
				return ErrConflict
			}
			result = receipt.Result
			return nil
		}
		assessment := e.Assess(answer)
		result = Result{Status: assessment.Status, Expected: assessment.Expected}
		if command.MissionID != "" {
			m, ok := s.catalog.Mission(command.MissionID)
			if !ok {
				return ErrNotFound
			}
			snapshot, ok := state.Missions[m.ID()]
			if !ok {
				return ErrNotFound
			}
			p, err := missions.RestoreProgress(m, snapshot)
			if err != nil {
				return err
			}
			step, err := m.Step(snapshot.Step)
			if err != nil {
				return ErrConflict
			}
			if snapshot.Step != command.Step || step.Challenge().ExerciseID() != e.ID() {
				return ErrConflict
			}
			xp, err := p.CompleteStep(m, command.Step, assessment.Correct())
			if err != nil {
				return err
			}
			state.Missions[m.ID()] = p.Snapshot()
			result.XP, result.Completed, result.NextStep = xp, p.Snapshot().Completed, p.Snapshot().Step
		}
		now := s.clock.Now()
		if assessment.Status != "needs_evaluation" {
			review := learning.NewReviewSchedule()
			if snapshot, exists := state.Reviews[e.Skill()]; exists {
				var err error
				review, err = learning.RestoreReview(snapshot)
				if err != nil {
					return err
				}
			}
			if err := review.Record(assessment.Correct(), now); err != nil {
				return err
			}
			state.Reviews[e.Skill()] = review.Snapshot()
		}
		state.Attempts = append(state.Attempts, Attempt{ExerciseID: e.ID(), Skill: e.Skill(), Answer: answer.Text(),
			Status: assessment.Status, Expected: assessment.Expected, At: now})
		state.Receipts[command.Key] = Receipt{Fingerprint: fingerprint, Result: result}
		return nil
	})
	return result, err
}

func (s *Service) Progress(ctx context.Context, student string) (Summary, error) {
	state, err := s.repo.Read(ctx, student)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{Missions: state.Missions, Reviews: state.Reviews, Due: []string{},
		History: []Attempt{}, Errors: map[string]int{}}
	for _, p := range state.Missions {
		summary.XP += p.XP
	}
	for skill, review := range state.Reviews {
		if !review.Due.After(s.clock.Now()) {
			summary.Due = append(summary.Due, skill)
		}
	}
	sort.Slice(summary.Due, func(i, j int) bool {
		a, b := state.Reviews[summary.Due[i]], state.Reviews[summary.Due[j]]
		if a.Failures != b.Failures {
			return a.Failures > b.Failures
		}
		if !a.Due.Equal(b.Due) {
			return a.Due.Before(b.Due)
		}
		return summary.Due[i] < summary.Due[j]
	})
	for _, a := range state.Attempts {
		if a.Status == "incorrect" {
			summary.Errors[a.Skill]++
		}
	}
	for i := len(state.Attempts) - 1; i >= 0 && len(summary.History) < 20; i-- {
		summary.History = append(summary.History, state.Attempts[i])
	}
	return summary, nil
}

// Training prioritises overdue failures, then new verbs. If every verb is in
// cooldown, no exercise is returned; mission retries remain available.
func (s *Service) Training(ctx context.Context, student string) (*ExerciseView, error) {
	summary, err := s.Progress(ctx, student)
	if err != nil {
		return nil, err
	}
	for _, skill := range summary.Due {
		base := strings.TrimPrefix(skill, "english.irregular.")
		if base != skill {
			return s.trainingView(base, summary.Reviews[skill].Attempts), nil
		}
	}
	for _, v := range s.catalog.Verbs() {
		if _, exists := summary.Reviews["english.irregular."+v.Base]; !exists {
			return s.trainingView(v.Base, 0), nil
		}
	}
	return nil, nil
}
func (s *Service) trainingView(base string, attempts int) *ExerciseView {
	kind := "past_simple"
	if attempts%2 == 1 {
		kind = "past_participle"
	}
	e, ok := s.catalog.Exercise("train-" + base + "-" + kind)
	if !ok {
		return nil
	}
	view := exerciseView(e)
	return &view
}
