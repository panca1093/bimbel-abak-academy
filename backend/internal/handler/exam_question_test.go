package handler

import (
	"errors"
	"strings"
	"testing"

	"akademi-bimbel/internal/service"
)

func TestQuestionRequest_toQuestion_defaultsPoints(t *testing.T) {
	req := questionRequest{Format: "essay", Body: "explain gravity"}
	q, err := req.toQuestion()
	if err != nil {
		t.Fatalf("toQuestion returned error: %v", err)
	}
	if q.PointCorrect != 1 {
		t.Errorf("PointCorrect default = %d, want 1", q.PointCorrect)
	}
	if q.PointWrong != 0 {
		t.Errorf("PointWrong default = %d, want 0", q.PointWrong)
	}
}

func TestQuestionRequest_toQuestion_appliesExplicitPoints(t *testing.T) {
	pc, pw := 3.0, 2.0
	req := questionRequest{Format: "essay", Body: "explain gravity", PointCorrect: &pc, PointWrong: &pw}
	q, err := req.toQuestion()
	if err != nil {
		t.Fatalf("toQuestion returned error: %v", err)
	}
	if q.PointCorrect != 3 {
		t.Errorf("PointCorrect = %d, want 3", q.PointCorrect)
	}
	if q.PointWrong != 2 {
		t.Errorf("PointWrong = %d, want 2", q.PointWrong)
	}
}

func TestQuestionRequest_toQuestion_rejectsFractionalPointCorrect(t *testing.T) {
	pc := 1.5
	req := questionRequest{Format: "essay", Body: "explain gravity", PointCorrect: &pc}
	_, err := req.toQuestion()
	if err == nil {
		t.Fatal("fractional point_correct should return error")
	}
	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("fractional point_correct should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "point_correct must be an integer") {
		t.Errorf("msg should mention 'point_correct must be an integer', got %q", err.Error())
	}
}

func TestQuestionRequest_toQuestion_rejectsFractionalPointWrong(t *testing.T) {
	pw := 0.5
	req := questionRequest{Format: "essay", Body: "explain gravity", PointWrong: &pw}
	_, err := req.toQuestion()
	if err == nil {
		t.Fatal("fractional point_wrong should return error")
	}
	if !errors.Is(err, service.ErrValidation) {
		t.Errorf("fractional point_wrong should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "point_wrong must be an integer") {
		t.Errorf("msg should mention 'point_wrong must be an integer', got %q", err.Error())
	}
}
