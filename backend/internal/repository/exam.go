package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"akademi-bimbel/internal/model"
)

// ErrSortOrderConflict — uq_question_order SQLSTATE 23505 — surfaced for service-layer mapping.
var ErrSortOrderConflict = errors.New("sort order conflict")

func scanTest(row interface{ Scan(dest ...any) error }, t *model.Test) error {
	var audioURL *string
	var audioPlayLimit *int
	err := row.Scan(
		&t.ID, &t.Title, &t.Subject, &t.Topic, &t.DurationMinutes,
		&audioURL, &audioPlayLimit, &t.CreatedAt,
	)
	if err != nil {
		return err
	}
	if audioURL != nil {
		t.AudioURL = audioURL
	}
	if audioPlayLimit != nil {
		t.AudioPlayLimit = audioPlayLimit
	}
	return nil
}

func scanQuestion(row interface{ Scan(dest ...any) error }, q *model.Question) error {
	var correctAnswer, explanation, difficulty, imageURL *string
	err := row.Scan(
		&q.ID, &q.TestID, &q.Format, &q.Body,
		&correctAnswer, &explanation, &difficulty, &imageURL,
		&q.SortOrder,
	)
	if err != nil {
		return err
	}
	if correctAnswer != nil {
		q.CorrectAnswer = correctAnswer
	}
	if explanation != nil {
		q.Explanation = explanation
	}
	if difficulty != nil {
		q.Difficulty = difficulty
	}
	if imageURL != nil {
		q.ImageURL = imageURL
	}
	return nil
}

func scanQuestionOption(row interface{ Scan(dest ...any) error }, o *model.QuestionOption) error {
	var imageURL *string
	var isCorrect bool
	err := row.Scan(
		&o.QuestionID, &o.Key, &o.Text, &imageURL, &isCorrect, &o.SortOrder,
	)
	if err != nil {
		return err
	}
	if imageURL != nil {
		o.ImageURL = imageURL
	}
	o.IsCorrect = isCorrect
	return nil
}

// TestFilter mirrors ProductFilter shape.
type TestFilter struct {
	Subject string
	Topic   string
	Cursor  string
	Limit   int
}

func (r *Repository) CreateTest(ctx context.Context, t *model.Test) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO test (title, subject, topic, duration_minutes, audio_url, audio_play_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		t.Title, t.Subject, t.Topic, t.DurationMinutes, t.AudioURL, t.AudioPlayLimit,
	).Scan(&t.ID, &t.CreatedAt)
	return err
}

