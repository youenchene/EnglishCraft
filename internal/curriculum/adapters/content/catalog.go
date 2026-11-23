package content

import (
	"encoding/json"
	"fmt"
	"strings"

	files "github.com/youenchene/EnglishCraft/content"
	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	missions "github.com/youenchene/EnglishCraft/internal/missions/domain"
	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

// Raw content DTOs are validated and rebuilt into immutable domain values.
type VerbRecord = core.Verb
type Catalog struct {
	verbs     []VerbRecord
	missions  []missions.Mission
	exercises map[string]learning.Exercise
	skills    []curriculum.Skill
}

func Load() (*Catalog, error) {
	c := &Catalog{verbs: []VerbRecord{}, missions: []missions.Mission{},
		exercises: map[string]learning.Exercise{}, skills: []curriculum.Skill{}}
	var library struct {
		Version int
		Verbs   []VerbRecord
	}
	if err := read("curriculum/verbs.v1.json", &library); err != nil {
		return nil, err
	}
	if library.Version != 1 || len(library.Verbs) < 30 {
		return nil, fmt.Errorf("invalid verb library")
	}
	seen := map[string]bool{}
	for _, v := range library.Verbs {
		if seen[v.Base] || len(v.Examples) < 2 || len(v.Categories) == 0 || v.Difficulty < 1 || v.Difficulty > 3 {
			return nil, fmt.Errorf("invalid teaching metadata for %q", v.Base)
		}
		seen[v.Base] = true
		skill, err := shared.NewSkillID("english.irregular." + v.Base)
		if err != nil {
			return nil, err
		}
		verb, err := curriculum.NewIrregularVerb("verb-"+v.Base, v.Base, v.Past, v.Participle, v.Translation, skill)
		if err != nil {
			return nil, fmt.Errorf("validate verb: %w", err)
		}
		if _, err := curriculum.NewTeachingIrregularVerb(verb, curriculum.TeachingMetadata{Examples: v.Examples,
			Difficulty: v.Difficulty, Categories: v.Categories}); err != nil {
			return nil, err
		}
		c.verbs = append(c.verbs, v)
		for _, kind := range []learning.ExerciseKind{learning.PastSimple, learning.PastParticiple, learning.BuildSentence} {
			accepted := strings.Split(v.Past, "/")
			prompt := "Donne le prétérit de " + v.Base + "."
			if kind == learning.PastParticiple {
				accepted = strings.Split(v.Participle, "/")
				prompt = "Donne le participe passé de " + v.Base + "."
			}
			if kind == learning.BuildSentence {
				accepted = []string{}
				prompt = "Écris une phrase au passé avec " + v.Base + ". (Évaluation humaine nécessaire)"
			}
			e, err := learning.NewExercise(learning.ExerciseSpec{ID: "train-" + v.Base + "-" + string(kind), Kind: kind,
				Prompt: prompt, Skill: skill.String(), Verb: v.Base, Accepted: accepted})
			if err != nil {
				return nil, err
			}
			c.exercises[e.ID()] = e
		}
	}
	var adventures struct {
		Version  int
		Missions []struct {
			ID, Title, Theme string
			Steps            []struct {
				Dialogue string
				XP       int
				Exercise learning.ExerciseSpec
			}
		}
	}
	if err := read("missions/missions.v1.json", &adventures); err != nil {
		return nil, err
	}
	for _, raw := range adventures.Missions {
		steps := []missions.MissionStep{}
		for _, step := range raw.Steps {
			e, err := learning.NewExercise(step.Exercise)
			if err != nil {
				return nil, err
			}
			if _, exists := c.exercises[e.ID()]; exists {
				return nil, fmt.Errorf("duplicate exercise")
			}
			c.exercises[e.ID()] = e
			challenge, err := missions.NewChallenge(e.ID())
			if err != nil {
				return nil, err
			}
			reward, err := missions.NewReward(step.XP)
			if err != nil {
				return nil, err
			}
			s, err := missions.NewMissionStep(step.Dialogue, challenge, reward)
			if err != nil {
				return nil, err
			}
			steps = append(steps, s)
		}
		m, err := missions.NewMission(raw.ID, raw.Title, raw.Theme, adventures.Version, steps)
		if err != nil {
			return nil, err
		}
		for _, previous := range c.missions {
			if previous.ID() == m.ID() {
				return nil, fmt.Errorf("duplicate mission")
			}
		}
		c.missions = append(c.missions, m)
	}
	skillIDs := map[string]bool{}
	for _, e := range c.exercises {
		skillIDs[e.Skill()] = true
	}
	for id := range skillIDs {
		skill, err := curriculum.NewSkill(id, strings.TrimPrefix(id, "english."))
		if err != nil {
			return nil, err
		}
		c.skills = append(c.skills, skill)
	}
	return c, nil
}
func read(path string, target any) error {
	data, err := files.Files.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read content: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode content: %w", err)
	}
	return nil
}
func (c *Catalog) Missions() []missions.Mission { return append([]missions.Mission{}, c.missions...) }
func (c *Catalog) Skills() []curriculum.Skill   { return append([]curriculum.Skill{}, c.skills...) }
func (c *Catalog) Verbs() []VerbRecord          { return append([]VerbRecord{}, c.verbs...) }
func (c *Catalog) Mission(id string) (missions.Mission, bool) {
	for _, m := range c.missions {
		if m.ID() == id {
			return m, true
		}
	}
	return missions.Mission{}, false
}
func (c *Catalog) Exercise(id string) (learning.Exercise, bool) {
	e, ok := c.exercises[id]
	return e, ok
}
