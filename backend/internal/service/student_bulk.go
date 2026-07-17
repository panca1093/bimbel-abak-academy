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
	ErrMissingCSVHeader = errors.New("csv missing required name/jenjang header")
	ErrRowLimitExceeded = errors.New("row limit exceeded")
)

type StudentBulkRow struct {
	Name      string
	Email     *string
	Jenjang   string
	Provinsi  *string
	Kota      *string
	Kecamatan *string
	KodePos   *string
}

type StudentBulkResultRow struct {
	Name         string
	Email        string
	Status       string
	Username     string
	TempPassword string
	Error        string
}

// ParseStudentBulkCSV reads a student-bulk upload. jenjang is required; nis is
// ignored if present; provinsi/kota/kecamatan/kode_pos are optional.
func ParseStudentBulkCSV(data []byte) ([]StudentBulkRow, error) {
	r := csv.NewReader(bytes.NewReader(data))

	header, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return nil, ErrMissingCSVHeader
		}
		return nil, ErrInvalidCSV
	}

	nameIdx, jenjangIdx, emailIdx := -1, -1, -1
	provinsiIdx, kotaIdx, kecamatanIdx, kodePosIdx := -1, -1, -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "name":
			nameIdx = i
		case "jenjang":
			jenjangIdx = i
		case "email":
			emailIdx = i
		case "provinsi":
			provinsiIdx = i
		case "kota":
			kotaIdx = i
		case "kecamatan":
			kecamatanIdx = i
		case "kode_pos":
			kodePosIdx = i
		// "nis" is intentionally ignored
		}
	}
	if nameIdx == -1 || jenjangIdx == -1 {
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

		row := StudentBulkRow{
			Name:    record[nameIdx],
			Jenjang: record[jenjangIdx],
		}
		if emailIdx != -1 && record[emailIdx] != "" {
			email := record[emailIdx]
			row.Email = &email
		}
		if provinsiIdx != -1 && record[provinsiIdx] != "" {
			v := record[provinsiIdx]
			row.Provinsi = &v
		}
		if kotaIdx != -1 && record[kotaIdx] != "" {
			v := record[kotaIdx]
			row.Kota = &v
		}
		if kecamatanIdx != -1 && record[kecamatanIdx] != "" {
			v := record[kecamatanIdx]
			row.Kecamatan = &v
		}
		if kodePosIdx != -1 && record[kodePosIdx] != "" {
			v := record[kodePosIdx]
			row.KodePos = &v
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ProcessStudentBulkRows applies RegisterStudent to each row, resolving
// province/city/district names to IDs before passing them to RegisterStudent.
func (s *Service) ProcessStudentBulkRows(ctx context.Context, schoolID string, rows []StudentBulkRow, onProgress func(pct int)) ([]StudentBulkResultRow, int, error) {
	results := make([]StudentBulkResultRow, len(rows))
	successCount := 0

	checkpoint := len(rows) / 10
	if checkpoint < 1 {
		checkpoint = 1
	}

	for i, r := range rows {
		result := StudentBulkResultRow{Name: r.Name}
		if r.Email != nil {
			result.Email = *r.Email
		}

		// Resolve address names to IDs (all-or-nothing).
		var provinsiID, kotaID, kecamatanID, kodePos *string
		addrCount := 0
		if r.Provinsi != nil {
			addrCount++
		}
		if r.Kota != nil {
			addrCount++
		}
		if r.Kecamatan != nil {
			addrCount++
		}
		if addrCount > 0 && addrCount < 3 {
			result.Status = "failed"
			result.Error = ErrIncompleteAddress.Error()
			results[i] = result
			if onProgress != nil && (i+1)%checkpoint == 0 {
				onProgress((i + 1) * 100 / len(rows))
			}
			continue
		}
		if addrCount == 3 {
			prov, err := s.storeRepo.GetProvinceByName(ctx, *r.Provinsi)
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			if prov == nil {
				result.Status = "failed"
				result.Error = ErrInvalidProvinsi.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			provinsiID = &prov.ID

			city, err := s.storeRepo.GetCityByNameInProvince(ctx, *r.Kota, *provinsiID)
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			if city == nil {
				result.Status = "failed"
				result.Error = ErrInvalidKota.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			kotaID = &city.ID

			district, err := s.storeRepo.GetDistrictByNameInCity(ctx, *r.Kecamatan, *kotaID)
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			if district == nil {
				result.Status = "failed"
				result.Error = ErrInvalidKecamatan.Error()
				results[i] = result
				if onProgress != nil && (i+1)%checkpoint == 0 {
					onProgress((i + 1) * 100 / len(rows))
				}
				continue
			}
			kecamatanID = &district.ID
		}
		if r.KodePos != nil {
			kodePos = r.KodePos
		}

		resp, err := s.RegisterStudent(ctx, schoolID, r.Name, r.Jenjang, r.Email, nil, nil, nil, nil, nil, provinsiID, kotaID, kecamatanID, kodePos)
		if err == nil {
			result.Status = "success"
			result.Username = resp.Username
			result.TempPassword = resp.TempPassword
			successCount++
		} else {
			result.Status = "failed"
			result.Error = err.Error()
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
	_ = w.Write([]string{"name", "email", "status", "username", "temp_password", "error"})
	for _, r := range results {
		_ = w.Write([]string{r.Name, r.Email, r.Status, r.Username, r.TempPassword, r.Error})
	}
	w.Flush()
	return buf.Bytes()
}
