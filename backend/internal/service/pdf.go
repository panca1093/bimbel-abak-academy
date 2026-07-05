package service

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"akademi-bimbel/internal/model"
)

func generateExamCardPDF(reg *model.RegistrationDetail, studentName, tenantName string) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A5", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "KARTU PESERTA UJIAN", "", 1, "L", false, 0, "")
	if tenantName != "" {
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 6, tenantName, "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)

	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Nama Peserta: %s", studentName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Ujian: %s", reg.Exam.Title), "", 1, "L", false, 0, "")

	scheduleText := "Jadwal: -"
	if reg.Exam.ScheduledAt != nil {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			loc = time.UTC
		}
		scheduleText = fmt.Sprintf("Jadwal: %s", reg.Exam.ScheduledAt.In(loc).Format("02 Jan 2006 15:04 WIB"))
	}
	pdf.CellFormat(0, 7, scheduleText, "", 1, "L", false, 0, "")

	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 28)
	pdf.CellFormat(0, 18, fmt.Sprintf("TOKEN: %s", reg.Token), "", 1, "C", false, 0, "")

	pdf.Ln(4)

	pdf.SetFont("Helvetica", "I", 9)
	if reg.Exam.RequiresCheckin {
		if reg.Exam.CheckInWindowMinutes != nil {
			pdf.MultiCell(0, 5, fmt.Sprintf("Harap check-in dalam waktu %d menit sebelum ujian.", *reg.Exam.CheckInWindowMinutes), "", "L", false)
		} else {
			pdf.MultiCell(0, 5, "Harap check-in sebelum ujian dimulai.", "", "L", false)
		}
	} else {
		pdf.MultiCell(0, 5, "Akses bebas pada waktu yang ditentukan.", "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}