func (r *Repository) GetTestByID(ctx context.Context, id uuid.UUID) (*model.Test, error) {
	out := &model.Test{}
	err := scanTest(r.pool.QueryRow(ctx,
		`SELECT id, title, subject, topic, duration_minutes, audio_url, audio_play_limit, created_at
		FROM test
		WHERE id = $1`,
		id,
	), out)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetTestDetail(ctx context.Context, id uuid.UUID) (*model.TestDetail, error) {
	test, err := r.GetTestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	questions, err := r.ListQuestions(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.TestDetail{
		Test:      *test,
		Questions: questions,
	}, nil
}

func (r *Repository) ListTests(ctx context.Context, filter TestFilter) ([]model.Test, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	query := `SELECT id, title, subject, topic, duration_minutes, audio_url, audio_play_limit, created_at
	FROM test WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Subject != "" {
		query += fmt.Sprintf(` AND subject = $%d`, argIdx)
		args = append(args, filter.Subject)
		argIdx++
	}
	if filter.Topic != "" {
		query += fmt.Sprintf(` AND topic = $%d`, argIdx)
		args = append(args, filter.Topic)
		argIdx++
	}
	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND id > $%d`, argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY id LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tests []model.Test
	for rows.Next() {
		t := model.Test{}
		if err := scanTest(rows, &t); err != nil {
			return nil, "", err
		}
		tests = append(tests, t)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(tests) > filter.Limit {
		nextCursor = tests[filter.Limit].ID.String()
		tests = tests[:filter.Limit]
	}

	return tests, nextCursor, nil
}

func (r *Repository) UpdateTest(ctx context.Context, id uuid.UUID, t *model.Test) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE test
		SET title = $1, subject = $2, topic = $3, duration_minutes = $4, audio_url = $5, audio_play_limit = $6
		WHERE id = $7`,
		t.Title, t.Subject, t.Topic, t.DurationMinutes, t.AudioURL, t.AudioPlayLimit, id,
	)
	return err
}

func (r *Repository) DeleteTest(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM test WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ListQuestions(ctx context.Context, testID uuid.UUID) ([]model.QuestionWithOptions, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, test_id, format, body, correct_answer, explanation, difficulty, image_url, sort_order
		FROM question
		WHERE test_id = $1
		ORDER BY sort_order`,
		testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []model.Question
	for rows.Next() {
		q := model.Question{}
		if err := scanQuestion(rows, &q); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, nil
	}

	questionIDs := make([]uuid.UUID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	opts, err := r.queryOptionsForQuestions(ctx, questionIDs)
	if err != nil {
		return nil, err
	}

	result := make([]model.QuestionWithOptions, len(questions))
	for i, q := range questions {
		result[i] = model.QuestionWithOptions{
			Question: q,
			Options:  opts[q.ID],
		}
	}
	return result, nil
}

func (r *Repository) queryOptionsForQuestions(ctx context.Context, questionIDs []uuid.UUID) (map[uuid.UUID][]model.QuestionOption, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT question_id, key, text, image_url, is_correct, sort_order
		FROM question_option
		WHERE question_id = ANY($1)
		ORDER BY question_id, sort_order`,
		questionIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[uuid.UUID][]model.QuestionOption, len(questionIDs))
	for rows.Next() {
		o := model.QuestionOption{}
		if err := scanQuestionOption(rows, &o); err != nil {
			return nil, err
		}
		out[o.QuestionID] = append(out[o.QuestionID], o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) CreateQuestionTx(ctx context.Context, tx pgx.Tx, q *model.Question, options []model.QuestionOption) error {
	err := tx.QueryRow(ctx,
		`INSERT INTO question (test_id, format, body, correct_answer, explanation, difficulty, image_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		q.TestID, q.Format, q.Body, q.CorrectAnswer, q.Explanation, q.Difficulty, q.ImageURL, q.SortOrder,
	).Scan(&q.ID)
	if err != nil {
		if isSortOrderConflict(err) {
			return ErrSortOrderConflict
		}
		return err
	}

	if err := insertQuestionOptions(ctx, tx, q.ID, options); err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateQuestionTx(ctx context.Context, tx pgx.Tx, q *model.Question, options []model.QuestionOption) error {
	_, err := tx.Exec(ctx,
		`UPDATE question
		SET format = $1, body = $2, correct_answer = $3, explanation = $4, difficulty = $5, image_url = $6, sort_order = $7
		WHERE id = $8`,
		q.Format, q.Body, q.CorrectAnswer, q.Explanation, q.Difficulty, q.ImageURL, q.SortOrder, q.ID,
	)
	if err != nil {
		if isSortOrderConflict(err) {
			return ErrSortOrderConflict
		}
		return err
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM question_option WHERE question_id = $1`,
		q.ID,
	); err != nil {
		return err
	}

	if err := insertQuestionOptions(ctx, tx, q.ID, options); err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM question WHERE id = $1`,
		id,
	)
	return err
}

func insertQuestionOptions(ctx context.Context, tx pgx.Tx, questionID uuid.UUID, options []model.QuestionOption) error {
	for _, o := range options {
		_, err := tx.Exec(ctx,
			`INSERT INTO question_option (question_id, key, text, image_url, is_correct, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			questionID, o.Key, o.Text, o.ImageURL, o.IsCorrect, o.SortOrder,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func isSortOrderConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_question_order" {
		return true
	}
	return false
}