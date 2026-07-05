package service

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
)

func TestBuildCredentialsResultCSV(t *testing.T) {
	rows := []StudentBulkResultRow{
		{Name: "Budi", NIS: "1001", Username: "sch_1001", TempPassword: "abc123"},
		{Name: "Siti", NIS: "1002", Error: "student_not_found"},
	}
	data := BuildCredentialsResultCSV(rows)

	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read back csv: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("want 3 records (header + 2 rows), got %d", len(records))
	}
	wantHeader := []string{"name", "nis", "username", "temp_password", "error"}
	for i, h := range wantHeader {
		if records[0][i] != h {
			t.Errorf("header[%d]: want %s, got %s", i, h, records[0][i])
		}
	}
	wantRow1 := []string{"Budi", "1001", "sch_1001", "abc123", ""}
	for i, v := range wantRow1 {
		if records[1][i] != v {
			t.Errorf("row1[%d]: want %s, got %s", i, v, records[1][i])
		}
	}
	wantRow2 := []string{"Siti", "1002", "", "", "student_not_found"}
	for i, v := range wantRow2 {
		if records[2][i] != v {
			t.Errorf("row2[%d]: want %s, got %s", i, v, records[2][i])
		}
	}
}

func TestReissueStudentCredentialsBulk_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	t.Run("explicit ids: success and not-found rows coexist, batch does not abort", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		nis := "b_" + uniqueSuffix()
		reg, err := svc.RegisterStudent(ctx, schoolID, "Budi Reissue", nis, nil, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}

		missingID := "00000000-0000-0000-0000-000000000000"
		data, err := svc.ReissueStudentCredentialsBulk(ctx, schoolID, []string{reg.ID, missingID}, false)
		if err != nil {
			t.Fatalf("ReissueStudentCredentialsBulk: %v", err)
		}

		r := csv.NewReader(strings.NewReader(string(data)))
		records, err := r.ReadAll()
		if err != nil {
			t.Fatalf("read back csv: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("want 3 records (header + 2 rows), got %d", len(records))
		}
		successRow := records[1]
		if successRow[0] != "Budi Reissue" || successRow[1] != nis || successRow[2] == "" || successRow[3] == "" || successRow[4] != "" {
			t.Errorf("unexpected success row: %+v", successRow)
		}
		if successRow[3] == reg.TempPassword {
			t.Error("reissued temp_password should differ from the original registration temp password (proves reissue, not a no-op)")
		}
		notFoundRow := records[2]
		if notFoundRow[4] != "student_not_found" {
			t.Errorf("want error=student_not_found for missing id, got %+v", notFoundRow)
		}
	})

	t.Run("cross-school id is treated as not found", func(t *testing.T) {
		schoolA := createTestSchool(t, svc)
		schoolB := createTestSchool(t, svc)
		reg, err := svc.RegisterStudent(ctx, schoolA, "Cross School", "b_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}

		data, err := svc.ReissueStudentCredentialsBulk(ctx, schoolB, []string{reg.ID}, false)
		if err != nil {
			t.Fatalf("ReissueStudentCredentialsBulk: %v", err)
		}
		r := csv.NewReader(strings.NewReader(string(data)))
		records, err := r.ReadAll()
		if err != nil {
			t.Fatalf("read back csv: %v", err)
		}
		if len(records) != 2 || records[1][4] != "student_not_found" {
			t.Errorf("want single not-found row, got %+v", records)
		}
	})

	t.Run("explicit ids over the row cap error out before any reissue call", func(t *testing.T) {
		ids := make([]string, maxBulkRows+1)
		for i := range ids {
			ids[i] = "00000000-0000-0000-0000-000000000000"
		}
		_, err := svc.ReissueStudentCredentialsBulk(ctx, "some-school", ids, false)
		if !errors.Is(err, ErrRowLimitExceeded) {
			t.Errorf("want ErrRowLimitExceeded, got %v", err)
		}
	})

	t.Run("all=true paginates the whole school", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		var regs []*StudentRegistrationResponse
		for i := 0; i < 3; i++ {
			reg, err := svc.RegisterStudent(ctx, schoolID, "All Student", "b_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
			if err != nil {
				t.Fatalf("RegisterStudent: %v", err)
			}
			regs = append(regs, reg)
		}

		data, err := svc.ReissueStudentCredentialsBulk(ctx, schoolID, nil, true)
		if err != nil {
			t.Fatalf("ReissueStudentCredentialsBulk: %v", err)
		}
		r := csv.NewReader(strings.NewReader(string(data)))
		records, err := r.ReadAll()
		if err != nil {
			t.Fatalf("read back csv: %v", err)
		}
		if len(records) != len(regs)+1 {
			t.Fatalf("want %d records (header + %d rows), got %d", len(regs)+1, len(regs), len(records))
		}
		origTempPasswordByNIS := make(map[string]string, len(regs))
		for _, reg := range regs {
			origTempPasswordByNIS[reg.NIS] = reg.TempPassword
		}
		for _, rec := range records[1:] {
			if rec[2] == "" || rec[3] == "" || rec[4] != "" {
				t.Errorf("unexpected row in all=true batch: %+v", rec)
			}
			if orig, ok := origTempPasswordByNIS[rec[1]]; !ok {
				t.Errorf("row NIS %q not among seeded students", rec[1])
			} else if rec[3] == orig {
				t.Errorf("reissued temp_password for NIS %q should differ from original registration temp password", rec[1])
			}
		}
	})

	t.Run("all=true school over the row cap errors out via pagination, not just explicit ids", func(t *testing.T) {
		origCap := bulkAllRowCap
		bulkAllRowCap = 2
		t.Cleanup(func() { bulkAllRowCap = origCap })

		schoolID := createTestSchool(t, svc)
		for i := 0; i < 3; i++ {
			if _, err := svc.RegisterStudent(ctx, schoolID, "Cap Student", "b_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil); err != nil {
				t.Fatalf("RegisterStudent: %v", err)
			}
		}

		_, err := svc.ReissueStudentCredentialsBulk(ctx, schoolID, nil, true)
		if !errors.Is(err, ErrRowLimitExceeded) {
			t.Errorf("want ErrRowLimitExceeded, got %v", err)
		}
	})
}
