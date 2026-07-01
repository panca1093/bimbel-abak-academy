package service

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

// gradeAnswer determines correctness per format rules (FR20, R3).
// For mcq the correct key is derived from options (is_correct=true).
// For multi_answer the correct keys are derived from options.
// For short/fill_blank the correct_answer field is used (trim+lower).
// For essay or unknown formats returns false (not auto-graded).
func gradeAnswer(format string, answer *string, correctAnswer *string, options []model.QuestionOption) bool {
	if answer == nil || *answer == "" {
		return false
	}

	switch format {
	case "mcq":
		return gradeMCQ(*answer, options)
	case "multi_answer":
		return gradeMultiAnswer(*answer, options)
	case "short", "fill_blank":
		if correctAnswer == nil {
			return false
		}
		return strings.EqualFold(
			strings.TrimSpace(*answer),
			strings.TrimSpace(*correctAnswer),
		)
	default:
		return false
	}
}

func gradeMCQ(answer string, options []model.QuestionOption) bool {
	var correctKey string
	for _, o := range options {
		if o.IsCorrect {
			correctKey = o.Key
			break
		}
	}
	if correctKey == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(correctKey))
}

func gradeMultiAnswer(answer string, options []model.QuestionOption) bool {
	selected := strings.Split(answer, ",")
	for i := range selected {
		selected[i] = strings.TrimSpace(selected[i])
	}
	sort.Strings(selected)

	var correct []string
	for _, o := range options {
		if o.IsCorrect {
			correct = append(correct, o.Key)
		}
	}
	sort.Strings(correct)

	if len(selected) != len(correct) {
		return false
	}
	for i := range selected {
		if selected[i] != correct[i] {
			return false
		}
	}
	return true
}

// gradeObjective grades all objective questions per R4 score normalization.
// Essay questions are returned with is_correct=nil, score=nil (not auto-graded).
//
// N = count of objective questions (mcq, multi_answer, short, fill_blank).
// Per-correct objective answer scores 100.0/N; incorrect/missing scores 0.
// If N = 0 the total score is 0.
func gradeObjective(questions []model.QuestionWithOptions, answers map[uuid.UUID]*string) ([]model.ExamSessionAnswer, float64) {
	n := 0
	for _, q := range questions {
		if q.Question.Format != "essay" {
			n++
		}
	}

	if n == 0 {
		graded := make([]model.ExamSessionAnswer, len(questions))
		for i, q := range questions {
			ans := answers[q.Question.ID]
			graded[i] = model.ExamSessionAnswer{
				QuestionID: q.Question.ID,
				Answer:     ans,
				IsCorrect:  nil,
				Score:      nil,
			}
		}
		return graded, 0
	}

	perCorrect := 100.0 / float64(n)
	graded := make([]model.ExamSessionAnswer, 0, len(questions))
	var totalScore float64

	for _, q := range questions {
		if q.Question.Format == "essay" {
			ans := answers[q.Question.ID]
			graded = append(graded, model.ExamSessionAnswer{
				QuestionID: q.Question.ID,
				Answer:     ans,
				IsCorrect:  nil,
				Score:      nil,
			})
			continue
		}

		ans := answers[q.Question.ID]
		correct := gradeAnswer(q.Question.Format, ans, q.Question.CorrectAnswer, q.Options)

		var answerScore float64
		if correct {
			answerScore = perCorrect
		}
		isCorrect := correct
		graded = append(graded, model.ExamSessionAnswer{
			QuestionID: q.Question.ID,
			Answer:     ans,
			IsCorrect:  &isCorrect,
			Score:      &answerScore,
		})
		totalScore += answerScore
	}

	return graded, totalScore
}
