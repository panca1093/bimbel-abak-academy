package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *Repository) GetExamByProductID(ctx context.Context, productID uuid.UUID) (*model.Exam, error) {
	out := &model.Exam{}
	err := scanExam(r.pool.QueryRow(ctx,
		`SELECT id, title, is_free, scheduled_at, requires_checkin, allow_leaderboard,
			cdn_bundle, bundle_url, bundle_generated_at, check_in_window_minutes, grace_window_minutes,
			max_attempts, timer_mode, duration_minutes, randomize, result_config, result_release_at,
			status, product_id, created_at
		FROM exam
		WHERE product_id = $1`,
		productID,
	), out)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return out, nil
}

// CreateExamRegistration inserts a row using ON CONFLICT DO NOTHING — outbox
// re-delivery (same OrderPaid event processed twice) collapses to a no-op when
// (student_id, exam_id) already exists. RowsAffected == 0 is success, not error.
func (r *Repository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, token, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (student_id, exam_id) DO NOTHING`,
		reg.StudentID, reg.ExamID, reg.Token, reg.Status,
	)
	return err
}

func (r *Repository) StampOrderItemFulfilledAt(ctx context.Context, tx pgx.Tx, orderID, productID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE order_items SET fulfilled_at = now() WHERE order_id = $1 AND product_id = $2`,
		orderID, productID,
	)
	return err
}

func (r *Repository) GetExamRegistrationsByStudent(ctx context.Context, studentID uuid.UUID) ([]model.RegistrationListItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT reg.id, reg.student_id, reg.exam_id, reg.token, reg.card_pdf_url,
			reg.checked_in_at, reg.attempts_used, reg.status, reg.created_at,
			e.title, e.scheduled_at
		FROM exam_registration reg
		JOIN exam e ON e.id = reg.exam_id
		WHERE reg.student_id = $1
		ORDER BY reg.created_at DESC`,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.RegistrationListItem
	for rows.Next() {
		var item model.RegistrationListItem
		var cardPDFURL *string
		var checkedInAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.StudentID, &item.ExamID, &item.Token, &cardPDFURL,
			&checkedInAt, &item.AttemptsUsed, &item.Status, &item.CreatedAt,
			&item.ExamTitle, &item.ScheduledAt,
		); err != nil {
			return nil, err
		}
		if cardPDFURL != nil {
			item.CardPDFURL = cardPDFURL
		}
		if checkedInAt != nil {
			item.CheckedInAt = checkedInAt
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.RegistrationListItem{}
	}
	return items, nil
}

func (r *Repository) GetExamRegistrationByID(ctx context.Context, regID, studentID uuid.UUID) (*model.RegistrationDetail, error) {
	var detail model.RegistrationDetail
	var cardPDFURL *string
	var checkedInAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT reg.id, reg.student_id, reg.exam_id, reg.token, reg.card_pdf_url,
			reg.checked_in_at, reg.attempts_used, reg.status, reg.created_at,
			e.id, e.title, e.scheduled_at, e.requires_checkin, e.check_in_window_minutes,
			e.timer_mode, e.duration_minutes, e.result_config
		FROM exam_registration reg
		JOIN exam e ON e.id = reg.exam_id
		WHERE reg.id = $1 AND reg.student_id = $2`,
		regID, studentID,
	).Scan(
		&detail.ID, &detail.StudentID, &detail.ExamID, &detail.Token, &cardPDFURL,
		&checkedInAt, &detail.AttemptsUsed, &detail.Status, &detail.CreatedAt,
		&detail.Exam.ID, &detail.Exam.Title, &detail.Exam.ScheduledAt, &detail.Exam.RequiresCheckin,
		&detail.Exam.CheckInWindowMinutes, &detail.Exam.TimerMode, &detail.Exam.DurationMinutes,
		&detail.Exam.ResultConfig,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if cardPDFURL != nil {
		detail.CardPDFURL = cardPDFURL
	}
	if checkedInAt != nil {
		detail.CheckedInAt = checkedInAt
	}
	return &detail, nil
}
// ---------- Session scan helpers ----------

func scanExamSession(row interface{ Scan(dest ...any) error }, s *model.ExamSession) error {
	return row.Scan(
		&s.ID, &s.RegistrationID, &s.StudentID, &s.ExamID,
		&s.AttemptNumber, &s.StartedAt, &s.SubmittedAt,
		&s.ExtendedUntil, &s.AdminSubmitted, &s.Score,
		&s.CertificateURL, &s.LastSavedAt, &s.Status, &s.CreatedAt,
	)
}

