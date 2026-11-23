package domain

import (
	"errors"
	"time"
)

// ReviewSnapshot is a persistence value, not a mutable entity.
type ReviewSnapshot struct {
	Stage    string    `json:"stage"`
	Streak   int       `json:"streak"`
	Attempts int       `json:"attempts"`
	Failures int       `json:"failures"`
	Due      time.Time `json:"due"`
	Last     time.Time `json:"last"`
}

type ReviewSchedule struct{ state ReviewSnapshot }

func NewReviewSchedule() ReviewSchedule {
	return ReviewSchedule{state: ReviewSnapshot{Stage: "discovery"}}
}
func RestoreReview(state ReviewSnapshot) (ReviewSchedule, error) {
	validStage := state.Stage == "discovery" || state.Stage == "learning" || state.Stage == "mastery"
	validCounts := state.Streak >= 0 && state.Attempts >= 0 && state.Failures >= 0 && state.Failures <= state.Attempts && state.Streak <= state.Attempts-state.Failures
	validEvidence := (state.Attempts == 0 && state.Stage == "discovery") ||
		(state.Attempts > 0 && state.Stage != "discovery" && !state.Last.IsZero() && state.Due.After(state.Last))
	validMastery := (state.Stage == "mastery") == (state.Streak >= 3)
	if !validStage || !validCounts || !validEvidence || !validMastery {
		return ReviewSchedule{}, errors.New("invalid review schedule")
	}
	return ReviewSchedule{state: state}, nil
}
func (r ReviewSchedule) Snapshot() ReviewSnapshot { return r.state }

// Record always records evidence, but only spaced successes advance mastery.
// Failure resets the streak and schedules four hours later. A success uses
// 1, 3, 7, 14, 30 days. This deliberately transparent V1 can be replaced by FSRS.
func (r *ReviewSchedule) Record(correct bool, now time.Time) error {
	if now.IsZero() || now.Before(r.state.Last) {
		return errors.New("invalid review time")
	}
	r.state.Attempts++
	if !correct {
		r.state.Failures++
		r.state.Streak = 0
		r.state.Stage = "learning"
		r.state.Due = now.Add(4 * time.Hour)
		r.state.Last = now
		return nil
	}
	spaced := r.state.Last.IsZero() || !now.Before(r.state.Due)
	if spaced {
		r.state.Streak++
		intervals := []int{1, 3, 7, 14, 30}
		index := min(r.state.Streak-1, len(intervals)-1)
		r.state.Due = now.Add(time.Duration(intervals[index]) * 24 * time.Hour)
	}
	r.state.Stage = "learning"
	if r.state.Streak >= 3 {
		r.state.Stage = "mastery"
	}
	r.state.Last = now
	return nil
}
