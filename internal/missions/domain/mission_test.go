package domain

import "testing"

func TestMissionRewardAndCompletionInvariants(t *testing.T) {
	challenge, err := NewChallenge("exercise")
	if err != nil {
		t.Fatal(err)
	}
	reward, err := NewReward(10)
	if err != nil {
		t.Fatal(err)
	}
	step, err := NewMissionStep("Alex", challenge, reward)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMission("m", "Mission", "forest", 1, []MissionStep{step})
	if err != nil {
		t.Fatal(err)
	}
	p := Start(m)
	if xp, err := p.CompleteStep(m, 0, false); err != nil || xp != 0 || p.Snapshot().Completed {
		t.Fatal("incorrect answer rewarded")
	}
	if xp, err := p.CompleteStep(m, 0, true); err != nil || xp != 10 || !p.Snapshot().Completed {
		t.Fatal("correct answer not rewarded")
	}
	if _, err := p.CompleteStep(m, 0, true); err == nil {
		t.Fatal("double reward accepted")
	}
	if _, err := NewReward(10000); err == nil {
		t.Fatal("invalid XP")
	}
	if _, err := NewMission("m", "Mission", "forest", 1, nil); err == nil {
		t.Fatal("empty mission")
	}
	if _, err := RestoreProgress(m, ProgressSnapshot{Version: 1, Step: 1}); err == nil {
		t.Fatal("inconsistent completion")
	}
}