func scanExamSessionAnswer(row interface{ Scan(dest ...any) error }, a *model.ExamSessionAnswer) error {
	return row.Scan(
		&a.SessionID, &a.QuestionID, &a.Answer, &a.IsCorrect, &a.Score,
		&a.GradedBy, &a.GradedAt, &a.GraderComment, &a.FlaggedForReview, &a.SavedAt,
	)
}

// ---------- Session repository methods ----------

// GetExamRegistrationByToken retrieves a registration by student ID and token.
// Returns ErrNotFound when no match exists.
func (r *Repository) GetExamRegistrationByToken(ctx context.Context, studentID uuid.UUID, token string) (*model.ExamRegistration, error) {
	var reg model.ExamRegistration
	var cardPDFURL *string
	var checkedInAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, exam_id, token, card_pdf_url, checked_in_at, attempts_used, status, created_at
		FROM exam_registration
		WHERE student_id = $1 AND token = $2`,
		studentID, token,
	).Scan(
		&reg.ID, &reg.StudentID, &reg.ExamID, &reg.Token,
		&cardPDFURL, &checkedInAt, &reg.AttemptsUsed, &reg.Status, &reg.CreatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if cardPDFURL != nil {
		reg.CardPDFURL = cardPDFURL
	}
	if checkedInAt != nil {
		reg.CheckedInAt = checkedInAt
	}
	return &reg, nil
}

// CheckInExamTx stamps checked_in_at (if NULL) and sets status='checked_in'.
func (r *Repository) CheckInExamTx(ctx context.Context, tx pgx.Tx, regID uuid.UUID) error {
	tag, err := tx.Exec(ctx,
		`UPDATE exam_registration
		SET checked_in_at = COALESCE(checked_in_at, now()), status = 'checked_in'
		WHERE id = $1`,
		regID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateExamSessionTx increments attempts_used, sets status='in_progress',
// optionally stamps checked_in_at when NULL, and inserts an exam_session row.
func (r *Repository) CreateExamSessionTx(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) (model.ExamSession, error) {
	_, err := tx.Exec(ctx,
		`UPDATE exam_registration
		SET attempts_used = attempts_used + 1,
		    status = 'in_progress',
		    checked_in_at = COALESCE(checked_in_at, now())
		WHERE id = $1`,
		reg.ID,
	)
	if err != nil {
		return model.ExamSession{}, err
	}

	var s model.ExamSession
	err = tx.QueryRow(ctx,
		`INSERT INTO exam_session (registration_id, student_id, exam_id, attempt_number, started_at, status)
		VALUES ($1, $2, $3, 1, now(), 'in_progress')
		RETURNING id, registration_id, student_id, exam_id, attempt_number, started_at,
			submitted_at, extended_until, admin_submitted, score, certificate_url,
			last_saved_at, status, created_at`,
		reg.ID, reg.StudentID, reg.ExamID,
	).Scan(
		&s.ID, &s.RegistrationID, &s.StudentID, &s.ExamID,
		&s.AttemptNumber, &s.StartedAt, &s.SubmittedAt,
		&s.ExtendedUntil, &s.AdminSubmitted, &s.Score,
		&s.CertificateURL, &s.LastSavedAt, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		return model.ExamSession{}, err
	}
	return s, nil
}

// GetExamSessionForStudent returns a session scoped to the owning student.
func (r *Repository) GetExamSessionForStudent(ctx context.Context, sessionID, studentID uuid.UUID) (*model.ExamSession, error) {
	var s model.ExamSession
	err := scanExamSession(r.pool.QueryRow(ctx,
		`SELECT id, registration_id, student_id, exam_id, attempt_number, started_at,
			submitted_at, extended_until, admin_submitted, score, certificate_url,
			last_saved_at, status, created_at
		FROM exam_session
		WHERE id = $1 AND student_id = $2`,
		sessionID, studentID,
	), &s)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// GetSessionWithQuestions returns the ordered test->question->option tree for an exam.
// Reuses GetTestDetail for each attached test.
func (r *Repository) GetSessionWithQuestions(ctx context.Context, examID uuid.UUID) ([]model.TestDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT et.test_id
		FROM exam_test et
		WHERE et.exam_id = $1
		ORDER BY et.sort_order ASC`,
		examID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testIDs []uuid.UUID
	for rows.Next() {
		var testID uuid.UUID
		if err := rows.Scan(&testID); err != nil {
			return nil, err
		}
		testIDs = append(testIDs, testID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(testIDs) == 0 {
		return nil, nil
	}

	result := make([]model.TestDetail, len(testIDs))
	for i, tid := range testIDs {
		detail, err := r.GetTestDetail(ctx, tid)
		if err != nil {
			return nil, err
		}
		result[i] = *detail
	}
	return result, nil
}

// GetSessionAnswers returns all answers for a session ordered by question_id.
func (r *Repository) GetSessionAnswers(ctx context.Context, sessionID uuid.UUID) ([]model.ExamSessionAnswer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT session_id, question_id, answer, is_correct, score, graded_by,
			graded_at, grader_comment, flagged_for_review, saved_at
		FROM exam_session_answer
		WHERE session_id = $1
		ORDER BY question_id`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []model.ExamSessionAnswer
	for rows.Next() {
		var a model.ExamSessionAnswer
		if err := scanExamSessionAnswer(rows, &a); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if answers == nil {
		answers = []model.ExamSessionAnswer{}
	}
	return answers, nil
}

// SaveAnswersTx upserts answers and stamps last_saved_at on the session.
func (r *Repository) SaveAnswersTx(ctx context.Context, sessionID uuid.UUID, answers []model.ExamSessionAnswer) error {
	for _, a := range answers {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO exam_session_answer (session_id, question_id, answer, is_correct, score, graded_by, graded_at, grader_comment, flagged_for_review, saved_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
			ON CONFLICT (session_id, question_id) DO UPDATE SET
				answer = EXCLUDED.answer,
				is_correct = EXCLUDED.is_correct,
				score = EXCLUDED.score,
				graded_by = EXCLUDED.graded_by,
				graded_at = EXCLUDED.graded_at,
				grader_comment = EXCLUDED.grader_comment,
				flagged_for_review = EXCLUDED.flagged_for_review,
				saved_at = now()`,
			sessionID, a.QuestionID, a.Answer, a.IsCorrect, a.Score,
			a.GradedBy, a.GradedAt, a.GraderComment, a.FlaggedForReview,
		)
		if err != nil {
			return err
		}
	}

	_, err := r.pool.Exec(ctx,
		`UPDATE exam_session SET last_saved_at = now() WHERE id = $1`,
		sessionID,
	)
	return err
}

