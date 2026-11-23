package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

func (s *Store) Children(ctx context.Context, parent string) ([]identity.ChildProfile, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,username FROM family_accounts WHERE parent_id=$1 AND role='child' ORDER BY username,id`, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	children := []identity.ChildProfile{}
	for rows.Next() {
		var child identity.ChildProfile
		if err := rows.Scan(&child.ID, &child.Username); err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, rows.Err()
}
func (s *Store) AddChild(ctx context.Context, parent string, child identity.Account) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var role string
	if err := tx.QueryRow(ctx, `SELECT role FROM family_accounts WHERE id=$1 FOR SHARE`, parent).Scan(&role); err != nil {
		return err
	}
	if role != "parent" {
		return identity.ErrInvalidCredentials
	}
	if _, err := tx.Exec(ctx, `INSERT INTO family_accounts(id,username,password_hash,role,parent_id) VALUES($1,$2,$3,'child',$4)`, child.ID, child.Username, child.PasswordHash, parent); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return identity.ErrUsernameTaken
		}
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO student_states(student_id,state) VALUES($1,'{"missions":{},"reviews":{},"attempts":[],"receipts":{}}')`, child.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
