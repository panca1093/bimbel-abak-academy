package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigration0028_MultiBlankQuestionAudio(t *testing.T) {
	ctx := context.Background()
	pool := newMigration0025Pool(t)

	applyMigrationsUpTo(t, pool, "0027_merchandise_product_type.up.sql")

	// Setup: create an exam_topic to reference in questions (0025 introduced this).
	var topicID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO exam_topic (name, subject) VALUES ('algebra', 'math') RETURNING id`,
	).Scan(&topicID))

	// FR-1: pre-0028, the CHECK rejects 'multi_blank'.
	_, err := pool.Exec(ctx,
		`INSERT INTO question (topic_id, format, body, point_correct, point_wrong)
		VALUES ($1, 'multi_blank', 'Test', 1, 0)`,
		topicID,
	)
	require.Error(t, err, "multi_blank must be rejected before the widening migration")

	// Pre-0028: audio_url column does not exist on question.
	var hasAudioURL bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'question' AND column_name = 'audio_url')`,
	).Scan(&hasAudioURL))
	require.False(t, hasAudioURL, "audio_url must not exist before migration 0028")

	// Pre-0028: question_blank table does not exist.
	var hasQuestionBlank bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'question_blank')`,
	).Scan(&hasQuestionBlank))
	require.False(t, hasQuestionBlank, "question_blank table must not exist before migration 0028")

	applyMigrationFile(t, pool, "0028_multi_blank_question_audio.up.sql")

	// FR-1: after up, inserting a multi_blank question succeeds.
	var qID1 uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO question (topic_id, format, body, point_correct, point_wrong)
		VALUES ($1, 'multi_blank', 'Ibu kota adalah {{1}}, tahun {{2}}', 2, 1) RETURNING id`,
		topicID,
	).Scan(&qID1))

	// FR-1: audio_url column now exists and can be set.
	_, err = pool.Exec(ctx,
		`UPDATE question SET audio_url = 'https://example.com/audio.mp3' WHERE id = $1`, qID1,
	)
	require.NoError(t, err, "audio_url must be updateable after migration")

	var audioURL *string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT audio_url FROM question WHERE id = $1`, qID1,
	).Scan(&audioURL))
	require.NotNil(t, audioURL)
	require.Equal(t, "https://example.com/audio.mp3", *audioURL)

	// FR-1: question_blank table now exists with correct schema.
	var hasQuestionBlankAfterUp bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'question_blank')`,
	).Scan(&hasQuestionBlankAfterUp))
	require.True(t, hasQuestionBlankAfterUp, "question_blank table must exist after migration 0028")

	// Insert question_blank rows for the multi_blank question.
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO question_blank (question_id, blank_index, correct_answer)
		VALUES ($1, 1, 'jakarta') RETURNING question_id`,
		qID1,
	).Scan(&qID1))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO question_blank (question_id, blank_index, correct_answer)
		VALUES ($1, 2, '1945') RETURNING question_id`,
		qID1,
	).Scan(&qID1))

	// Verify primary key works: duplicate (question_id, blank_index) is rejected.
	_, err = pool.Exec(ctx,
		`INSERT INTO question_blank (question_id, blank_index, correct_answer)
		VALUES ($1, 1, 'bandung')`,
		qID1,
	)
	require.Error(t, err, "duplicate (question_id, blank_index) must violate PRIMARY KEY")

	// Verify FK CASCADE on question delete.
	_, err = pool.Exec(ctx,
		`DELETE FROM question WHERE id = $1`, qID1,
	)
	require.NoError(t, err)
	var blankCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM question_blank WHERE question_id = $1`, qID1,
	).Scan(&blankCount))
	require.Equal(t, 0, blankCount, "question_blank rows must be deleted when question is deleted (ON DELETE CASCADE)")

	// Control: an unknown format still violates the widened CHECK.
	_, err = pool.Exec(ctx,
		`INSERT INTO question (topic_id, format, body, point_correct, point_wrong)
		VALUES ($1, 'bogus', 'Bogus', 1, 0)`, topicID,
	)
	require.Error(t, err, "an unknown format must still violate the widened CHECK")

	// FR-2: down with pre-existing multi_blank and audio_url data is fail-safe.
	var qID2 uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO question (topic_id, format, body, point_correct, point_wrong, audio_url)
		VALUES ($1, 'multi_blank', 'Another {{1}}', 1, 0, 'https://example.com/other.mp3') RETURNING id`,
		topicID,
	).Scan(&qID2))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO question_blank (question_id, blank_index, correct_answer)
		VALUES ($1, 1, 'answer') RETURNING question_id`,
		qID2,
	).Scan(&qID2))

	applyMigrationFile(t, pool, "0028_multi_blank_question_audio.down.sql")

	// FR-2: multi_blank rows are converted to fill_blank, not dropped.
	var formatAfterDown string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT format FROM question WHERE id = $1`, qID2,
	).Scan(&formatAfterDown))
	require.Equal(t, "fill_blank", formatAfterDown, "down must convert multi_blank rows to fill_blank")

	// FR-2: audio_url column is dropped.
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'question' AND column_name = 'audio_url')`,
	).Scan(&hasAudioURL))
	require.False(t, hasAudioURL, "audio_url column must be dropped by down")

	// FR-2: question_blank table is dropped.
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'question_blank')`,
	).Scan(&hasQuestionBlank))
	require.False(t, hasQuestionBlank, "question_blank table must be dropped by down")

	// FR-2: after down, the narrow CHECK is back: multi_blank is rejected again.
	_, err = pool.Exec(ctx,
		`INSERT INTO question (topic_id, format, body, point_correct, point_wrong)
		VALUES ($1, 'multi_blank', 'Post-down {{1}}', 1, 0)`, topicID,
	)
	require.Error(t, err, "narrow CHECK must be restored after down, rejecting multi_blank")

	// All original 5 formats are still accepted.
	for _, fmt := range []string{"mcq", "multi_answer", "short", "fill_blank", "essay"} {
		_, err = pool.Exec(ctx,
			`INSERT INTO question (topic_id, format, body, point_correct, point_wrong)
			VALUES ($1, $2, $3, 1, 0)`,
			topicID, fmt, fmt+" question",
		)
		require.NoError(t, err, "format %s must be accepted after down", fmt)
	}
}
