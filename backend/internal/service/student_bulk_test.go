package service

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"akademi-bimbel/internal/repository"
)

func TestParseStudentBulkCSV(t *testing.T) {
	t.Run("valid CSV with jenjang and email", func(t *testing.T) {
		data := []byte("name,jenjang,email\nBudi,sma,budi@example.com\nSiti,sma,\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("want 2 rows, got %d", len(rows))
		}
		if rows[0].Name != "Budi" || rows[0].Jenjang != "sma" || rows[0].Email == nil || *rows[0].Email != "budi@example.com" {
			t.Errorf("unexpected row 0: %+v", rows[0])
		}
		if rows[1].Email != nil {
			t.Errorf("want nil email for empty column, got %v", *rows[1].Email)
		}
	})

	t.Run("jenjang-only CSV (no email)", func(t *testing.T) {
		data := []byte("name,jenjang\nBudi,sma\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 1 || rows[0].Name != "Budi" || rows[0].Jenjang != "sma" {
			t.Errorf("unexpected rows: %+v", rows)
		}
	})

	t.Run("name-only CSV missing jenjang returns error", func(t *testing.T) {
		data := []byte("name,email\nBudi,a@b.com\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader, got %v", err)
		}
	})

	t.Run("nis header ignored, jenjang still required", func(t *testing.T) {
		data := []byte("name,nis,email\nBudi,1001,budi@example.com\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader (jenjang missing), got %v", err)
		}
	})

	t.Run("nis header present with jenjang is ignored", func(t *testing.T) {
		data := []byte("name,jenjang,nis,email\nBudi,sma,1001,budi@example.com\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 1 || rows[0].Name != "Budi" || rows[0].Jenjang != "sma" {
			t.Errorf("unexpected rows: %+v", rows)
		}
	})

	t.Run("optional address columns parsed when present", func(t *testing.T) {
		data := []byte("name,jenjang,email,provinsi,kota,kecamatan,kode_pos\nBudi,sma,b@b.com,Jawa Barat,Bandung,Coblong,40131\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("want 1 row, got %d", len(rows))
		}
		if rows[0].Name != "Budi" || rows[0].Jenjang != "sma" {
			t.Errorf("unexpected name/jenjang: %+v", rows[0])
		}
		if rows[0].Provinsi == nil || *rows[0].Provinsi != "Jawa Barat" {
			t.Errorf("want provinsi 'Jawa Barat', got %v", rows[0].Provinsi)
		}
		if rows[0].Kota == nil || *rows[0].Kota != "Bandung" {
			t.Errorf("want kota 'Bandung', got %v", rows[0].Kota)
		}
		if rows[0].Kecamatan == nil || *rows[0].Kecamatan != "Coblong" {
			t.Errorf("want kecamatan 'Coblong', got %v", rows[0].Kecamatan)
		}
		if rows[0].KodePos == nil || *rows[0].KodePos != "40131" {
			t.Errorf("want kode_pos '40131', got %v", rows[0].KodePos)
		}
	})

	t.Run("optional address columns absent not an error", func(t *testing.T) {
		data := []byte("name,jenjang\nBudi,sma\n")
		rows, err := ParseStudentBulkCSV(data)
		if err != nil {
			t.Fatalf("ParseStudentBulkCSV: %v", err)
		}
		if len(rows) != 1 || rows[0].Name != "Budi" || rows[0].Jenjang != "sma" {
			t.Errorf("unexpected rows: %+v", rows)
		}
		if rows[0].Provinsi != nil || rows[0].Kota != nil || rows[0].Kecamatan != nil || rows[0].KodePos != nil {
			t.Errorf("optional address fields should be nil when columns absent, got %+v", rows[0])
		}
	})

	t.Run("missing name header", func(t *testing.T) {
		data := []byte("jenjang,email\nsma,a@b.com\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader, got %v", err)
		}
	})

	t.Run("unparseable bytes returns ErrInvalidCSV", func(t *testing.T) {
		data := []byte("name,jenjang\n\"Budi,sma\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrInvalidCSV) {
			t.Errorf("want ErrInvalidCSV, got %v", err)
		}
	})

	t.Run("ragged row shorter than header returns ErrInvalidCSV, not a panic", func(t *testing.T) {
		data := []byte("name,jenjang\nBudi\n")
		_, err := ParseStudentBulkCSV(data)
		if !errors.Is(err, ErrInvalidCSV) {
			t.Errorf("want ErrInvalidCSV, got %v", err)
		}
	})

	t.Run("exactly 1000 data rows is fine", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString("name,jenjang\n")
		for i := 0; i < maxBulkRows; i++ {
			sb.WriteString("Student,sma\n")
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
		sb.WriteString("name,jenjang\n")
		for i := 0; i < maxBulkRows+1; i++ {
			sb.WriteString("Student,sma\n")
		}
		_, err := ParseStudentBulkCSV([]byte(sb.String()))
		if !errors.Is(err, ErrRowLimitExceeded) {
			t.Errorf("want ErrRowLimitExceeded, got %v", err)
		}
	})
}

