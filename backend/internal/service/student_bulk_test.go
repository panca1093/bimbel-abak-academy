package service

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
)

func TestParseStudentBulkCSV(t *testing.T) {
	t.Run("valid CSV with all columns", func(t *testing.T) {
		data := []byte("name,nis,email\nBudi,1001,budi@example.com\nSiti,1002,\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("want 2 rows, got %d", len(rows))
		}
		if rows[0].Name != "Budi" || rows[0].NIS != "1001" || rows[0].Email == nil || *rows[0].Email != "budi@example.com" {
			t.Errorf("unexpected row 0: %+v", rows[0])
		}
		if rows[1].Email != nil {
			t.Errorf("want nil email for empty column, got %v", *rows[1].Email)
		}
	})

	t.Run("email header case-insensitive and optional", func(t *testing.T) {
		data := []byte("Name,NIS\nBudi,1001\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 1 || rows[0].Name != "Budi" || rows[0].NIS != "1001" {
			t.Errorf("unexpected rows: %+v", rows)
		}
	})

	t.Run("missing name header", func(t *testing.T) {
		data := []byte("nis,email\n1001,a@b.com\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader, got %v", err)
		}
	})

	t.Run("missing nis header", func(t *testing.T) {
		data := []byte("name,email\nBudi,a@b.com\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader, got %v", err)
		}
	})

	t.Run("unparseable bytes returns ErrInvalidCSV", func(t *testing.T) {
		data := []byte("name,nis\n\"Budi,1001\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrInvalidCSV) {
			t.Errorf("want ErrInvalidCSV, got %v", err)
		}
	})

	t.Run("ragged row shorter than header returns ErrInvalidCSV, not a panic", func(t *testing.T) {
		data := []byte("name,nis,email\nBudi\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrInvalidCSV) {
			t.Errorf("want ErrInvalidCSV, got %v", err)
		}
	})

	t.Run("exactly 1000 data rows is fine", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString("name,nis\n")
		for i := 0; i < maxBulkRows; i++ {
			sb.WriteString("Student,1\n")
		}
		rows, err := ParseStudentBulkCSV([]byte(sb.String()))
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != maxBulkRows {
			t.Errorf("want %d rows, got %d", maxBulkRows, len(rows))
		}
	})

	t.Run("1001 data rows exceeds limit", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString("name,nis\n")
		for i := 0; i < maxBulkRows+1; i++ {
			sb.WriteString("Student,1\n")
		}
		_, err := ParseStudentBulkCSV([]byte(sb.String()))
		if !errors.Is(err, ErrRowLimitExceeded) {
			t.Errorf("want ErrRowLimitExceeded, got %v", err)
		}
	})
}

func TestBuildStudentBulkResultCSV(t *testing.T) {
	results := []StudentBulkResultRow{
		{Name: "Budi", NIS: "1001", Email: "budi@example.com", Status: "success", Username: "sch_1001", TempPassword: "abc123"},
		{Name: "Siti", NIS: "1002", Status: "failed", Error: "nis already registered in this school"},
	}
	data := BuildStudentBulkResultCSV(results)

	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read back csv: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("want 3 records (header + 2 rows), got %d", len(records))
	}
	wantHeader := []string{"name", "nis", "email", "status", "username", "temp_password", "error"}
	for i, h := range wantHeader {
		if records[0][i] != h {
			t.Errorf("header[%d]: want %s, got %s", i, h, records[0][i])
		}
	}
	wantRow1 := []string{"Budi", "1001", "budi@example.com", "success", "sch_1001", "abc123", ""}
	for i, v := range wantRow1 {
		if records[1][i] != v {
			t.Errorf("row1[%d]: want %s, got %s", i, v, records[1][i])
		}
	}
	wantRow2 := []string{"Siti", "1002", "", "failed", "", "", "nis already registered in this school"}
	for i, v := range wantRow2 {
		if records[2][i] != v {
			t.Errorf("row2[%d]: want %s, got %s", i, v, records[2][i])
		}
	}
}

func TestProcessStudentBulkRows_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	t.Run("all-success batch", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		rows := []StudentBulkRow{
			{Name: "Budi", NIS: "b_" + uniqueSuffix()},
			{Name: "Siti", NIS: "b_" + uniqueSuffix()},
		}
		var progressCalls []int
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, func(pct int) {
			progressCalls = append(progressCalls, pct)
		})
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != len(rows) {
			t.Errorf("want successCount=%d, got %d", len(rows), successCount)
		}
		for _, r := range results {
			if r.Status != "success" || r.Username == "" || r.TempPassword == "" || r.Error != "" {
				t.Errorf("unexpected result row: %+v", r)
			}
		}
		if len(progressCalls) == 0 || progressCalls[len(progressCalls)-1] != 100 {
			t.Errorf("want progress calls ending at 100, got %v", progressCalls)
		}
	})

	t.Run("duplicate NIS row fails, others succeed", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		dupNIS := "b_" + uniqueSuffix()
		if _, err := svc.RegisterStudent(ctx, schoolID, "Existing", dupNIS, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("seed RegisterStudent: %v", err)
		}
		rows := []StudentBulkRow{
			{Name: "New Student", NIS: dupNIS},
			{Name: "Another", NIS: "b_" + uniqueSuffix()},
		}
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, func(int) {})
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != 1 {
			t.Errorf("want successCount=1, got %d", successCount)
		}
		if results[0].Status != "failed" || results[0].Error != ErrDuplicateNIS.Error() {
			t.Errorf("want failed row with ErrDuplicateNIS text, got %+v", results[0])
		}
		if results[1].Status != "success" {
			t.Errorf("want second row to succeed, got %+v", results[1])
		}
	})

	t.Run("deactivated school: every row fails, successCount 0", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		if _, err := svc.ChangeSchoolStatus(ctx, schoolID, "deactivated"); err != nil {
			t.Fatalf("ChangeSchoolStatus: %v", err)
		}
		rows := []StudentBulkRow{
			{Name: "A", NIS: "b_" + uniqueSuffix()},
			{Name: "B", NIS: "b_" + uniqueSuffix()},
		}
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, func(int) {})
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != 0 {
			t.Errorf("want successCount=0, got %d", successCount)
		}
		for _, r := range results {
			if r.Status != "failed" || r.Error != ErrSchoolDeactivated.Error() {
				t.Errorf("want every row failed with ErrSchoolDeactivated, got %+v", r)
			}
		}
	})
}