// SubmitSessionTx performs a CAS submit of a session, writes graded answers,
// and sets the overall score. Returns the number of rows affected by the CAS update.
func (r *Repository) SubmitSessionTx(ctx context.Context, tx pgx.Tx, sessionID uuid.UUID, graded []model.ExamSessionAnswer, score float64, adminSubmitted bool) (int64, error) {
	query := `UPDATE exam_session SET status = 'submitted', submitted_at = now()`
	if adminSubmitted {
		query += `, admin_submitted = true`
	}
	query += ` WHERE id = $1 AND status = 'in_progress'`

	tag, err := tx.Exec(ctx, query, sessionID)
	if err != nil {
		return 0, err
	}

	if tag.RowsAffected() == 1 {
		for _, a := range graded {
			_, err := tx.Exec(ctx,
				`INSERT INTO exam_session_answer (session_id, question_id, answer, is_correct, score, graded_by, graded_at, grader_comment, flagged_for_review, saved_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
				ON CONFLICT (session_id, question_id) DO UPDATE SET
					answer = EXCLUDED.answer,
					is_correct = EXCLUDED.is_correct,
					score = EXCLUDED.score,
					graded_by = EXCLUDED.graded_by,
					graded_at = now(),
					grader_comment = EXCLUDED.grader_comment,
					flagged_for_review = EXCLUDED.flagged_for_review,
					saved_at = now()`,
				sessionID, a.QuestionID, a.Answer, a.IsCorrect, a.Score,
				a.GradedBy, a.GradedAt, a.GraderComment, a.FlaggedForReview,
			)
			if err != nil {
				return 0, err
			}
		}

		_, err = tx.Exec(ctx,
			`UPDATE exam_session SET score = $1 WHERE id = $2`,
			score, sessionID,
		)
		if err != nil {
			return 0, err
		}
	}

	return tag.RowsAffected(), nil
}

// LogViolation records an integrity event for a session.
func (r *Repository) LogViolation(ctx context.Context, v model.SessionViolationLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_violation_log (session_id, student_id, violation_type, occurred_at)
		VALUES ($1, $2, $3, $4)`,
		v.SessionID, v.StudentID, v.ViolationType, v.OccurredAt,
	)
	return err
}

// ReopenSession extends a session by the given minutes. Only applies to
// in_progress or submitted sessions. Returns ErrNotFound if no session matched.
func (r *Repository) ReopenSession(ctx context.Context, sessionID uuid.UUID, minutes int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE exam_session
		SET extended_until = now() + make_interval(mins => $2)
		WHERE id = $1 AND status IN ('in_progress', 'submitted')`,
		sessionID, minutes,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetExamForSession retrieves an exam by ID. Delegates to GetExamByID.
func (r *Repository) GetExamForSession(ctx context.Context, examID uuid.UUID) (*model.Exam, error) {
	return r.GetExamByID(ctx, examID)
}
