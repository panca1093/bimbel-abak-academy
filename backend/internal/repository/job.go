package repository

import (
	"context"

	"akademi-bimbel/internal/model"
)

const jobColumns = `id, type, status, progress, input_url, result_url, error, created_by, created_at, updated_at`

func scanJob(row interface {
	Scan(dest ...any) error
}, j *model.Job) error {
	return row.Scan(
		&j.ID, &j.Type, &j.Status, &j.Progress, &j.InputURL, &j.ResultURL,
		&j.Error, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt,
	)
}

func (r *Repository) CreateJob(ctx context.Context, j *model.Job) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO job (type, input_url, created_by) VALUES ($1, $2, $3)
		 RETURNING id, status, progress, created_at, updated_at`,
		j.Type, j.InputURL, j.CreatedBy,
	).Scan(&j.ID, &j.Status, &j.Progress, &j.CreatedAt, &j.UpdatedAt)
}

func (r *Repository) GetJobByID(ctx context.Context, id string) (*model.Job, error) {
	j := &model.Job{}
	err := scanJob(r.pool.QueryRow(ctx,
		`SELECT `+jobColumns+` FROM job WHERE id = $1`,
		id,
	), j)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

// ClaimNextJob atomically claims the oldest queued job in a single
// UPDATE...RETURNING statement so concurrent pollers can never claim the
// same row (FOR UPDATE SKIP LOCKED on the inner SELECT).
func (r *Repository) ClaimNextJob(ctx context.Context) (*model.Job, error) {
	j := &model.Job{}
	err := scanJob(r.pool.QueryRow(ctx,
		`UPDATE job SET status = 'running', updated_at = now()
		 WHERE id = (SELECT id FROM job WHERE status = 'queued' ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		 RETURNING `+jobColumns,
	), j)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

func (r *Repository) UpdateJobProgress(ctx context.Context, id string, progress int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE job SET progress = $1, updated_at = now() WHERE id = $2`,
		progress, id,
	)
	return err
}

func (r *Repository) FinishJob(ctx context.Context, id, status string, progress int, resultURL, errMsg *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE job SET status = $1, progress = $2, result_url = $3, error = $4, updated_at = now() WHERE id = $5`,
		status, progress, resultURL, errMsg, id,
	)
	return err
}
