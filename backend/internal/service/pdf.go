package service

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"akademi-bimbel/internal/assets"
	"akademi-bimbel/internal/model"
)

func generateExamCardPDF(reg *model.RegistrationDetail, studentName, tenantName string, photoBytes []byte) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A5", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pageWidth, pageHeight := pdf.GetPageSize()
	marginLeft, marginTop, marginRight, _ := pdf.GetMargins()
	usableWidth := pageWidth - marginLeft - marginRight
	usableHeight := pageHeight - marginTop - 15

	drawBorder(pdf, marginLeft, marginTop, usableWidth, usableHeight)
	embedLogoInCorner(pdf, pageWidth, marginTop)

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

	embedPhotoBox(pdf, marginLeft+usableWidth-50, marginTop+20, photoBytes)

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

func drawBorder(pdf *gofpdf.Fpdf, x, y, width, height float64) {
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.5)
	pdf.Rect(x, y, width, height, "")
}

func embedLogoInCorner(pdf *gofpdf.Fpdf, pageWidth, marginTop float64) {
	logoReader := bytes.NewReader(assets.LogoMarkPNG)
	opts := gofpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   true,
	}
	logoWidth := 25.0
	logoHeight := 25.0
	x := pageWidth - logoWidth - 8
	y := marginTop + 5

	if err := pdf.RegisterImageOptionsReader("logo", opts, logoReader); err != nil {
		return
	}
	pdf.ImageOptions("logo", x, y, logoWidth, logoHeight, false, opts, 0, "")
}

func embedPhotoBox(pdf *gofpdf.Fpdf, x, y float64, photoBytes []byte) {
	boxWidth := 48.0
	boxHeight := 55.0

	pdf.SetDrawColor(100, 100, 100)
	pdf.SetLineWidth(0.5)
	pdf.Rect(x, y, boxWidth, boxHeight, "")

	if len(photoBytes) > 0 {
		photoReader := bytes.NewReader(photoBytes)
		imageType := detectImageType(photoBytes)
		if imageType == "" {
			return
		}

		opts := gofpdf.ImageOptions{
			ImageType: imageType,
			ReadDpi:   true,
		}
		innerX := x + 1
		innerY := y + 1
		innerWidth := boxWidth - 2
		innerHeight := boxHeight - 2

		if err := pdf.RegisterImageOptionsReader("photo", opts, photoReader); err != nil {
			return
		}
		pdf.ImageOptions("photo", innerX, innerY, innerWidth, innerHeight, false, opts, 0, "")
	}
}

func detectImageType(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if len(data) >= 6 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}
	return ""
}