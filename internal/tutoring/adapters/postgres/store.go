package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"time"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) Enabled(ctx context.Context, student string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT COALESCE((SELECT enabled FROM ai_preferences WHERE student_id=$1),false)`, student).Scan(&enabled)
	return enabled, err
}
func (s *Store) SetEnabled(ctx context.Context, student string, enabled bool) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO ai_preferences(student_id,enabled) VALUES($1,$2) ON CONFLICT(student_id) DO UPDATE SET enabled=excluded.enabled`, student, enabled)
	return err
}
func (s *Store) Reserve(ctx context.Context, a tutor.Allowance) (string, error) {
	if a.ReserveMicros <= 0 || a.DailyMicros <= 0 || a.StudentDailyMicros <= 0 {
		return "", tutor.ErrBudget
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Global lock is intentionally simple and safe for a small family installation.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(726343813)`); err != nil {
		return "", err
	}
	var enabled bool
	if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT enabled FROM ai_preferences WHERE student_id=$1),false)`, a.Student).Scan(&enabled); err != nil {
		return "", err
	}
	if !enabled {
		return "", tutor.ErrBudget
	}
	day := a.At.UTC().Format("2006-01-02")
	var provider, student int64
	var seconds int
	err = tx.QueryRow(ctx, `SELECT COALESCE(sum(reserved_micros) FILTER(WHERE provider=$2),0),
 COALESCE(sum(reserved_micros) FILTER(WHERE student_id=$3),0),
 COALESCE(sum(voice_seconds) FILTER(WHERE student_id=$3),0) FROM ai_usage WHERE day=$1`, day, a.Provider, a.Student).Scan(&provider, &student, &seconds)
	if err != nil {
		return "", err
	}
	if a.ReserveMicros > a.DailyMicros-provider || a.ReserveMicros > a.StudentDailyMicros-student {
		return "", tutor.ErrBudget
	}
	if a.VoiceSeconds > 0 && a.VoiceSeconds > a.DailyVoiceSeconds-seconds {
		return "", tutor.ErrBudget
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	id := hex.EncodeToString(raw)
	_, err = tx.Exec(ctx, `INSERT INTO ai_usage(id,student_id,provider,day,reserved_micros,voice_seconds) VALUES($1,$2,$3,$4,$5,$6)`,
		id, a.Student, a.Provider, day, a.ReserveMicros, a.VoiceSeconds)
	if err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}
func (s *Store) Record(ctx context.Context, id string, u tutor.Usage, status string) error {
	if u.InputTokens < 0 || u.OutputTokens < 0 {
		return tutor.ErrInvalid
	}
	result, err := s.pool.Exec(ctx, `UPDATE ai_usage SET input_tokens=$2,output_tokens=$3,status=$4 WHERE id=$1`, id, u.InputTokens, u.OutputTokens, status)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return tutor.ErrInvalid
	}
	return nil
}
func (s *Store) Spending(ctx context.Context, student string, at time.Time) ([]tutor.Spending, error) {
	rows, err := s.pool.Query(ctx, `SELECT provider,sum(reserved_micros),count(*),sum(voice_seconds) FROM ai_usage WHERE student_id=$1 AND day=$2 GROUP BY provider ORDER BY provider`, student, at.UTC().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []tutor.Spending{}
	for rows.Next() {
		var value tutor.Spending
		if err := rows.Scan(&value.Provider, &value.ReservedMicros, &value.Calls, &value.VoiceSeconds); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
func (s *Store) Save(ctx context.Context, r tutor.VoiceRecord) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO voice_sessions(id,student_id,mission_id,expires_at) VALUES($1,$2,$3,$4)`, r.Session.ID, r.Student, r.Mission, r.Session.Expires)
	return err
}
func (s *Store) Find(ctx context.Context, student, id string) (tutor.VoiceRecord, error) {
	var r tutor.VoiceRecord
	err := s.pool.QueryRow(ctx, `SELECT id,student_id,mission_id,expires_at FROM voice_sessions WHERE id=$1 AND student_id=$2`, id, student).Scan(&r.Session.ID, &r.Student, &r.Mission, &r.Session.Expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, tutor.ErrInvalid
	}
	return r, err
}
func (s *Store) Lookup(ctx context.Context, id string) (tutor.VoiceRecord, error) {
	var r tutor.VoiceRecord
	err := s.pool.QueryRow(ctx, `SELECT id,student_id,mission_id,expires_at FROM voice_sessions WHERE id=$1`, id).Scan(&r.Session.ID, &r.Student, &r.Mission, &r.Session.Expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, tutor.ErrInvalid
	}
	return r, err
}
func (s *Store) Expired(ctx context.Context, at time.Time) ([]tutor.VoiceRecord, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,student_id,mission_id,expires_at FROM voice_sessions WHERE expires_at<=$1`, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []tutor.VoiceRecord{}
	for rows.Next() {
		var r tutor.VoiceRecord
		if err := rows.Scan(&r.Session.ID, &r.Student, &r.Mission, &r.Session.Expires); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func (s *Store) Remove(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM voice_sessions WHERE id=$1`, id)
	return err
}
