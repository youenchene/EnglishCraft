package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
)

func (s *Store) Read(ctx context.Context, student string) (core.StudentState, error) {
	var data []byte
	if err := s.pool.QueryRow(ctx, `SELECT state FROM student_states WHERE student_id=$1`, student).Scan(&data); err != nil {
		return core.StudentState{}, fmt.Errorf("read student state: %w", err)
	}
	return decodeState(data)
}
func decodeState(data []byte) (core.StudentState, error) {
	state := core.NewState()
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("decode student state: %w", err)
	}
	if state.Missions == nil || state.Reviews == nil || state.Receipts == nil {
		return state, fmt.Errorf("invalid student state")
	}
	return state, nil
}

func (s *Store) Transact(ctx context.Context, student string, mutate func(*core.StudentState) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin student transaction: %w", err)
	}
	// Rollback is harmless after a successful commit and releases locks on errors.
	defer func() { _ = tx.Rollback(ctx) }()
	var data []byte
	if err := tx.QueryRow(ctx, `SELECT state FROM student_states WHERE student_id=$1 FOR UPDATE`, student).Scan(&data); err != nil {
		return fmt.Errorf("lock student state: %w", err)
	}
	state, err := decodeState(data)
	if err != nil {
		return err
	}
	if err := mutate(&state); err != nil {
		return err
	}
	data, err = json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode student state: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE student_states SET state=$2, updated_at=now() WHERE student_id=$1`, student, data); err != nil {
		return fmt.Errorf("persist student state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit student transaction: %w", err)
	}
	return nil
}
