package postgres

import (
	"context"
	"encoding/json"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
)

func (s *Store) PutImport(ctx context.Context, student string, value tutor.SchoolImport) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO school_imports(id,student_id,document) VALUES($1,$2,$3)`, value.ID, student, raw)
	return err
}
func (s *Store) Imports(ctx context.Context, student string, approvedOnly bool) ([]tutor.SchoolImport, error) {
	rows, err := s.pool.Query(ctx, `SELECT document,approved FROM school_imports WHERE student_id=$1 AND (NOT $2 OR approved) ORDER BY id`, student, approvedOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []tutor.SchoolImport{}
	for rows.Next() {
		var value tutor.SchoolImport
		var raw []byte
		var approved bool
		if err := rows.Scan(&raw, &approved); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		value.Approved = approved
		result = append(result, value)
	}
	return result, rows.Err()
}
func (s *Store) ApproveImport(ctx context.Context, student, id string) error {
	result, err := s.pool.Exec(ctx, `UPDATE school_imports SET approved=true WHERE student_id=$1 AND id=$2`, student, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return tutor.ErrInvalid
	}
	return nil
}
