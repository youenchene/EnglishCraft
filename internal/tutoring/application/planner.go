package application

import (
	"context"
	"sort"
)

type PlanStage struct {
	Name    string   `json:"name"`
	Minutes int      `json:"minutes"`
	Skills  []string `json:"skills"`
	Reason  string   `json:"reason"`
}
type Plan struct {
	Minutes int         `json:"minutes"`
	Stages  []PlanStage `json:"stages"`
}

func (s *Service) Plan(ctx context.Context, student string) (Plan, error) {
	summary, err := s.learning.Progress(ctx, student)
	if err != nil {
		return Plan{}, err
	}
	review := append([]string{}, summary.Due...)
	if len(review) > 3 {
		review = review[:3]
	}
	consolidation := []string{}
	for skill, count := range summary.Errors {
		if count >= 2 && summary.Reviews[skill].Stage != "mastery" {
			consolidation = append(consolidation, skill)
		}
	}
	sort.Strings(consolidation)
	reason := "Consolidation : au moins deux erreurs observées."
	if len(consolidation) == 0 {
		reason = "Découverte : verbe du catalogue scolaire non encore étudié."
		for _, verb := range s.learning.Verbs() {
			skill := "english.irregular." + verb.Base
			if _, seen := summary.Reviews[skill]; !seen {
				consolidation = append(consolidation, skill)
				break
			}
		}
	}
	if len(consolidation) > 2 {
		consolidation = consolidation[:2]
	}
	return Plan{Minutes: 18, Stages: []PlanStage{
		{Name: "Révisions", Minutes: 4, Skills: review, Reason: "Échéances atteintes ; échecs prioritaires."},
		{Name: "Découverte ou consolidation", Minutes: 4, Skills: consolidation, Reason: reason},
		{Name: "Aventure ludique", Minutes: 5, Skills: consolidation, Reason: "Réutiliser les notions dans une mission."},
		{Name: "Évaluation courte", Minutes: 3, Skills: consolidation, Reason: "Correction déterministe, sans maîtrise décidée par l'IA."},
		{Name: "Bilan", Minutes: 2, Skills: []string{}, Reason: "Relire les corrections et les prochaines échéances."},
	}}, nil
}
