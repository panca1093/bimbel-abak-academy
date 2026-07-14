package integration_test

import (
	"context"
	"testing"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise the REAL service + REAL Postgres (no shim/fake), so a
// regression in the production preservation logic would fail them. They
// replace the tautological shim/fake tests in internal/service/{store,course}_test.go.

// Finding 1 / item-13: editing only level must not blank the title (or other fields).
func TestUpdateCourse_PreservesTitleAndFields_RealService(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	course, err := env.svc.CreateCourse(ctx, "Matematika Dasar", "SMA", "Matematika", "Budi", service.RoleAdminStore)
	require.NoError(t, err)

	// Partial update: only level changes; title/subject/instructor sent as "".
	updated, err := env.svc.UpdateCourse(ctx, course.ID.String(), "", "SMP", "", "", service.RoleAdminStore)
	require.NoError(t, err)

	assert.Equal(t, "SMP", updated.Level, "level should update to SMP")
	assert.Equal(t, "Matematika Dasar", updated.Title, "title must be preserved, not blanked")
	assert.Equal(t, "Matematika", updated.Subject, "subject must be preserved")
	assert.Equal(t, "Budi", updated.InstructorName, "instructor must be preserved")

	// Verify the persisted row, not just the returned struct.
	var title, subject, instructor string
	require.NoError(t, env.pool.QueryRow(ctx,
		`SELECT title, subject, instructor_name FROM course WHERE id = $1`,
		course.ID).Scan(&title, &subject, &instructor))
	assert.Equal(t, "Matematika Dasar", title, "persisted title must be preserved")
	assert.Equal(t, "Matematika", subject, "persisted subject must be preserved")
	assert.Equal(t, "Budi", instructor, "persisted instructor must be preserved")
}

// Finding 6 / Bug C: UpdateProduct must preserve Type/WeightGrams/ImageURL from
// the stored row when the request omits them.
func TestUpdateProduct_PreservesTypeWeightImage_RealService(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	created, err := env.svc.CreateProduct(ctx, model.Product{
		Type:        "book",
		Name:        "Buku Original",
		WeightGrams: 500,
		ImageURL:    "http://example.com/cover.jpg",
		Price:       10000,
		Stock:       10,
		Status:      "published",
	}, service.RoleAdminStore)
	require.NoError(t, err)

	// Edit only Name; Type/WeightGrams/ImageURL are zero in the request.
	updated, err := env.svc.UpdateProduct(ctx, created.ID, model.Product{
		Name:   "Buku Renamed",
		Price:  10000,
		Stock:  10,
		Status: "published",
	}, service.RoleAdminStore)
	require.NoError(t, err)

	assert.Equal(t, "book", updated.Type, "type must be preserved")
	assert.Equal(t, 500, updated.WeightGrams, "weight_grams must be preserved")
	assert.Equal(t, "http://example.com/cover.jpg", updated.ImageURL, "image_url must be preserved")

	var ptype string
	var weight int
	var image string
	require.NoError(t, env.pool.QueryRow(ctx,
		`SELECT type, weight_grams, image_url FROM product WHERE id = $1`,
		created.ID).Scan(&ptype, &weight, &image))
	assert.Equal(t, "book", ptype)
	assert.Equal(t, 500, weight)
	assert.Equal(t, "http://example.com/cover.jpg", image)
}

// Finding 7 / Bug D: Repository.UpdateLesson must preserve position when only
// title/video/duration are supplied (the SQL SET clause omits position).
func TestUpdateLesson_PreservesPosition_RealRepo(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	course, err := env.svc.CreateCourse(ctx, "Course L", "beginner", "math", "Mr. A", service.RoleAdminStore)
	require.NoError(t, err)
	sec, err := env.svc.CreateSection(ctx, course.ID.String(), "Intro", service.RoleAdminStore)
	require.NoError(t, err)

	l1, err := env.svc.CreateLesson(ctx, course.ID.String(), sec.ID.String(), "Welcome", "https://v/1", 300, service.RoleAdminStore)
	require.NoError(t, err)
	require.Equal(t, 0, l1.Position)

	l2, err := env.svc.CreateLesson(ctx, course.ID.String(), sec.ID.String(), "Basics", "https://v/2", 400, service.RoleAdminStore)
	require.NoError(t, err)
	require.Equal(t, 1, l2.Position)

	// Call the real repository boundary (the Bug D fix lives in the SQL SET clause).
	repo := repository.New(env.pool)
	updated, err := repo.UpdateLesson(ctx, l2.ID, model.Lesson{
		Title:           "Basics Updated",
		VideoURL:        "https://v/2",
		DurationSeconds: 400,
	})
	require.NoError(t, err)

	assert.Equal(t, "Basics Updated", updated.Title)
	assert.Equal(t, 1, updated.Position, "position must be preserved, not reset to 0")

	// Verify the persisted row.
	var pos int
	require.NoError(t, env.pool.QueryRow(ctx,
		`SELECT position FROM lesson WHERE id = $1`, l2.ID).Scan(&pos))
	assert.Equal(t, 1, pos, "persisted position must be preserved")
}