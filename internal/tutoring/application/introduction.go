package application

import "context"

// Welcome copy is deliberately separate from bilingual curriculum dialogue and
// exercise instructions. It establishes the role-play without exposing answers.
func opening(teaching TeachingContext) string {
	story, question := adventureStory(teaching)
	return "Hi! I'm Alex, your block-world guide. " + story +
		"\n\nSpeak English; short answers are perfect. I'll ask one question at a time. Say 'help' for a hint, or 'repeat' to hear a question again. You can interrupt me.\n\n" + question
}

func adventureStory(teaching TeachingContext) (string, string) {
	switch teaching.Step.MissionID {
	case "lost-diamond":
		return "A diamond is missing from our cave. Together, we'll follow clues to find it.", "Where should we look first?"
	case "mysterious-village":
		return "Our village has a mystery. We'll explore the library, the tower, and the river to find clues.", "Where would you like to go first?"
	case "post-match":
		return "The match has just finished. You're a player, and I'm the reporter. We'll tell the story of the match.", "How did the match go?"
	default:
		return "We are exploring a new block world. We'll make a story together.", "Where would you like to start?"
	}
}

// Introduction is a local, authenticated read. Previewing the story never starts
// microphone capture, reserves money, calls an LLM, or changes learning evidence.
func (s *Service) Introduction(ctx context.Context, student, mission string) (string, error) {
	teaching, err := s.Context(ctx, student, mission)
	if err != nil {
		return "", err
	}
	return opening(teaching), nil
}
