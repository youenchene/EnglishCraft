package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	curriculum "github.com/youenchene/EnglishCraft/internal/curriculum/domain"
	practice "github.com/youenchene/EnglishCraft/internal/learning/application/practice"
	learning "github.com/youenchene/EnglishCraft/internal/learning/domain"
	shared "github.com/youenchene/EnglishCraft/internal/shared/domain"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) NextForStudent(ctx context.Context, studentID string) (curriculum.IrregularVerb, error) {
	const query = `
		SELECT v.id, v.base_form, v.past_simple, v.past_participle, v.translation_fr, v.skill_id
		FROM irregular_verbs v
		LEFT JOIN practice_attempts a ON a.verb_id = v.id AND a.student_id = $1
		GROUP BY v.id
		ORDER BY COUNT(a.id), MAX(a.answered_at) NULLS FIRST, v.id
		LIMIT 1`
	return scanVerb(s.pool.QueryRow(ctx, query, studentID))
}

func (s *Store) ByID(ctx context.Context, id string) (curriculum.IrregularVerb, error) {
	const query = `
		SELECT id, base_form, past_simple, past_participle, translation_fr, skill_id
		FROM irregular_verbs WHERE id = $1`
	return scanVerb(s.pool.QueryRow(ctx, query, id))
}

type rowScanner interface{ Scan(dest ...any) error }

func scanVerb(row rowScanner) (curriculum.IrregularVerb, error) {
	var id, base, past, participle, translation, rawSkillID string
	if err := row.Scan(&id, &base, &past, &participle, &translation, &rawSkillID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return curriculum.IrregularVerb{}, practice.ErrVerbNotFound
		}
		return curriculum.IrregularVerb{}, fmt.Errorf("scan irregular verb: %w", err)
	}
	skillID, err := shared.NewSkillID(rawSkillID)
	if err != nil {
		return curriculum.IrregularVerb{}, fmt.Errorf("restore skill ID: %w", err)
	}
	verb, err := curriculum.NewIrregularVerb(id, base, past, participle, translation, skillID)
	if err != nil {
		return curriculum.IrregularVerb{}, fmt.Errorf("restore irregular verb: %w", err)
	}
	return verb, nil
}

func (s *Store) ByStudentAndSkill(ctx context.Context, studentID string, skillID shared.SkillID) (practice.ProgressSnapshot, error) {
	const query = `
		SELECT correct_answers, total_answers, validation_level, next_review_at, version
		FROM skill_progress WHERE student_id = $1 AND skill_id = $2`
	var correct, total int
	var version int64
	var level learning.ValidationLevel
	var nextReviewAt time.Time
	if err := s.pool.QueryRow(ctx, query, studentID, skillID.String()).Scan(&correct, &total, &level, &nextReviewAt, &version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return practice.ProgressSnapshot{}, practice.ErrProgressNotFound
		}
		return practice.ProgressSnapshot{}, fmt.Errorf("scan skill progress: %w", err)
	}
	progress, err := learning.RestoreSkillProgress(studentID, skillID, correct, total, level, nextReviewAt)
	if err != nil {
		return practice.ProgressSnapshot{}, fmt.Errorf("restore skill progress: %w", err)
	}
	return practice.ProgressSnapshot{Progress: progress, Version: version}, nil
}

func (s *Store) Record(ctx context.Context, attempt learning.PracticeAttempt, progress learning.SkillProgress, expectedVersion int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin record transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertAttempt = `
		INSERT INTO practice_attempts
			(id, student_id, verb_id, skill_id, past_simple_answer, past_participle_answer, correct, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.Exec(ctx, insertAttempt,
		attempt.ID(), attempt.StudentID(), attempt.VerbID(), attempt.SkillID().String(),
		attempt.PastSimple(), attempt.PastParticiple(), attempt.Correct(), attempt.AnsweredAt(),
	); err != nil {
		return fmt.Errorf("insert practice attempt: %w", err)
	}

	var commandTag pgconn.CommandTag
	if expectedVersion == 0 {
		const insertProgress = `
		INSERT INTO skill_progress
			(student_id, skill_id, correct_answers, total_answers, validation_level, next_review_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 1)
		ON CONFLICT (student_id, skill_id) DO NOTHING`
		commandTag, err = tx.Exec(ctx, insertProgress,
			progress.StudentID(), progress.SkillID().String(), progress.CorrectAnswers(),
			progress.TotalAnswers(), progress.Level(), progress.NextReviewAt(), attempt.AnsweredAt(),
		)
	} else {
		const updateProgress = `
		UPDATE skill_progress SET
			correct_answers = $3, total_answers = $4, validation_level = $5,
			next_review_at = $6, updated_at = $7, version = version + 1
		WHERE student_id = $1 AND skill_id = $2 AND version = $8`
		commandTag, err = tx.Exec(ctx, updateProgress,
			progress.StudentID(), progress.SkillID().String(), progress.CorrectAnswers(),
			progress.TotalAnswers(), progress.Level(), progress.NextReviewAt(), attempt.AnsweredAt(), expectedVersion,
		)
	}
	if err != nil {
		return fmt.Errorf("upsert skill progress: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return practice.ErrConcurrentUpdate
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit record transaction: %w", err)
	}
	return nil
}
