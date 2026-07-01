package handler

import "testing"

func TestQuestionRequest_toQuestion_defaultsPoints(t *testing.T) {
	req := questionRequest{Format: "essay", Body: "explain gravity"}
	q := req.toQuestion()
	if q.PointCorrect != 1 {
		t.Errorf("PointCorrect default = %d, want 1", q.PointCorrect)
	}
	if q.PointWrong != 0 {
		t.Errorf("PointWrong default = %d, want 0", q.PointWrong)
	}
}

func TestQuestionRequest_toQuestion_appliesExplicitPoints(t *testing.T) {
	pc, pw := 3, 2
	req := questionRequest{Format: "essay", Body: "explain gravity", PointCorrect: &pc, PointWrong: &pw}
	q := req.toQuestion()
	if q.PointCorrect != 3 {
		t.Errorf("PointCorrect = %d, want 3", q.PointCorrect)
	}
	if q.PointWrong != 2 {
		t.Errorf("PointWrong = %d, want 2", q.PointWrong)
	}
}
