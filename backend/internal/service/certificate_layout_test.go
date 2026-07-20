package service

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateLayout_UnknownFieldIDRejected(t *testing.T) {
	l := Layout{
		Page: Page{WidthMm: 297, HeightMm: 210},
		Fields: []LayoutField{
			{ID: "not_a_real_field", XMm: 10, YMm: 10, WMm: 50, Align: "left"},
		},
	}
	err := ValidateLayout(l)
	if err == nil {
		t.Fatal("expected error for unknown field id, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestValidateLayout_DuplicateFieldIDRejected(t *testing.T) {
	l := Layout{
		Page: Page{WidthMm: 297, HeightMm: 210},
		Fields: []LayoutField{
			{ID: "title", XMm: 10, YMm: 10, WMm: 50, Align: "left"},
			{ID: "title", XMm: 20, YMm: 20, WMm: 50, Align: "left"},
		},
	}
	err := ValidateLayout(l)
	if err == nil {
		t.Fatal("expected error for duplicate field id, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestValidateLayout_OutOfPageBoxRejected(t *testing.T) {
	cases := []struct {
		name  string
		field LayoutField
	}{
		{"negative x", LayoutField{ID: "title", XMm: -5, YMm: 10, WMm: 50, Align: "left"}},
		{"negative y", LayoutField{ID: "title", XMm: 10, YMm: -5, WMm: 50, Align: "left"}},
		{"right edge overflow", LayoutField{ID: "title", XMm: 280, YMm: 10, WMm: 50, Align: "left"}},
		{"bottom edge overflow", LayoutField{ID: "title", XMm: 10, YMm: 250, WMm: 50, Align: "left"}},
		{"logo height overflow", LayoutField{ID: "logo", XMm: 10, YMm: 200, WMm: 20, Align: "left", HMm: 30}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := Layout{
				Page:   Page{WidthMm: 297, HeightMm: 210},
				Fields: []LayoutField{tc.field},
			}
			err := ValidateLayout(l)
			if err == nil {
				t.Fatalf("expected error for out-of-page box (%s), got nil", tc.name)
			}
			if !errors.Is(err, ErrValidation) {
				t.Errorf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateLayout_InPageBoxAccepted(t *testing.T) {
	l := Layout{
		Page: Page{WidthMm: 297, HeightMm: 210},
		Fields: []LayoutField{
			{ID: "title", XMm: 10, YMm: 10, WMm: 50, Align: "left"},
			{ID: "logo", XMm: 10, YMm: 10, WMm: 20, Align: "left", HMm: 20},
		},
	}
	if err := ValidateLayout(l); err != nil {
		t.Errorf("expected in-page boxes to be accepted, got %v", err)
	}
}

func TestDefaultLayout_RoundTripsThroughJSON(t *testing.T) {
	for _, tmpl := range []string{"classic", "modern", "elegant"} {
		t.Run(tmpl, func(t *testing.T) {
			original := defaultLayout(tmpl)

			if err := ValidateLayout(original); err != nil {
				t.Fatalf("default layout for %s failed validation: %v", tmpl, err)
			}

			b, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var roundTripped Layout
			if err := json.Unmarshal(b, &roundTripped); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			b2, err := json.Marshal(roundTripped)
			if err != nil {
				t.Fatalf("second Marshal failed: %v", err)
			}

			if string(b) != string(b2) {
				t.Errorf("layout for %s did not round-trip unchanged\nfirst:  %s\nsecond: %s", tmpl, b, b2)
			}
		})
	}
}

func TestDefaultLayout_CoversAllClosedFieldIDs(t *testing.T) {
	want := []string{
		"title", "subtitle", "student_name", "exam_title",
		"completion_text", "date", "certificate_number", "logo",
	}
	for _, tmpl := range []string{"classic", "modern", "elegant"} {
		l := defaultLayout(tmpl)
		seen := make(map[string]bool, len(l.Fields))
		for _, f := range l.Fields {
			seen[f.ID] = true
		}
		for _, id := range want {
			if !seen[id] {
				t.Errorf("defaultLayout(%s) missing field id %q", tmpl, id)
			}
		}
	}
}
