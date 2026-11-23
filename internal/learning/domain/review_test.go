package domain

import (
	"testing"
	"time"
)

func TestReviewAcrossSimulatedDaysAndFailure(t *testing.T) {
	r := NewReviewSchedule()
	day := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if r.Snapshot().Stage != "discovery" {
		t.Fatal("expected discovery")
	}
	for _, tc := range []struct {
		day, interval int
		stage         string
	}{
		{day: 0, interval: 1, stage: "learning"},
		{day: 1, interval: 3, stage: "learning"},
		{day: 4, interval: 7, stage: "mastery"},
		{day: 11, interval: 14, stage: "mastery"},
		{day: 25, interval: 30, stage: "mastery"},
	} {
		now := day.Add(time.Duration(tc.day) * 24 * time.Hour)
		if err := r.Record(true, now); err != nil {
			t.Fatal(err)
		}
		if got := r.Snapshot(); got.Stage != tc.stage || !got.Due.Equal(now.Add(time.Duration(tc.interval)*24*time.Hour)) {
			t.Fatalf("day %d unexpected state %+v", tc.day, got)
		}
	}
	now := day.Add(26 * 24 * time.Hour)
	if err := r.Record(false, now); err != nil {
		t.Fatal(err)
	}
	if got := r.Snapshot(); got.Stage != "learning" || got.Streak != 0 || got.Failures != 1 || !got.Due.Equal(now.Add(4*time.Hour)) {
		t.Fatalf("failure must reset mastery: %+v", got)
	}
}
func TestEarlySuccessCannotFarmMasteryOrPushDueDate(t *testing.T) {
	r := NewReviewSchedule()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for range 20 {
		if err := r.Record(true, now); err != nil {
			t.Fatal(err)
		}
	}
	if got := r.Snapshot(); got.Streak != 1 || got.Attempts != 20 || got.Stage != "learning" || !got.Due.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("early repetitions advanced mastery: %+v", got)
	}
	if err := r.Record(true, now.Add(-time.Second)); err == nil {
		t.Fatal("time travel accepted")
	}
	if _, err := RestoreReview(ReviewSnapshot{Stage: "mastery", Attempts: 1, Failures: 2}); err == nil {
		t.Fatal("invalid evidence accepted")
	}
}
