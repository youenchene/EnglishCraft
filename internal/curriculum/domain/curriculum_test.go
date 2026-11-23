package domain

import "testing"

func TestCurriculumInvariantsAndDefensiveCopies(t *testing.T) {
	if _, err := NewSkill("bogus", "x"); err == nil {
		t.Fatal("invalid skill ID")
	}
	s, err := NewSkill("english.past", "Past")
	if err != nil {
		t.Fatal(err)
	}
	o, err := NewLearningObjective(s, "Use past tense")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewPrerequisite(s)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewCurriculumModule("past", []LearningObjective{o}, []Prerequisite{p})
	if err != nil {
		t.Fatal(err)
	}
	copy := m.Objectives()
	copy[0] = LearningObjective{}
	if m.Objectives()[0].Skill().ID() == "" {
		t.Fatal("module mutated through returned slice")
	}
	if _, err := NewCurriculumModule("past", nil, nil); err == nil {
		t.Fatal("empty objectives")
	}
}
