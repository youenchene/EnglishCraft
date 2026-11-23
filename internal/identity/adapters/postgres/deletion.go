package postgres

import (
	"context"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

func (r *Store) DeleteChild(ctx context.Context, parent, child string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM family_accounts WHERE id=$1 AND parent_id=$2 AND role='child')`, child, parent).Scan(&owned); err != nil {
		return err
	}
	if !owned {
		return identity.ErrInvalidCredentials
	}
	// Also erase the preserved prototype's tables, which predate foreign keys.
	for _, statement := range []string{`DELETE FROM practice_attempts WHERE student_id=$1`, `DELETE FROM skill_progress WHERE student_id=$1`, `DELETE FROM family_accounts WHERE id=$1`} {
		if _, err := tx.Exec(ctx, statement, child); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
