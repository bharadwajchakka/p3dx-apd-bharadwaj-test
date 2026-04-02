package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cdpg/dx/apd-go/internal/domain"
)

type FormSubmissionRepo struct {
	db *pgxpool.Pool
}

func NewFormSubmissionRepo(db *pgxpool.Pool) *FormSubmissionRepo {
	return &FormSubmissionRepo{db: db}
}

func (r *FormSubmissionRepo) Create(ctx context.Context, s *domain.FormSubmission) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}

	payload, err := json.Marshal(s.Payload)
	if err != nil {
		return fmt.Errorf("marshal form payload: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO form_submissions (id, payload, created_at)
		VALUES ($1, $2, $3)
	`, s.ID, payload, s.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert form submission: %w", err)
	}
	return nil
}

func (r *FormSubmissionRepo) List(ctx context.Context) ([]*domain.FormSubmission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, payload, created_at
		FROM form_submissions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*domain.FormSubmission
	for rows.Next() {
		var sub domain.FormSubmission
		var payloadRaw []byte
		if err := rows.Scan(&sub.ID, &payloadRaw, &sub.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan form submission: %w", err)
		}
		if err := json.Unmarshal(payloadRaw, &sub.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal form payload: %w", err)
		}
		subs = append(subs, &sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}
