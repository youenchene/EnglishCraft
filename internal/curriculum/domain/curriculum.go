package domain

import (
	"errors"
	"strings"
)

type Skill struct{ id, title string }

func NewSkill(id, title string) (Skill, error) {
	if !strings.HasPrefix(id, "english.") || title == "" {
		return Skill{}, errors.New("invalid skill")
	}
	return Skill{id: id, title: title}, nil
}
func (s Skill) ID() string    { return s.id }
func (s Skill) Title() string { return s.title }

type Prerequisite struct{ skill Skill }

func NewPrerequisite(skill Skill) (Prerequisite, error) {
	if skill.ID() == "" {
		return Prerequisite{}, errors.New("invalid prerequisite")
	}
	return Prerequisite{skill: skill}, nil
}
func (p Prerequisite) Skill() Skill { return p.skill }

type LearningObjective struct {
	skill       Skill
	description string
}

func NewLearningObjective(skill Skill, description string) (LearningObjective, error) {
	if skill.ID() == "" || description == "" {
		return LearningObjective{}, errors.New("invalid objective")
	}
	return LearningObjective{skill: skill, description: description}, nil
}
func (o LearningObjective) Skill() Skill        { return o.skill }
func (o LearningObjective) Description() string { return o.description }

type CurriculumModule struct {
	id            string
	objectives    []LearningObjective
	prerequisites []Prerequisite
}

func NewCurriculumModule(id string, objectives []LearningObjective, prerequisites []Prerequisite) (CurriculumModule, error) {
	if id == "" || len(objectives) == 0 {
		return CurriculumModule{}, errors.New("invalid curriculum module")
	}
	for _, o := range objectives {
		if o.Skill().ID() == "" {
			return CurriculumModule{}, errors.New("invalid module objective")
		}
	}
	return CurriculumModule{id: id, objectives: append([]LearningObjective{}, objectives...),
		prerequisites: append([]Prerequisite{}, prerequisites...)}, nil
}
func (m CurriculumModule) ID() string { return m.id }
func (m CurriculumModule) Objectives() []LearningObjective {
	return append([]LearningObjective{}, m.objectives...)
}
func (m CurriculumModule) Prerequisites() []Prerequisite {
	return append([]Prerequisite{}, m.prerequisites...)
}
