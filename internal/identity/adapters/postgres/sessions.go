package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) Login(ctx context.Context, username string, check func(*identity.Account) (identity.Session, error)) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin login: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var account identity.Account
	var locked *time.Time
	err = tx.QueryRow(ctx, `SELECT id,username,password_hash,role,failed_logins,locked_until
		FROM family_accounts WHERE username=$1 FOR UPDATE`, username).
		Scan(&account.ID, &account.Username, &account.PasswordHash, &account.Role, &account.Failed, &locked)
	if errors.Is(err, pgx.ErrNoRows) {
		// Synthetic bcrypt digest (cost 12), never an account credential.
		// Run the same verification path to avoid a cheap username oracle.
		_, _ = check(&identity.Account{PasswordHash: "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"})
		return identity.ErrInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("find login account: %w", err)
	}
	if locked != nil {
		account.LockedUntil = *locked
	}
	session, authErr := check(&account)
	if _, err := tx.Exec(ctx, `UPDATE family_accounts SET failed_logins=$2,locked_until=$3 WHERE id=$1`,
		account.ID, account.Failed, account.LockedUntil); err != nil {
		return fmt.Errorf("persist login limits: %w", err)
	}
	if authErr == nil {
		if _, err := tx.Exec(ctx, `INSERT INTO auth_sessions(token_hash,account_id,expires_at) VALUES($1,$2,$3)`,
			session.Hash, session.AccountID, session.Expires); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit login: %w", err)
	}
	return authErr
}
func (s *Store) Resolve(ctx context.Context, hash string, now time.Time) (identity.Principal, error) {
	var p identity.Principal
	err := s.pool.QueryRow(ctx, `SELECT a.id,a.role FROM auth_sessions s JOIN family_accounts a ON a.id=s.account_id
		WHERE s.token_hash=$1 AND s.expires_at>$2`, hash, now).Scan(&p.SubjectID, &p.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, identity.ErrInvalidCredentials
	}
	return p, err
}
func (s *Store) Revoke(ctx context.Context, hash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash=$1`, hash)
	return err
}
func (s *Store) ChildForParent(ctx context.Context, parent string) (string, error) {
	var child string
	err := s.pool.QueryRow(ctx, `SELECT id FROM family_accounts WHERE parent_id=$1 AND role='child' ORDER BY id LIMIT 1`, parent).Scan(&child)
	return child, err
}
func (s *Store) Provision(ctx context.Context, f identity.Family) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `INSERT INTO family_accounts(id,username,password_hash,role) VALUES($1,$2,$3,'parent')`,
		f.Parent.ID, f.Parent.Username, f.Parent.PasswordHash); err != nil {
		return fmt.Errorf("provision parent (already provisioned?): %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO family_accounts(id,username,password_hash,role,parent_id) VALUES($1,$2,$3,'child',$4)`,
		f.Child.ID, f.Child.Username, f.Child.PasswordHash, f.Parent.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO student_states(student_id,state) VALUES($1,'{"missions":{},"reviews":{},"attempts":[],"receipts":{}}')`,
		f.Child.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
