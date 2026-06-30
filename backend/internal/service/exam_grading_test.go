package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"akademi-bimbel/internal/model"
)

func TestGrading_mcq_caseInsensitive(t *testing.T) {
	options := []model.QuestionOption{
		{Key: "a", Text: "Option A", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "Option B", SortOrder: 2},
	}

	tests := []struct {
		name   string
		answer *string
		want   bool
	}{
		{"exact match", strPtr("a"), true},
		{"case-insensitive", strPtr("A"), true},
		{"wrong answer", strPtr("b"), false},
		{"nil answer", nil, false},
		{"empty answer", strPtr(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradeAnswer("mcq", tt.answer, nil, options)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrading_mcq_trim(t *testing.T) {
	options := []model.QuestionOption{
		{Key: "a", Text: "Option A", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "Option B", SortOrder: 2},
	}

	got := gradeAnswer("mcq", strPtr("  a  "), nil, options)
	assert.True(t, got, "mcq should trim whitespace")
}

func TestGrading_mcq_noCorrectOption(t *testing.T) {
	got := gradeAnswer("mcq", strPtr("a"), nil, nil)
	assert.False(t, got, "mcq with no options should return false")
}

func TestGrading_multiAnswer_setEquality(t *testing.T) {
	options := []model.QuestionOption{
		{Key: "a", Text: "A", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "B", IsCorrect: true, SortOrder: 2},
		{Key: "c", Text: "C", IsCorrect: true, SortOrder: 3},
		{Key: "d", Text: "D", SortOrder: 4},
	}

	tests := []struct {
		name   string
		answer *string
		want   bool
	}{
		{"all correct order a,b,c", strPtr("a,b,c"), true},
		{"all correct order c,b,a", strPtr("c,b,a"), true},
		{"all correct with spaces", strPtr("  a , b , c  "), true},
		{"missing one option", strPtr("a,b"), false},
		{"extra option", strPtr("a,b,c,d"), false},
		{"empty answer", strPtr(""), false},
		{"nil answer", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradeAnswer("multi_answer", tt.answer, nil, options)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrading_short_trimLower(t *testing.T) {
	tests := []struct {
		name          string
		answer        *string
		correctAnswer *string
		want          bool
	}{
		{"exact match", strPtr("hello"), strPtr("hello"), true},
		{"case-insensitive", strPtr("Hello"), strPtr("hello"), true},
		{"trimmed", strPtr("  hello  "), strPtr("hello"), true},
		{"wrong answer", strPtr("world"), strPtr("hello"), false},
		{"nil answer", nil, strPtr("hello"), false},
		{"empty answer", strPtr(""), strPtr("hello"), false},
		{"nil correctAnswer", strPtr("hello"), nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradeAnswer("short", tt.answer, tt.correctAnswer, nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrading_fillBlank_trimLower(t *testing.T) {
	tests := []struct {
		name          string
		answer        *string
		correctAnswer *string
		want          bool
	}{
		{"exact match", strPtr("jakarta"), strPtr("jakarta"), true},
		{"case-insensitive", strPtr("Jakarta"), strPtr("jakarta"), true},
		{"wrong answer", strPtr("bandung"), strPtr("jakarta"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradeAnswer("fill_blank", tt.answer, tt.correctAnswer, nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrading_essay_returnsFalse(t *testing.T) {
	got := gradeAnswer("essay", strPtr("some long essay"), nil, nil)
	assert.False(t, got, "essay should not be auto-graded")
}

func TestGrading_unknownFormat(t *testing.T) {
	got := gradeAnswer("unknown_format", strPtr("anything"), nil, nil)
	assert.False(t, got, "unknown format should return false")
}

func TestGradingObjective_allEssays_scoreZero(t *testing.T) {
	q1 := model.Question{ID: uuid.New(), Format: "essay", Body: "Essay 1"}
	q2 := model.Question{ID: uuid.New(), Format: "essay", Body: "Essay 2"}
	questions := []model.QuestionWithOptions{
		{Question: q1},
		{Question: q2},
	}
	answers := map[uuid.UUID]*string{
		q1.ID: strPtr("My essay answer 1"),
		q2.ID: strPtr("My essay answer 2"),
	}

	graded, score := gradeObjective(questions, answers)

	assert.Equal(t, 0.0, score, "all-essay exam should score 0")
	assert.Len(t, graded, 2)
	for _, g := range graded {
		assert.Nil(t, g.IsCorrect, "essay IsCorrect should be nil")
		assert.Nil(t, g.Score, "essay Score should be nil")
	}
}

func TestGradingObjective_nIsZero_scoreZero(t *testing.T) {
	questions := []model.QuestionWithOptions{}
	answers := map[uuid.UUID]*string{}

	graded, score := gradeObjective(questions, answers)

	assert.Equal(t, 0.0, score)
	assert.Empty(t, graded)
}

func TestGradingObjective_allCorrect_score100(t *testing.T) {
	q1 := model.Question{ID: uuid.New(), Format: "mcq", Body: "2+2"}
	q2 := model.Question{ID: uuid.New(), Format: "short", Body: "Capital of France", CorrectAnswer: strPtr("Paris")}
	questions := []model.QuestionWithOptions{
		{
			Question: q1,
			Options: []model.QuestionOption{
				{Key: "a", Text: "3", SortOrder: 1},
				{Key: "b", Text: "4", IsCorrect: true, SortOrder: 2},
			},
		},
		{
			Question: q2,
		},
	}
	answers := map[uuid.UUID]*string{
		q1.ID: strPtr("b"),
		q2.ID: strPtr("Paris"),
	}

	graded, score := gradeObjective(questions, answers)

	assert.InDelta(t, 100.0, score, 0.005)
	assert.Len(t, graded, 2)
	for _, g := range graded {
		assert.True(t, *g.IsCorrect)
		assert.NotNil(t, g.Score)
	}
}

func TestGradingObjective_partial_proportional(t *testing.T) {
	q1 := model.Question{ID: uuid.New(), Format: "mcq", Body: "2+2"}
	q2 := model.Question{ID: uuid.New(), Format: "short", Body: "Capital of France", CorrectAnswer: strPtr("Paris")}
	q3 := model.Question{ID: uuid.New(), Format: "fill_blank", Body: "___ is the largest planet", CorrectAnswer: strPtr("Jupiter")}
	questions := []model.QuestionWithOptions{
		{
			Question: q1,
			Options: []model.QuestionOption{
				{Key: "a", Text: "3", SortOrder: 1},
				{Key: "b", Text: "4", IsCorrect: true, SortOrder: 2},
			},
		},
		{
			Question: q2,
		},
		{
			Question: q3,
		},
	}
	// Only q1 correct, q2 wrong, q3 missing
	answers := map[uuid.UUID]*string{
		q1.ID: strPtr("b"),
		q2.ID: strPtr("London"),
		// q3 missing -> scores 0
	}

	graded, score := gradeObjective(questions, answers)

	// N=3, perCorrect = 100/3 ≈ 33.33, 1 correct = ~33.33
	assert.InDelta(t, 100.0/3.0, score, 0.005)
	assert.Len(t, graded, 3)

	// q1 correct
	assert.True(t, *graded[0].IsCorrect)
	assert.InDelta(t, 100.0/3.0, *graded[0].Score, 0.005)
	// q2 wrong
	assert.False(t, *graded[1].IsCorrect)
	assert.Equal(t, 0.0, *graded[1].Score)
	// q3 missing -> wrong (0 score)
	assert.False(t, *graded[2].IsCorrect)
	assert.Equal(t, 0.0, *graded[2].Score)
}

func TestGradingObjective_mixedWithEssay(t *testing.T) {
	qEssay := model.Question{ID: uuid.New(), Format: "essay", Body: "Write about X"}
	qMcq := model.Question{ID: uuid.New(), Format: "mcq", Body: "2+2"}
	questions := []model.QuestionWithOptions{
		{
			Question: qEssay,
		},
		{
			Question: qMcq,
			Options: []model.QuestionOption{
				{Key: "a", Text: "3", SortOrder: 1},
				{Key: "b", Text: "4", IsCorrect: true, SortOrder: 2},
			},
		},
	}
	answers := map[uuid.UUID]*string{
		qEssay.ID: strPtr("My essay"),
		qMcq.ID:  strPtr("b"),
	}

	// N=1 (only mcq), perCorrect = 100.0
	graded, score := gradeObjective(questions, answers)

	assert.InDelta(t, 100.0, score, 0.005)
	assert.Len(t, graded, 2)

	// essay: ungraded
	assert.Nil(t, graded[0].IsCorrect)
	assert.Nil(t, graded[0].Score)
	assert.Equal(t, qEssay.ID, graded[0].QuestionID)

	// mcq: correct
	assert.True(t, *graded[1].IsCorrect)
	assert.InDelta(t, 100.0, *graded[1].Score, 0.005)
	assert.Equal(t, qMcq.ID, graded[1].QuestionID)
}

func TestGradingObjective_missingAnswerScoresZero(t *testing.T) {
	q := model.Question{ID: uuid.New(), Format: "short", Body: "Capital", CorrectAnswer: strPtr("Paris")}
	questions := []model.QuestionWithOptions{
		{Question: q},
	}
	// Missing answer for the question
	answers := map[uuid.UUID]*string{}

	graded, score := gradeObjective(questions, answers)

	assert.Equal(t, 0.0, score)
	assert.Len(t, graded, 1)
	assert.False(t, *graded[0].IsCorrect)
	assert.Equal(t, 0.0, *graded[0].Score)
}

func TestGradingObjective_multiAnswerCorrect(t *testing.T) {
	qID := uuid.New()
	questions := []model.QuestionWithOptions{
		{
			Question: model.Question{ID: qID, Format: "multi_answer", Body: "Select A B C"},
			Options: []model.QuestionOption{
				{Key: "a", Text: "A", IsCorrect: true, SortOrder: 1},
				{Key: "b", Text: "B", IsCorrect: true, SortOrder: 2},
				{Key: "c", Text: "C", IsCorrect: true, SortOrder: 3},
				{Key: "d", Text: "D", SortOrder: 4},
			},
		},
	}
	answers := map[uuid.UUID]*string{
		qID: strPtr("c,a,b"),
	}

	graded, score := gradeObjective(questions, answers)

	assert.InDelta(t, 100.0, score, 0.005)
	assert.Len(t, graded, 1)
	assert.True(t, *graded[0].IsCorrect)
}
