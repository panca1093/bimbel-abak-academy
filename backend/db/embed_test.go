package dbmigrations

import (
	"io/fs"
	"sort"
	"strings"
	"testing"
)

func TestEmbed_contains_system_config_migration(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/0012_system_config.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("0012_system_config.up.sql not found in embedded FS")
	}
}

func TestEmbed_0012_sorts_after_0011(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/*.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	sort.Strings(entries)

	var found bool
	for _, e := range entries {
		if e == "migrations/0011_profile_photo.up.sql" {
			found = true
		}
		if e == "migrations/0012_system_config.up.sql" {
			if !found {
				t.Fatal("0012_system_config.up.sql appears before 0011_profile_photo.up.sql in sorted order")
			}
			return
		}
	}
	t.Fatal("0012_system_config.up.sql not found in sorted migration list after 0011")
}

func TestEmbed_contains_exam_migration(t *testing.T) {
	for _, p := range []string{
		"migrations/0014_exam.up.sql",
		"migrations/0014_exam.down.sql",
	} {
		entries, err := fs.Glob(FS, p)
		if err != nil {
			t.Fatalf("Glob %s: %v", p, err)
		}
		if len(entries) == 0 {
			t.Fatalf("%s not found in embedded FS", p)
		}
	}
}

func TestEmbed_0014_sorts_after_0013(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/*.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	sort.Strings(entries)

	var found bool
	for _, e := range entries {
		if e == "migrations/0013_shipped_status.up.sql" {
			found = true
		}
		if e == "migrations/0014_exam.up.sql" {
			if !found {
				t.Fatal("0014_exam.up.sql appears before 0013_shipped_status.up.sql in sorted order")
			}
			return
		}
	}
	t.Fatal("0014_exam.up.sql not found in sorted migration list after 0013")
}

func TestEmbed_0014_up_creates_all_nine_tables(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0014_exam.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	tables := []string{
		"CREATE TABLE IF NOT EXISTS test",
		"CREATE TABLE IF NOT EXISTS question",
		"CREATE TABLE IF NOT EXISTS question_option",
		"CREATE TABLE IF NOT EXISTS exam",
		"CREATE TABLE IF NOT EXISTS exam_test",
		"CREATE TABLE IF NOT EXISTS exam_registration",
		"CREATE TABLE IF NOT EXISTS exam_session",
		"CREATE TABLE IF NOT EXISTS exam_session_answer",
		"CREATE TABLE IF NOT EXISTS session_violation_log",
	}
	for _, stmt := range tables {
		if !strings.Contains(string(body), stmt) {
			t.Errorf("0014_exam.up.sql missing expected statement: %s", stmt)
		}
	}
}

func TestEmbed_0014_up_enforces_question_format_check(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0014_exam.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(body)
	if !strings.Contains(s, "CHECK") || !strings.Contains(s, "format IN") {
		t.Fatal("expected CHECK (... format IN ...) on question in 0014_exam.up.sql")
	}
	for _, v := range []string{"mcq", "multi_answer", "short", "fill_blank", "essay"} {
		if !strings.Contains(s, v) {
			t.Errorf("format CHECK missing value %q", v)
		}
	}
}

func TestEmbed_0014_up_cascades_authoring_deletes(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0014_exam.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(body)
	if !strings.Contains(s, "ON DELETE CASCADE") {
		t.Fatal("expected ON DELETE CASCADE in 0014_exam.up.sql")
	}
	if !strings.Contains(s, "uq_question_order") {
		t.Errorf("expected named unique index uq_question_order on (test_id, sort_order)")
	}
}

func TestEmbed_0016_contains_certificate_migration(t *testing.T) {
	for _, p := range []string{
		"migrations/0016_exam_certificate.up.sql",
		"migrations/0016_exam_certificate.down.sql",
	} {
		entries, err := fs.Glob(FS, p)
		if err != nil {
			t.Fatalf("Glob %s: %v", p, err)
		}
		if len(entries) == 0 {
			t.Fatalf("%s not found in embedded FS", p)
		}
	}
}

func TestEmbed_0016_sorts_after_0015(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/*.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	sort.Strings(entries)

	var found bool
	for _, e := range entries {
		if e == "migrations/0015_exam_scoring.up.sql" {
			found = true
		}
		if e == "migrations/0016_exam_certificate.up.sql" {
			if !found {
				t.Fatal("0016_exam_certificate.up.sql appears before 0015_exam_scoring.up.sql in sorted order")
			}
			return
		}
	}
	t.Fatal("0016_exam_certificate.up.sql not found in sorted migration list after 0015")
}

func TestEmbed_0016_up_adds_certificate_columns(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0016_exam_certificate.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(body)

	if !strings.Contains(s, "ALTER TABLE exam ADD COLUMN IF NOT EXISTS certificate_template") {
		t.Error("up.sql missing ALTER TABLE exam ADD COLUMN IF NOT EXISTS certificate_template")
	}
	if !strings.Contains(s, "ALTER TABLE exam_session ADD COLUMN IF NOT EXISTS certificate_generated_at") {
		t.Error("up.sql missing ALTER TABLE exam_session ADD COLUMN IF NOT EXISTS certificate_generated_at")
	}
	if !strings.Contains(s, "TEXT NOT NULL DEFAULT 'classic'") {
		t.Error("certificate_template missing TEXT NOT NULL DEFAULT 'classic'")
	}
	if !strings.Contains(s, "TIMESTAMPTZ") {
		t.Error("certificate_generated_at missing TIMESTAMPTZ type")
	}
}

func TestEmbed_0016_up_enforces_certificate_template_check(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0016_exam_certificate.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(body)

	if !strings.Contains(s, "chk_certificate_template") {
		t.Error("up.sql missing named constraint chk_certificate_template")
	}
	if !strings.Contains(s, "DROP CONSTRAINT IF EXISTS chk_certificate_template") {
		t.Error("up.sql missing DROP CONSTRAINT IF EXISTS chk_certificate_template before ADD CONSTRAINT")
	}
	for _, v := range []string{"classic", "modern", "elegant"} {
		if !strings.Contains(s, v) {
			t.Errorf("chk_certificate_template CHECK missing value %q", v)
		}
	}
}

func TestEmbed_0016_down_drops_certificate_columns(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0016_exam_certificate.down.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(body)

	if !strings.Contains(s, "DROP CONSTRAINT IF EXISTS chk_certificate_template") {
		t.Error("down.sql missing DROP CONSTRAINT IF EXISTS chk_certificate_template")
	}
	if !strings.Contains(s, "DROP COLUMN IF EXISTS certificate_template") {
		t.Error("down.sql missing DROP COLUMN IF EXISTS certificate_template")
	}
	if !strings.Contains(s, "DROP COLUMN IF EXISTS certificate_generated_at") {
		t.Error("down.sql missing DROP COLUMN IF EXISTS certificate_generated_at")
	}
}

func TestEmbed_0014_down_drops_in_reverse_dependency_order(t *testing.T) {
	body, err := fs.ReadFile(FS, "migrations/0014_exam.down.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Reverse-dependency order for drops: leaves first, parents last.
	// Match the full DROP statement (terminated by ';') so substrings like
	// `exam_session` don't match `exam_session_answer` first.
	wantOrder := []string{
		"DROP TABLE IF EXISTS session_violation_log;",
		"DROP TABLE IF EXISTS exam_session_answer;",
		"DROP TABLE IF EXISTS exam_session;",
		"DROP TABLE IF EXISTS exam_registration;",
		"DROP TABLE IF EXISTS exam_test;",
		"DROP TABLE IF EXISTS exam;",
		"DROP TABLE IF EXISTS question_option;",
		"DROP TABLE IF EXISTS question;",
		"DROP TABLE IF EXISTS test;",
	}
	lastOffset := -1
	for _, stmt := range wantOrder {
		idx := strings.Index(string(body), stmt)
		if idx < 0 {
			t.Errorf("down file missing %q", stmt)
			continue
		}
		if idx <= lastOffset {
			t.Errorf("down file out of reverse-dependency order: %q at offset %d follows prior at %d", stmt, idx, lastOffset)
		}
		lastOffset = idx
	}
}
