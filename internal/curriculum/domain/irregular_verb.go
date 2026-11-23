package domain

import (
	"errors"
	"strings"

	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

var ErrInvalidIrregularVerb = errors.New("invalid irregular verb")

type IrregularVerb struct {
	id             string
	baseForm       string
	pastSimple     string
	pastParticiple string
	translation    string
	skillID        shared.SkillID
	metadata       TeachingMetadata
}

type TeachingMetadata struct {
	Examples   []string
	Difficulty int
	Categories []string
}

func NewTeachingIrregularVerb(verb IrregularVerb, metadata TeachingMetadata) (IrregularVerb, error) {
	validMetadata := len(metadata.Examples) >= 2 && metadata.Difficulty >= 1 && metadata.Difficulty <= 3 && len(metadata.Categories) > 0
	if verb.ID() == "" || !validMetadata {
		return IrregularVerb{}, ErrInvalidIrregularVerb
	}
	for _, example := range metadata.Examples {
		if strings.TrimSpace(example) == "" {
			return IrregularVerb{}, ErrInvalidIrregularVerb
		}
	}
	for _, category := range metadata.Categories {
		if strings.TrimSpace(category) == "" {
			return IrregularVerb{}, ErrInvalidIrregularVerb
		}
	}
	verb.metadata = TeachingMetadata{Examples: append([]string{}, metadata.Examples...), Difficulty: metadata.Difficulty,
		Categories: append([]string{}, metadata.Categories...)}
	return verb, nil
}
func (v IrregularVerb) Metadata() TeachingMetadata {
	return TeachingMetadata{Examples: append([]string{}, v.metadata.Examples...), Difficulty: v.metadata.Difficulty,
		Categories: append([]string{}, v.metadata.Categories...)}
}

func NewIrregularVerb(id, baseForm, pastSimple, pastParticiple, translation string, skillID shared.SkillID) (IrregularVerb, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(baseForm) == "" ||
		strings.TrimSpace(pastSimple) == "" || strings.TrimSpace(pastParticiple) == "" ||
		strings.TrimSpace(translation) == "" || skillID == "" {
		return IrregularVerb{}, ErrInvalidIrregularVerb
	}
	return IrregularVerb{
		id:             id,
		baseForm:       normalize(baseForm),
		pastSimple:     normalize(pastSimple),
		pastParticiple: normalize(pastParticiple),
		translation:    strings.TrimSpace(translation),
		skillID:        skillID,
	}, nil
}

func (v IrregularVerb) ID() string              { return v.id }
func (v IrregularVerb) BaseForm() string        { return v.baseForm }
func (v IrregularVerb) PastSimple() string      { return v.pastSimple }
func (v IrregularVerb) PastParticiple() string  { return v.pastParticiple }
func (v IrregularVerb) Translation() string     { return v.translation }
func (v IrregularVerb) SkillID() shared.SkillID { return v.skillID }

func (v IrregularVerb) Check(pastSimple, pastParticiple string) VerbCorrection {
	pastSimpleCorrect := matchesVariant(pastSimple, v.pastSimple)
	pastParticipleCorrect := matchesVariant(pastParticiple, v.pastParticiple)

	return VerbCorrection{
		PastSimpleCorrect:     pastSimpleCorrect,
		PastParticipleCorrect: pastParticipleCorrect,
		ExpectedPastSimple:    v.pastSimple,
		ExpectedParticiple:    v.pastParticiple,
	}
}

type VerbCorrection struct {
	PastSimpleCorrect     bool
	PastParticipleCorrect bool
	ExpectedPastSimple    string
	ExpectedParticiple    string
}

func (c VerbCorrection) Correct() bool {
	return c.PastSimpleCorrect && c.PastParticipleCorrect
}

func normalize(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func matchesVariant(answer, forms string) bool {
	for _, form := range strings.Split(forms, "/") {
		if normalize(answer) == normalize(form) {
			return true
		}
	}
	return false
}
