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

// scanTestWithCount is used by ListTests where the SELECT also LEFT JOINs a
// grouped question count; keeps GetByID/CreateTest untouched.
func scanTestWithCount(row interface{ Scan(dest ...any) error }, t *model.Test) error {
	var audioURL *string
	var audioPlayLimit *int
	err := row.Scan(
		&t.ID, &t.Title, &t.Subject, &t.Topic, &t.DurationMinutes,
		&audioURL, &audioPlayLimit, &t.QuestionCount, &t.CreatedAt,
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

	// LEFT JOIN with a grouped count keeps tests without questions counted as 0.
	// We count per test_id and join back — the GROUP BY in the subquery avoids
	// inflating the row count over the outer filters (subject/topic/cursor).
	query := `SELECT t.id, t.title, t.subject, t.topic, t.duration_minutes,
		t.audio_url, t.audio_play_limit, COALESCE(q.cnt, 0), t.created_at
	FROM test t
	LEFT JOIN (
		SELECT test_id, COUNT(*) AS cnt FROM question GROUP BY test_id
	) q ON q.test_id = t.id
	WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Subject != "" {
		query += fmt.Sprintf(` AND t.subject = $%d`, argIdx)
		args = append(args, filter.Subject)
		argIdx++
	}
	if filter.Topic != "" {
		query += fmt.Sprintf(` AND t.topic = $%d`, argIdx)
		args = append(args, filter.Topic)
		argIdx++
	}
	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND t.id > $%d`, argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY t.id LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tests []model.Test
	for rows.Next() {
		t := model.Test{}
		if err := scanTestWithCount(rows, &t); err != nil {
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
	var updatedID uuid.UUID
	err := tx.QueryRow(ctx,
		`UPDATE question
		SET format = $1, body = $2, correct_answer = $3, explanation = $4, difficulty = $5, image_url = $6, sort_order = $7
		WHERE id = $8 RETURNING id`,
		q.Format, q.Body, q.CorrectAnswer, q.Explanation, q.Difficulty, q.ImageURL, q.SortOrder, q.ID,
	).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
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

// ExamFilter mirrors TestFilter/ProductFilter for cursor-paginated ListExams.
type ExamFilter struct {
	Cursor string
	Limit  int
}

func scanExam(row interface{ Scan(dest ...any) error }, e *model.Exam) error {
	err := row.Scan(
		&e.ID, &e.Title, &e.IsFree, &e.ScheduledAt,
		&e.RequiresCheckin, &e.AllowLeaderboard, &e.CDNBundle,
		&e.BundleURL, &e.BundleGeneratedAt,
		&e.CheckInWindowMinutes, &e.GraceWindowMinutes, &e.MaxAttempts,
		&e.TimerMode, &e.DurationMinutes, &e.Randomize,
		&e.ResultConfig, &e.ResultReleaseAt, &e.Status, &e.ProductID, &e.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func scanExamListItem(row interface{ Scan(dest ...any) error }, out *model.ExamListItem) error {
	var productPrice *int64
	var productStatus *string
	err := row.Scan(
		&out.ID, &out.Title, &out.IsFree, &out.ScheduledAt,
		&out.RequiresCheckin, &out.AllowLeaderboard, &out.CDNBundle,
		&out.BundleURL, &out.BundleGeneratedAt,
		&out.CheckInWindowMinutes, &out.GraceWindowMinutes, &out.MaxAttempts,
		&out.TimerMode, &out.DurationMinutes, &out.Randomize,
		&out.ResultConfig, &out.ResultReleaseAt, &out.Status, &out.ProductID, &out.CreatedAt,
		&productPrice, &productStatus,
	)
	if err != nil {
		return err
	}
	if productPrice != nil {
		out.ProductPrice = *productPrice
	}
	if productStatus != nil {
		out.ProductStatus = *productStatus
	}
	return nil
}

// CreateProductAndExamTx opens a transaction, inserts the linked product (type=exam,
// status=draft, price=0), inserts the exam, and commits. Both rows are returned.
func (r *Repository) CreateProductAndExamTx(ctx context.Context, e *model.Exam) (model.Exam, model.Product, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Exam{}, model.Product{}, err
	}
	defer tx.Rollback(ctx)

	p := model.Product{
		Type:   "exam",
		Name:   e.Title,
		Status: "draft",
		Price:  0,
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO product (type, name, description, price, stock, status, weight_grams, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		p.Type, p.Name, p.Description, p.Price, p.Stock, p.Status, p.WeightGrams, p.ImageURL,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return model.Exam{}, model.Product{}, err
	}

	productID, err := uuid.Parse(p.ID)
	if err != nil {
		return model.Exam{}, model.Product{}, err
	}
	if err := createExam(ctx, tx, e, &productID); err != nil {
		return model.Exam{}, model.Product{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Exam{}, model.Product{}, err
	}
	return *e, p, nil
}

// CreateExamTx inserts an exam row inside an existing transaction; caller supplies
// productID and is responsible for committing/rolling back.
func (r *Repository) CreateExamTx(ctx context.Context, tx pgx.Tx, e *model.Exam) error {
	productID := e.ProductID
	if productID == nil {
		return errors.New("CreateExamTx requires exam.ProductID to be set")
	}
	return createExam(ctx, tx, e, productID)
}

func createExam(ctx context.Context, tx pgx.Tx, e *model.Exam, productID *uuid.UUID) error {
	return tx.QueryRow(ctx,
		`INSERT INTO exam (title, is_free, scheduled_at, requires_checkin, allow_leaderboard,
			cdn_bundle, bundle_url, bundle_generated_at, check_in_window_minutes, grace_window_minutes,
			max_attempts, timer_mode, duration_minutes, randomize, result_config, result_release_at,
			status, product_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at`,
		e.Title, e.IsFree, e.ScheduledAt, e.RequiresCheckin, e.AllowLeaderboard,
		e.CDNBundle, e.BundleURL, e.BundleGeneratedAt, e.CheckInWindowMinutes, e.GraceWindowMinutes,
		e.MaxAttempts, e.TimerMode, e.DurationMinutes, e.Randomize, e.ResultConfig, e.ResultReleaseAt,
		e.Status, productID,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *Repository) GetExamByID(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	out := &model.Exam{}
	err := scanExam(r.pool.QueryRow(ctx,
		`SELECT id, title, is_free, scheduled_at, requires_checkin, allow_leaderboard,
			cdn_bundle, bundle_url, bundle_generated_at, check_in_window_minutes, grace_window_minutes,
			max_attempts, timer_mode, duration_minutes, randomize, result_config, result_release_at,
			status, product_id, created_at
		FROM exam
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

func (r *Repository) ListExams(ctx context.Context, filter ExamFilter) ([]model.ExamListItem, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	query := `SELECT e.id, e.title, e.is_free, e.scheduled_at, e.requires_checkin, e.allow_leaderboard,
		e.cdn_bundle, e.bundle_url, e.bundle_generated_at, e.check_in_window_minutes, e.grace_window_minutes,
		e.max_attempts, e.timer_mode, e.duration_minutes, e.randomize, e.result_config, e.result_release_at,
		e.status, e.product_id, e.created_at, p.price, p.status
	FROM exam e
	LEFT JOIN product p ON p.id = e.product_id
	WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND e.id > $%d`, argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY e.id LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []model.ExamListItem
	for rows.Next() {
		var item model.ExamListItem
		if err := scanExamListItem(rows, &item); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(items) > filter.Limit {
		nextCursor = items[filter.Limit].ID.String()
		items = items[:filter.Limit]
	}

	return items, nextCursor, nil
}

func (r *Repository) GetExamDetail(ctx context.Context, id uuid.UUID) (*model.ExamDetail, error) {
	detail := &model.ExamDetail{}
	err := scanExam(r.pool.QueryRow(ctx,
		`SELECT e.id, e.title, e.is_free, e.scheduled_at, e.requires_checkin, e.allow_leaderboard,
			e.cdn_bundle, e.bundle_url, e.bundle_generated_at, e.check_in_window_minutes, e.grace_window_minutes,
			e.max_attempts, e.timer_mode, e.duration_minutes, e.randomize, e.result_config, e.result_release_at,
			e.status, e.product_id, e.created_at
		FROM exam e
		WHERE e.id = $1`,
		id,
	), &detail.Exam)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var productPrice *int64
	var productStatus *string
	err = r.pool.QueryRow(ctx,
		`SELECT p.price, p.status FROM product p WHERE p.id = $1`,
		detail.ProductID,
	).Scan(&productPrice, &productStatus)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if productPrice != nil {
		detail.ProductPrice = *productPrice
	}
	if productStatus != nil {
		detail.ProductStatus = *productStatus
	}

	rows, err := r.pool.Query(ctx,
		`SELECT et.id, et.exam_id, et.test_id, et.sort_order,
			t.id, t.title, t.subject, t.topic, t.duration_minutes,
			COALESCE(q.cnt, 0)
		FROM exam_test et
		JOIN test t ON t.id = et.test_id
		LEFT JOIN (
			SELECT test_id, COUNT(*) AS cnt FROM question GROUP BY test_id
		) q ON q.test_id = t.id
		WHERE et.exam_id = $1
		ORDER BY et.sort_order ASC`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []model.ExamTestEntry
	for rows.Next() {
		var entry model.ExamTestEntry
		var topic *string
		if err := rows.Scan(
			&entry.ID, &entry.ExamID, &entry.TestID, &entry.SortOrder,
			&entry.Test.ID, &entry.Test.Title, &entry.Test.Subject, &topic, &entry.Test.DurationMinutes,
			&entry.Test.QuestionCount,
		); err != nil {
			return nil, err
		}
		entry.Test.Topic = topic
		tests = append(tests, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tests == nil {
		tests = []model.ExamTestEntry{}
	}
	detail.Tests = tests

	return detail, nil
}

func (r *Repository) UpdateExam(ctx context.Context, id uuid.UUID, e *model.Exam) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE exam
		SET title = $1, is_free = $2, scheduled_at = $3, requires_checkin = $4, allow_leaderboard = $5,
			cdn_bundle = $6, bundle_url = $7, bundle_generated_at = $8,
			check_in_window_minutes = $9, grace_window_minutes = $10, max_attempts = $11,
			timer_mode = $12, duration_minutes = $13, randomize = $14,
			result_config = $15, result_release_at = $16, status = $17
		WHERE id = $18`,
		e.Title, e.IsFree, e.ScheduledAt, e.RequiresCheckin, e.AllowLeaderboard,
		e.CDNBundle, e.BundleURL, e.BundleGeneratedAt,
		e.CheckInWindowMinutes, e.GraceWindowMinutes, e.MaxAttempts,
		e.TimerMode, e.DurationMinutes, e.Randomize,
		e.ResultConfig, e.ResultReleaseAt, e.Status, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ReplaceExamTestsTx atomically replaces all exam_test links for an exam. Caller
// supplies the tx; verification of test existence is the caller's responsibility
// (mirrors ReplaceProductCourses).
func (r *Repository) ReplaceExamTestsTx(ctx context.Context, tx pgx.Tx, examID uuid.UUID, tests []model.ExamTest) error {
	if _, err := tx.Exec(ctx, `DELETE FROM exam_test WHERE exam_id = $1`, examID); err != nil {
		return err
	}
	for _, t := range tests {
		if _, err := tx.Exec(ctx,
			`INSERT INTO exam_test (exam_id, test_id, sort_order) VALUES ($1, $2, $3)`,
			examID, t.TestID, t.SortOrder,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpdateProductPriceTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID, price int64) error {
	_, err := tx.Exec(ctx,
		`UPDATE product SET price = $1, updated_at = now() WHERE id = $2`,
		price, productID,
	)
	return err
}