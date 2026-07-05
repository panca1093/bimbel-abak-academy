package service

import (
	"context"
	"errors"
	"testing"

	"akademi-bimbel/internal/model"
)

func TestEnqueueStudentBulkJobFromData_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	schoolID := createTestSchool(t, svc)
	reg, err := svc.RegisterStudent(ctx, schoolID, "Job Creator", "jc_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterStudent: %v", err)
	}
	createdBy := reg.ID
	fileKey := "student-bulk/" + schoolID + "/" + uniqueSuffix() + "-students.csv"

	t.Run("valid csv creates a queued job pointing at the file key", func(t *testing.T) {
		csv := []byte("name,nis\nBudi,1001\n")
		jobID, err := svc.enqueueStudentBulkJobFromData(ctx, schoolID, createdBy, fileKey, csv)
		if err != nil {
			t.Fatalf("enqueueStudentBulkJobFromData: %v", err)
		}
		if jobID == "" {
			t.Fatal("want non-empty job id")
		}

		job, err := svc.storeRepo.GetJobByID(ctx, jobID)
		if err != nil {
			t.Fatalf("GetJobByID: %v", err)
		}
		if job == nil {
			t.Fatal("job not found after creation")
		}
		if job.Type != "student_bulk" {
			t.Errorf("Type: want student_bulk, got %s", job.Type)
		}
		if job.Status != "queued" {
			t.Errorf("Status: want queued, got %s", job.Status)
		}
		if job.InputURL == nil || *job.InputURL != fileKey {
			t.Errorf("InputURL: want %s, got %v", fileKey, job.InputURL)
		}
		if job.CreatedBy != createdBy {
			t.Errorf("CreatedBy: want %s, got %s", createdBy, job.CreatedBy)
		}
	})

	t.Run("csv missing name/nis header propagates ErrMissingCSVHeader, no job created", func(t *testing.T) {
		csv := []byte("foo,bar\nx,y\n")
		_, err := svc.enqueueStudentBulkJobFromData(ctx, schoolID, createdBy, fileKey, csv)
		if !errors.Is(err, ErrMissingCSVHeader) {
			t.Errorf("want ErrMissingCSVHeader, got %v", err)
		}
	})

	t.Run("over row-limit csv propagates ErrRowLimitExceeded", func(t *testing.T) {
		csv := "name,nis\n"
		for i := 0; i < maxBulkRows+1; i++ {
			csv += "Student,nis" + uniqueSuffix() + "\n"
		}
		_, err := svc.enqueueStudentBulkJobFromData(ctx, schoolID, createdBy, fileKey, []byte(csv))
		if !errors.Is(err, ErrRowLimitExceeded) {
			t.Errorf("want ErrRowLimitExceeded, got %v", err)
		}
	})
}

func TestGetJobStatus_Integration(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	schoolID := createTestSchool(t, svc)
	owner, err := svc.RegisterStudent(ctx, schoolID, "Job Owner", "jo_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterStudent: %v", err)
	}
	other, err := svc.RegisterStudent(ctx, schoolID, "Not Owner", "no_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterStudent: %v", err)
	}

	fileKey := "student-bulk/" + schoolID + "/" + uniqueSuffix() + "-students.csv"
	job := &model.Job{Type: "student_bulk", InputURL: &fileKey, CreatedBy: owner.ID}
	if err := repo.CreateJob(ctx, job); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	t.Run("owner can read status, no presign attempted when result_url is nil", func(t *testing.T) {
		resp, err := svc.GetJobStatus(ctx, job.ID, owner.ID)
		if err != nil {
			t.Fatalf("GetJobStatus: %v", err)
		}
		if resp.ID != job.ID {
			t.Errorf("ID: want %s, got %s", job.ID, resp.ID)
		}
		if resp.Type != "student_bulk" {
			t.Errorf("Type: want student_bulk, got %s", resp.Type)
		}
		if resp.Status != "queued" {
			t.Errorf("Status: want queued, got %s", resp.Status)
		}
		if resp.ResultURL != nil {
			t.Errorf("ResultURL: want nil, got %v", *resp.ResultURL)
		}
		if resp.CreatedAt == "" || resp.UpdatedAt == "" {
			t.Error("want non-empty CreatedAt/UpdatedAt")
		}
	})

	t.Run("different requester gets ErrJobNotFound", func(t *testing.T) {
		_, err := svc.GetJobStatus(ctx, job.ID, other.ID)
		if !errors.Is(err, ErrJobNotFound) {
			t.Errorf("want ErrJobNotFound, got %v", err)
		}
	})

	t.Run("nonexistent job id gets ErrJobNotFound", func(t *testing.T) {
		_, err := svc.GetJobStatus(ctx, "00000000-0000-0000-0000-000000000000", owner.ID)
		if !errors.Is(err, ErrJobNotFound) {
			t.Errorf("want ErrJobNotFound, got %v", err)
		}
	})
}
