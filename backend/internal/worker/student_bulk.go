package worker

import (
	"context"
	"fmt"
	"log/slog"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/service"
)

// studentBulkProcessor covers just the row-processing step so *service.Service
// (concrete, real-DB-backed) can be swapped for a fake at the worker-dispatch level.
type studentBulkProcessor interface {
	ProcessStudentBulkRows(ctx context.Context, schoolID string, rows []service.StudentBulkRow, onProgress func(int)) ([]service.StudentBulkResultRow, int, error)
}

// runStudentBulkJob downloads the job's input CSV, processes each row through
// the unmodified RegisterStudent path, uploads the per-row report, and
// finishes the job. Any failure before a result CSV is durably uploaded
// finishes the job as failed with the job's progress left unchanged.
func (w *Worker) runStudentBulkJob(ctx context.Context, job model.Job) {
	user, err := w.jobRepo.GetUserByID(ctx, job.CreatedBy)
	if err != nil || user == nil || user.SchoolID == nil {
		w.failStudentBulkJob(ctx, job, fmt.Sprintf("resolve owning school: %v", err))
		return
	}
	schoolID := *user.SchoolID

	if job.InputURL == nil {
		w.failStudentBulkJob(ctx, job, "missing input_url")
		return
	}
	data, err := w.objectStore.GetObjectBytes(ctx, w.privateBucket, *job.InputURL)
	if err != nil {
		w.failStudentBulkJob(ctx, job, fmt.Sprintf("download input: %v", err))
		return
	}

	rows, err := service.ParseStudentBulkCSV(data)
	if err != nil {
		w.failStudentBulkJob(ctx, job, fmt.Sprintf("parse csv: %v", err))
		return
	}

	onProgress := func(pct int) {
		if err := w.jobRepo.UpdateJobProgress(ctx, job.ID, pct); err != nil {
			slog.Error("update job progress", "job_id", job.ID, "err", err)
		}
	}

	results, successCount, err := w.svc.ProcessStudentBulkRows(ctx, schoolID, rows, onProgress)
	if err != nil {
		w.failStudentBulkJob(ctx, job, fmt.Sprintf("process rows: %v", err))
		return
	}

	reportCSV := service.BuildStudentBulkResultCSV(results)
	resultKey := fmt.Sprintf("student-bulk/%s/results/%s.csv", schoolID, job.ID)
	if err := w.objectStore.PutObjectBytes(ctx, w.privateBucket, resultKey, reportCSV, "text/csv"); err != nil {
		w.failStudentBulkJob(ctx, job, fmt.Sprintf("upload result: %v", err))
		return
	}

	status := "succeeded"
	var errMsg *string
	if successCount == 0 {
		status = "failed"
		msg := fmt.Sprintf("student_bulk job %s: all %d rows failed", job.ID, len(rows))
		errMsg = &msg
	}
	if err := w.jobRepo.FinishJob(ctx, job.ID, status, 100, &resultKey, errMsg); err != nil {
		slog.Error("finish job", "job_id", job.ID, "err", err)
	}
}

func (w *Worker) failStudentBulkJob(ctx context.Context, job model.Job, msg string) {
	if err := w.jobRepo.FinishJob(ctx, job.ID, "failed", job.Progress, nil, &msg); err != nil {
		slog.Error("finish job", "job_id", job.ID, "err", err)
	}
}
