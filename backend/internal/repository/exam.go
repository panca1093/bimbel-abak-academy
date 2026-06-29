package repository

import (
	"akademi-bimbel/internal/model"
)

// Slice 0 skeleton — CRUD methods land in Task 4.

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