package application

import "context"

type StudentDeletion interface {
	DeleteChild(context.Context, string, string) error
}
type Deletion struct {
	sessions *Sessions
	repo     StudentDeletion
}

func NewDeletion(sessions *Sessions, repo StudentDeletion) *Deletion {
	return &Deletion{sessions: sessions, repo: repo}
}
func (s *Deletion) Delete(ctx context.Context, parent Principal) error {
	return s.DeleteSelected(ctx, parent, "")
}
func (s *Deletion) DeleteSelected(ctx context.Context, parent Principal, selected string) error {
	if parent.Role != RoleParent {
		return ErrInvalidCredentials
	}
	child, err := s.sessions.SelectChild(ctx, parent, selected)
	if err != nil {
		return err
	}
	return s.repo.DeleteChild(ctx, parent.SubjectID, child)
}
