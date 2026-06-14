package repository

import (
	"context"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

func (r *Repository) ListSections(ctx context.Context, productID uuid.UUID) ([]model.CourseSection, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, title, position, created_at
		FROM course_section
		WHERE product_id = $1
		ORDER BY position ASC`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []model.CourseSection
	for rows.Next() {
		s := model.CourseSection{}
		err := rows.Scan(&s.ID, &s.ProductID, &s.Title, &s.Position, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, rows.Err()
}

func (r *Repository) CreateSection(ctx context.Context, s model.CourseSection) (model.CourseSection, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO course_section (product_id, title, position)
		VALUES ($1, $2, $3)
		RETURNING id, product_id, title, position, created_at`,
		s.ProductID, s.Title, s.Position,
	).Scan(&s.ID, &s.ProductID, &s.Title, &s.Position, &s.CreatedAt)
	return s, err
}

func (r *Repository) UpdateSection(ctx context.Context, id uuid.UUID, title string) (model.CourseSection, error) {
	s := model.CourseSection{}
	err := r.pool.QueryRow(ctx,
		`UPDATE course_section SET title = $1 WHERE id = $2
		RETURNING id, product_id, title, position, created_at`,
		title, id,
	).Scan(&s.ID, &s.ProductID, &s.Title, &s.Position, &s.CreatedAt)
	return s, err
}

func (r *Repository) DeleteSection(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM course_section WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ReorderSections(ctx context.Context, productID uuid.UUID, orderedIDs []uuid.UUID) error {
	for i, id := range orderedIDs {
		_, err := r.pool.Exec(ctx,
			`UPDATE course_section SET position = $1 WHERE id = $2`,
			i, id,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateLesson(ctx context.Context, l model.Lesson) (model.Lesson, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lesson (section_id, title, video_url, duration_seconds, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, section_id, title, video_url, duration_seconds, position, created_at`,
		l.SectionID, l.Title, l.VideoURL, l.DurationSeconds, l.Position,
	).Scan(&l.ID, &l.SectionID, &l.Title, &l.VideoURL, &l.DurationSeconds, &l.Position, &l.CreatedAt)
	return l, err
}

func (r *Repository) UpdateLesson(ctx context.Context, id uuid.UUID, l model.Lesson) (model.Lesson, error) {
	result := model.Lesson{}
	err := r.pool.QueryRow(ctx,
		`UPDATE lesson SET title = $1, video_url = $2, duration_seconds = $3, position = $4 WHERE id = $5
		RETURNING id, section_id, title, video_url, duration_seconds, position, created_at`,
		l.Title, l.VideoURL, l.DurationSeconds, l.Position, id,
	).Scan(&result.ID, &result.SectionID, &result.Title, &result.VideoURL, &result.DurationSeconds, &result.Position, &result.CreatedAt)
	return result, err
}

func (r *Repository) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM lesson WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ReorderLessons(ctx context.Context, sectionID uuid.UUID, orderedIDs []uuid.UUID) error {
	for i, id := range orderedIDs {
		_, err := r.pool.Exec(ctx,
			`UPDATE lesson SET position = $1 WHERE id = $2`,
			i, id,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListLessonsBySection(ctx context.Context, sectionID uuid.UUID) ([]model.Lesson, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, section_id, title, video_url, duration_seconds, position, created_at
		FROM lesson
		WHERE section_id = $1
		ORDER BY position ASC`,
		sectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []model.Lesson
	for rows.Next() {
		l := model.Lesson{}
		err := rows.Scan(&l.ID, &l.SectionID, &l.Title, &l.VideoURL, &l.DurationSeconds, &l.Position, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}
