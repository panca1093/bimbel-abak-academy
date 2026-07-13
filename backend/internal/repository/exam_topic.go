package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

// TopicFilter is the filter for the topic list endpoint (FR-16).
type TopicFilter struct {
	Subject string
}

// CreateTopic inserts a new exam_topic row.
func (r *Repository) CreateTopic(ctx context.Context, t *model.ExamTopic) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO exam_topic (name, subject) VALUES ($1, $2) RETURNING id, created_at`,
		t.Name, t.Subject,
	).Scan(&t.ID, &t.CreatedAt)
	return err
}

// GetTopicByID returns a single topic by ID.
func (r *Repository) GetTopicByID(ctx context.Context, id uuid.UUID) (*model.ExamTopic, error) {
	var t model.ExamTopic
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, subject, created_at FROM exam_topic WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.Subject, &t.CreatedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

// ListTopics returns all topics with a per-topic question count (FR-16).
func (r *Repository) ListTopics(ctx context.Context, filter TopicFilter) ([]model.ExamTopic, error) {
	query := `SELECT et.id, et.name, et.subject, COUNT(q.id)
FROM exam_topic et
LEFT JOIN question q ON q.topic_id = et.id
WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Subject != "" {
		query += fmt.Sprintf(` AND et.subject = $%d`, argIdx)
		args = append(args, filter.Subject)
		argIdx++
	}

	query += ` GROUP BY et.id, et.name, et.subject ORDER BY et.subject, et.name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []model.ExamTopic
	for rows.Next() {
		var t model.ExamTopic
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.QuestionCount); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if topics == nil {
		topics = []model.ExamTopic{}
	}
	return topics, nil
}

// UpdateTopic changes a topic's name and subject.
func (r *Repository) UpdateTopic(ctx context.Context, id uuid.UUID, t *model.ExamTopic) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE exam_topic SET name = $1, subject = $2 WHERE id = $3`,
		t.Name, t.Subject, id,
	)
	return err
}

// DeleteTopic removes a topic by ID.
func (r *Repository) DeleteTopic(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM exam_topic WHERE id = $1`,
		id,
	)
	return err
}

// CountQuestionsByTopic returns the number of questions referencing a topic.
func (r *Repository) CountQuestionsByTopic(ctx context.Context, topicID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM question WHERE topic_id = $1`,
		topicID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
