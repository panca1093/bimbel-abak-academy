package repository

import (
	"context"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

// --- Course CRUD ---

func (r *Repository) CreateCourse(ctx context.Context, c model.Course) (model.Course, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO course (title, level, subject, instructor_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, level, subject, instructor_name, created_at, updated_at`,
		c.Title, c.Level, c.Subject, c.InstructorName,
	).Scan(&c.ID, &c.Title, &c.Level, &c.Subject, &c.InstructorName, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *Repository) ListCourses(ctx context.Context) ([]model.Course, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, level, subject, instructor_name, created_at, updated_at
		FROM course
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		err := rows.Scan(&c.ID, &c.Title, &c.Level, &c.Subject, &c.InstructorName, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func (r *Repository) UpdateCourse(ctx context.Context, id uuid.UUID, c model.Course) (model.Course, error) {
	result := model.Course{}
	err := r.pool.QueryRow(ctx,
		`UPDATE course
		SET title = $1, level = $2, subject = $3, instructor_name = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, title, level, subject, instructor_name, created_at, updated_at`,
		c.Title, c.Level, c.Subject, c.InstructorName, id,
	).Scan(&result.ID, &result.Title, &result.Level, &result.Subject, &result.InstructorName, &result.CreatedAt, &result.UpdatedAt)
	return result, err
}

// GetCoursesByProductID returns all courses linked to a product via product_course.
// Returns an empty slice (not error) when no links exist.
func (r *Repository) GetCoursesByProductID(ctx context.Context, productID uuid.UUID) ([]model.Course, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.title, c.level, c.subject, c.instructor_name, c.created_at, c.updated_at
		FROM course c
		JOIN product_course pc ON pc.course_id = c.id
		WHERE pc.product_id = $1
		ORDER BY c.title`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		err := rows.Scan(&c.ID, &c.Title, &c.Level, &c.Subject, &c.InstructorName, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

// CountLessonsByCourse counts all lessons across all sections of a course.
func (r *Repository) CountLessonsByCourse(ctx context.Context, courseID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)
		FROM lesson l
		JOIN section s ON l.section_id = s.id
		WHERE s.course_id = $1`,
		courseID,
	).Scan(&count)
	return count, err
}

// --- Section CRUD (re-keyed to course_id) ---

func (r *Repository) ListSections(ctx context.Context, courseID uuid.UUID) ([]model.Section, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, course_id, title, position, created_at
		FROM section
		WHERE course_id = $1
		ORDER BY position ASC`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []model.Section
	for rows.Next() {
		s := model.Section{}
		err := rows.Scan(&s.ID, &s.CourseID, &s.Title, &s.Position, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, rows.Err()
}

func (r *Repository) CreateSection(ctx context.Context, s model.Section) (model.Section, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO section (course_id, title, position)
		VALUES ($1, $2, $3)
		RETURNING id, course_id, title, position, created_at`,
		s.CourseID, s.Title, s.Position,
	).Scan(&s.ID, &s.CourseID, &s.Title, &s.Position, &s.CreatedAt)
	return s, err
}

func (r *Repository) UpdateSection(ctx context.Context, id uuid.UUID, title string) (model.Section, error) {
	s := model.Section{}
	err := r.pool.QueryRow(ctx,
		`UPDATE section SET title = $1 WHERE id = $2
		RETURNING id, course_id, title, position, created_at`,
		title, id,
	).Scan(&s.ID, &s.CourseID, &s.Title, &s.Position, &s.CreatedAt)
	return s, err
}

func (r *Repository) DeleteSection(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM section WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ReorderSections(ctx context.Context, courseID uuid.UUID, orderedIDs []uuid.UUID) error {
	for i, id := range orderedIDs {
		_, err := r.pool.Exec(ctx,
			`UPDATE section SET position = $1 WHERE id = $2`,
			i, id,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// --- Lesson CRUD (unchanged, references section_id) ---

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
