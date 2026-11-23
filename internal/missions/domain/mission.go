package domain

import "errors"

type Reward struct{ xp int }

func NewReward(xp int) (Reward, error) {
	if xp < 1 || xp > 100 {
		return Reward{}, errors.New("invalid reward")
	}
	return Reward{xp: xp}, nil
}
func (r Reward) XP() int { return r.xp }

type Challenge struct{ exerciseID string }

func NewChallenge(id string) (Challenge, error) {
	if id == "" {
		return Challenge{}, errors.New("invalid challenge")
	}
	return Challenge{exerciseID: id}, nil
}
func (c Challenge) ExerciseID() string { return c.exerciseID }

type MissionStep struct {
	dialogue  string
	challenge Challenge
	reward    Reward
}

func NewMissionStep(dialogue string, challenge Challenge, reward Reward) (MissionStep, error) {
	if dialogue == "" || challenge.ExerciseID() == "" || reward.XP() == 0 {
		return MissionStep{}, errors.New("invalid mission step")
	}
	return MissionStep{dialogue: dialogue, challenge: challenge, reward: reward}, nil
}
func (s MissionStep) Dialogue() string     { return s.dialogue }
func (s MissionStep) Challenge() Challenge { return s.challenge }
func (s MissionStep) Reward() Reward       { return s.reward }

type Mission struct {
	id, title, theme string
	version          int
	steps            []MissionStep
}

func NewMission(id, title, theme string, version int, steps []MissionStep) (Mission, error) {
	if id == "" || title == "" || theme == "" || version < 1 || len(steps) == 0 {
		return Mission{}, errors.New("invalid mission")
	}
	for _, step := range steps {
		if step.Challenge().ExerciseID() == "" {
			return Mission{}, errors.New("invalid step")
		}
	}
	return Mission{id: id, title: title, theme: theme, version: version, steps: append([]MissionStep{}, steps...)}, nil
}
func (m Mission) ID() string    { return m.id }
func (m Mission) Title() string { return m.title }
func (m Mission) Theme() string { return m.theme }
func (m Mission) Version() int  { return m.version }
func (m Mission) Length() int   { return len(m.steps) }
func (m Mission) Step(index int) (MissionStep, error) {
	if index < 0 || index >= len(m.steps) {
		return MissionStep{}, errors.New("invalid step index")
	}
	return m.steps[index], nil
}

type ProgressSnapshot struct {
	Version   int  `json:"version"`
	Step      int  `json:"step"`
	XP        int  `json:"xp"`
	Completed bool `json:"completed"`
}
type MissionProgress struct {
	missionID string
	state     ProgressSnapshot
}

func Start(m Mission) MissionProgress {
	return MissionProgress{missionID: m.ID(), state: ProgressSnapshot{Version: m.Version()}}
}
func RestoreProgress(m Mission, state ProgressSnapshot) (MissionProgress, error) {
	validPosition := state.Step >= 0 && state.Step <= m.Length() && state.Completed == (state.Step == m.Length())
	expectedXP := 0
	if validPosition {
		for i := range state.Step {
			step, err := m.Step(i)
			if err != nil {
				return MissionProgress{}, err
			}
			expectedXP += step.Reward().XP()
		}
	}
	if state.Version != m.Version() || state.XP != expectedXP || !validPosition {
		return MissionProgress{}, errors.New("invalid mission progress")
	}
	return MissionProgress{missionID: m.ID(), state: state}, nil
}
func (p MissionProgress) Snapshot() ProgressSnapshot { return p.state }
func (p *MissionProgress) CompleteStep(m Mission, expectedStep int, correct bool) (int, error) {
	if p.missionID != m.ID() || p.state.Version != m.Version() || p.state.Completed || p.state.Step != expectedStep {
		return 0, errors.New("stale mission step")
	}
	if !correct {
		return 0, nil
	}
	step, err := m.Step(p.state.Step)
	if err != nil {
		return 0, err
	}
	xp := step.Reward().XP()
	p.state.XP += xp
	p.state.Step++
	p.state.Completed = p.state.Step == m.Length()
	return xp, nil
}