func TestBuildStudentBulkResultCSV(t *testing.T) {
	results := []StudentBulkResultRow{
		{Name: "Budi", Email: "budi@example.com", Status: "success", Username: "sch_1001", TempPassword: "abc123"},
		{Name: "Siti", Status: "failed", Error: "some error"},
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
	wantHeader := []string{"name", "email", "status", "username", "temp_password", "error"}
	for i, h := range wantHeader {
		if records[0][i] != h {
			t.Errorf("header[%d]: want %s, got %s", i, h, records[0][i])
		}
	}
	wantRow1 := []string{"Budi", "budi@example.com", "success", "sch_1001", "abc123", ""}
	for i, v := range wantRow1 {
		if records[1][i] != v {
			t.Errorf("row1[%d]: want %s, got %s", i, v, records[1][i])
		}
	}
	wantRow2 := []string{"Siti", "", "failed", "", "", "some error"}
	for i, v := range wantRow2 {
		if records[2][i] != v {
			t.Errorf("row2[%d]: want %s, got %s", i, v, records[2][i])
		}
	}
}

func TestProcessStudentBulkRows_Integration(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	// Seed region data for name-resolution tests.
	seedTestRegionData(t, repo)

	t.Run("all-success batch with jenjang only (no address)", func(t *testing.T) {
		schoolID := seedSchoolWithJenjang(t, svc, repo, []string{"sma"})
		rows := []StudentBulkRow{
			{Name: "Budi", Jenjang: "sma"},
			{Name: "Siti", Jenjang: "sma"},
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

	t.Run("address names resolved correctly", func(t *testing.T) {
		schoolID := seedSchoolWithJenjang(t, svc, repo, []string{"sma"})
		sulsel, _ := repo.GetProvinceByName(ctx, "SULAWESI SELATAN")
		if sulsel == nil {
			t.Fatal("SULAWESI SELATAN should exist in seeded data")
		}
		makassar, _ := repo.GetCityByNameInProvince(ctx, "KOTA MAKASSAR", sulsel.ID)
		if makassar == nil {
			t.Fatal("KOTA MAKASSAR should exist")
		}

		sulselProv := "SULAWESI SELATAN"
		makassarKota := "KOTA MAKASSAR"
		mariso := "MARISO"
		kodePos := "90222"
		rows := []StudentBulkRow{
			{Name: "Andi", Jenjang: "sma", Provinsi: &sulselProv, Kota: &makassarKota, Kecamatan: &mariso, KodePos: &kodePos},
		}
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, nil)
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != 1 {
			t.Fatalf("want successCount=1, got %d: %+v", successCount, results)
		}
		if results[0].Status != "success" || results[0].Error != "" {
			t.Errorf("want success, got %+v", results[0])
		}
	})

	t.Run("partial address per row produces row-level error", func(t *testing.T) {
		schoolID := seedSchoolWithJenjang(t, svc, repo, []string{"sma"})
		sulselProv := "SULAWESI SELATAN"
		rows := []StudentBulkRow{
			{Name: "Partial", Jenjang: "sma", Provinsi: &sulselProv},
		}
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, nil)
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != 0 {
			t.Errorf("want successCount=0 for partial address, got %d", successCount)
		}
		if results[0].Status != "failed" || results[0].Error == "" {
			t.Errorf("want failed with error for partial address, got %+v", results[0])
		}
	})

	t.Run("unresolvable province name produces row-level error", func(t *testing.T) {
		schoolID := seedSchoolWithJenjang(t, svc, repo, []string{"sma"})
		bogusProv := "NONEXISTENT PROVINCE"
		makassarKota := "KOTA MAKASSAR"
		mariso := "MARISO"
		rows := []StudentBulkRow{
			{Name: "Bogus", Jenjang: "sma", Provinsi: &bogusProv, Kota: &makassarKota, Kecamatan: &mariso},
		}
		results, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, nil)
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != 0 {
			t.Errorf("want successCount=0 for unresolvable province, got %d", successCount)
		}
		if results[0].Status != "failed" || results[0].Error == "" {
			t.Errorf("want failed with error for unresolvable province, got %+v", results[0])
		}
	})

	t.Run("deactivated school: every row fails, successCount 0", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		if _, err := svc.ChangeSchoolStatus(ctx, schoolID, "deactivated"); err != nil {
			t.Fatalf("ChangeSchoolStatus: %v", err)
		}
		rows := []StudentBulkRow{
			{Name: "A", Jenjang: "sma"},
			{Name: "B", Jenjang: "sma"},
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

	t.Run("unexpected error (not one of the 3 known sentinels) is a row failure, not a batch abort", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		cancelCtx, cancel := context.WithCancel(ctx)
		rows := []StudentBulkRow{
			{Name: "First", Jenjang: "sma"},
			{Name: "Second", Jenjang: "sma"},
			{Name: "Third", Jenjang: "sma"},
		}

		callCount := 0
		results, successCount, err := svc.ProcessStudentBulkRows(cancelCtx, schoolID, rows, func(int) {
			callCount++
			if callCount == 1 {
				cancel()
			}
		})
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: want nil error (row failures must not abort the batch), got %v", err)
		}
		if len(results) != len(rows) {
			t.Fatalf("want a report row for every input row, got %d", len(results))
		}
		if successCount != 1 {
			t.Errorf("want successCount=1 (only the row processed before cancellation), got %d", successCount)
		}
		if results[0].Status != "success" {
			t.Errorf("want first row to succeed before cancellation, got %+v", results[0])
		}
		for _, r := range results[1:] {
			if r.Status != "failed" || r.Error == "" {
				t.Errorf("want rows after cancellation to be reported as failed with a non-empty error, got %+v", r)
			}
			if r.Error == ErrSchoolDeactivated.Error() || r.Error == ErrMissingField.Error() {
				t.Errorf("this row's failure must be the unexpected context-cancellation error, not one of the known sentinels: %+v", r)
			}
		}
	})

	t.Run("progress callback: monotonic non-decreasing, checkpoint every 5 rows for a 50-row batch", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		rows := make([]StudentBulkRow, 50)
		for i := range rows {
			rows[i] = StudentBulkRow{Name: "Student", Jenjang: "sma"}
		}
		var progressCalls []int
		_, successCount, err := svc.ProcessStudentBulkRows(ctx, schoolID, rows, func(pct int) {
			progressCalls = append(progressCalls, pct)
		})
		if err != nil {
			t.Fatalf("ProcessStudentBulkRows: %v", err)
		}
		if successCount != len(rows) {
			t.Fatalf("want successCount=%d, got %d", len(rows), successCount)
		}
		if len(progressCalls) < 10 {
			t.Fatalf("want at least 10 progress calls for a 50-row batch, got %d: %v", len(progressCalls), progressCalls)
		}
		for i := 1; i < len(progressCalls); i++ {
			if progressCalls[i] < progressCalls[i-1] {
				t.Errorf("want monotonically non-decreasing progress, got %v", progressCalls)
				break
			}
		}
		if progressCalls[len(progressCalls)-1] != 100 {
			t.Errorf("want final progress call to be 100, got %v", progressCalls)
		}
	})
}

