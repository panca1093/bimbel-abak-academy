package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"strings"
)

const maxBulkRows = 1000

var (
	ErrInvalidCSV       = errors.New("invalid csv")
	ErrMissingCSVHeader = errors.New("csv missing required name/nis header")
	ErrRowLimitExceeded = errors.New("row limit exceeded")
)

type StudentBulkRow struct {
	Name  string
	NIS   string
	Email *string
}

type StudentBulkResultRow struct {
	Name         string
	NIS          string
	Email        string
	Status       string
	Username     string
	TempPassword string
	Error        string
}

// ParseStudentBulkCSV reads a student-bulk upload: a required header row
// (name/nis case-insensitive, email optional) followed by up to maxBulkRows
// data rows.
func ParseStudentBulkCSV(data []byte) ([]StudentBulkRow, error) {
	r := csv.NewReader(bytes.NewReader(data))

	header, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return nil, ErrMissingCSVHeader
		}
		return nil, ErrInvalidCSV
	}

	nameIdx, nisIdx, emailIdx := -1, -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "name":
			nameIdx = i
		case "nis":
			nisIdx = i
		case "email":
			emailIdx = i
		}
	}
	if nameIdx == -1 || nisIdx == -1 {
		return nil, ErrMissingCSVHeader
	}

	var rows []StudentBulkRow
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrInvalidCSV
		}
		if len(rows)+1 > maxBulkRows {
			return nil, ErrRowLimitExceeded
		}

		row := StudentBulkRow{Name: record[nameIdx], NIS: record[nisIdx]}
		if emailIdx != -1 && record[emailIdx] != "" {
			email := record[emailIdx]
			row.Email = &email
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ProcessStudentBulkRows applies the existing RegisterStudent to each row,
// assembling a per-row report. Only the known per-row validation failures
// (duplicate NIS, deactivated school, missing field) are captured as row
// errors; anything else aborts the batch and propagates.
func (s *Service) ProcessStudentBulkRows(ctx context.Context, schoolID string, rows []StudentBulkRow, onProgress func(pct int)) ([]StudentBulkResultRow, int, error) {
	results := make([]StudentBulkResultRow, len(rows))
	successCount := 0

	checkpoint := len(rows) / 10
	if checkpoint < 1 {
		checkpoint = 1
	}

	for i, r := range rows {
		result := StudentBulkResultRow{Name: r.Name, NIS: r.NIS}
		if r.Email != nil {
			result.Email = *r.Email
		}

		resp, err := s.RegisterStudent(ctx, schoolID, r.Name, r.NIS, r.Email, nil, nil, nil, nil, nil)
		switch {
		case err == nil:
			result.Status = "success"
			result.Username = resp.Username
			result.TempPassword = resp.TempPassword
			successCount++
		case errors.Is(err, ErrDuplicateNIS), errors.Is(err, ErrSchoolDeactivated), errors.Is(err, ErrMissingField):
			result.Status = "failed"
			result.Error = err.Error()
		default:
			return nil, 0, err
		}

		results[i] = result

		if onProgress != nil && (i+1)%checkpoint == 0 {
			onProgress((i + 1) * 100 / len(rows))
		}
	}

	if onProgress != nil {
		onProgress(100)
	}

	return results, successCount, nil
}

// BuildStudentBulkResultCSV writes the per-row report as CSV bytes.
func BuildStudentBulkResultCSV(results []StudentBulkResultRow) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"name", "nis", "email", "status", "username", "temp_password", "error"})
	for _, r := range results {
		_ = w.Write([]string{r.Name, r.NIS, r.Email, r.Status, r.Username, r.TempPassword, r.Error})
	}
	w.Flush()
	return buf.Bytes()
}