// seedTestRegionData inserts deterministic region data into the shared test DB.
func seedTestRegionData(t *testing.T, repo *repository.Repository) {
	t.Helper()
	ctx := context.Background()
	pool := repo.Pool()

	// Delete in FK order.
	_, _ = pool.Exec(ctx, `DELETE FROM district`)
	_, _ = pool.Exec(ctx, `DELETE FROM city`)
	_, _ = pool.Exec(ctx, `DELETE FROM province`)

	_, err := pool.Exec(ctx, `INSERT INTO province (id, name) VALUES ('73', 'SULAWESI SELATAN')`)
	if err != nil {
		t.Fatalf("insert province sulsel: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO province (id, name) VALUES ('35', 'JAWA TIMUR')`)
	if err != nil {
		t.Fatalf("insert province jatim: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO city (id, province_id, name) VALUES ('7371', '73', 'KOTA MAKASSAR')`)
	if err != nil {
		t.Fatalf("insert city makassar: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO city (id, province_id, name) VALUES ('3578', '35', 'KOTA SURABAYA')`)
	if err != nil {
		t.Fatalf("insert city surabaya: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO district (id, city_id, name) VALUES ('7371010', '7371', 'MARISO')`)
	if err != nil {
		t.Fatalf("insert district mariso: %v", err)
	}
}
