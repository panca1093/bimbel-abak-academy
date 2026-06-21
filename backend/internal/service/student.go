package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type DashboardCourseSummary struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Progress     float64 `json:"progress"`
	TotalLessons int     `json:"total_lessons"`
	DoneLessons  int     `json:"done_lessons"`
}

type DashboardPendingOrder struct {
	ID      string  `json:"id"`
	Product *string `json:"product,omitempty"`
	Amount  float64 `json:"amount"`
}

type DashboardStudySummary struct {
	VisitedLectures      int     `json:"visited_lectures"`
	TotalLectures        int     `json:"total_lectures"`
	EnrolledCoursesCount int     `json:"enrolled_courses_count"`
	CompletedCourses     int     `json:"completed_courses"`
	TotalMinutes         float64 `json:"total_minutes"`
}

// DashboardLeaderboardEntry is a placeholder — populated when ranking data exists.
type DashboardLeaderboardEntry struct{}

type DashboardRanking struct {
	Position    *int                      `json:"position"`
	Points      *float64                  `json:"points"`
	Leaderboard []DashboardLeaderboardEntry `json:"leaderboard"`
}

type DashboardView struct {
	EnrolledCourses []DashboardCourseSummary `json:"enrolled_courses"`
	PendingOrder    *DashboardPendingOrder   `json:"pending_order,omitempty"`
	StudySummary    DashboardStudySummary    `json:"study_summary"`
	Ranking         DashboardRanking         `json:"ranking"`
	ExamProgress    []interface{}            `json:"exam_progress"`
	PopularLessons  []interface{}            `json:"popular_lessons"`
}

type PresignedUploadURL struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Fields    map[string]string `json:"fields"`
	Key       string            `json:"key"`
	PublicURL string            `json:"public_url"`
}

func (s *Service) GetDashboard(ctx context.Context, studentID string) (*DashboardView, error) {
	sID, err := parseUUID(studentID)
	if err != nil {
		return nil, err
	}

	sessions, err := s.storeRepo.ListActiveSessionsByStudent(ctx, sID)
	if err != nil {
		return nil, err
	}

	courses := make([]DashboardCourseSummary, 0, len(sessions))

	var visitedLectures int
	var totalLectures int
	var completedCourses int
	var allCompletedLessonIDs []uuid.UUID

	for _, sess := range sessions {
		course, err := s.storeRepo.GetCourseByID(ctx, sess.CourseID)
		if err != nil {
			return nil, err
		}
		total, err := s.storeRepo.CountLessonsByCourse(ctx, sess.CourseID)
		if err != nil {
			return nil, err
		}
		done := len(sess.CompletedLessons)
		var progress float64
		if total > 0 {
			progress = float64(done) / float64(total)
		}

		visitedLectures += done
		totalLectures += total
		if done >= total {
			completedCourses++
		}

		for lessonID := range sess.CompletedLessons {
			allCompletedLessonIDs = append(allCompletedLessonIDs, lessonID)
		}

		courses = append(courses, DashboardCourseSummary{
			ID:           course.ID.String(),
			Title:        course.Title,
			Progress:     progress,
			TotalLessons: total,
			DoneLessons:  done,
		})
	}

	var totalMinutes float64
	if len(allCompletedLessonIDs) > 0 {
		totalSeconds, err := s.storeRepo.SumCompletedLessonMinutes(ctx, allCompletedLessonIDs)
		if err != nil {
			return nil, err
		}
		totalMinutes = float64(totalSeconds) / 60.0
	}

	pending, err := s.getPendingOrder(ctx, sID)
	if err != nil {
		return nil, err
	}

	return &DashboardView{
		EnrolledCourses: courses,
		PendingOrder:    pending,
		StudySummary: DashboardStudySummary{
			VisitedLectures:      visitedLectures,
			TotalLectures:        totalLectures,
			EnrolledCoursesCount: len(sessions),
			CompletedCourses:     completedCourses,
			TotalMinutes:         totalMinutes,
		},
		Ranking: DashboardRanking{
			Position:    nil,
			Points:      nil,
			Leaderboard: []DashboardLeaderboardEntry{},
		},
		ExamProgress:   []interface{}{},
		PopularLessons: []interface{}{},
	}, nil
}

func (s *Service) getPendingOrder(ctx context.Context, studentID uuid.UUID) (*DashboardPendingOrder, error) {
	orders, _, err := s.storeRepo.ListOrders(ctx, repository.OrderFilter{
		StudentID: &studentID,
		Status:    "payment_pending",
		Limit:     1,
	})
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	order, err := s.storeRepo.GetOrderByID(ctx, orders[0].ID)
	if err != nil {
		return nil, err
	}
	if order.ID.String() == "" {
		return nil, nil
	}
	var product *string
	if len(order.Items) > 0 {
		name := order.Items[0].Name
		product = &name
	}
	return &DashboardPendingOrder{
		ID:      order.ID.String(),
		Product: product,
		Amount:  order.Total,
	}, nil
}

func (s *Service) ListSchools(ctx context.Context) ([]*model.School, error) {
	return s.repo.ListSchools(ctx)
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, name, email, username, phone, address, targetExam *string, grade *int, schoolID *string) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if schoolID != nil {
		if _, err := uuid.Parse(*schoolID); err != nil {
			return nil, ErrInvalidUUID
		}
	}

	var normalizedEmail *string
	if email != nil {
		e := normalizeEmail(*email)
		normalizedEmail = &e
	}

	if err := s.repo.UpdateUserProfile(ctx, userID, name, normalizedEmail, username, phone, address, targetExam, grade, schoolID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	updated, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}
	return updated, nil
}

// presignStorage returns a MinIO client whose endpoint matches the host the
// browser uses (MinioPublicEndpoint). Presigned URLs bind the host into the
// signature, so they must be signed for the public host — not the internal
// docker hostname the API container connects through.
func (s *Service) presignStorage() *minio.Client {
	s.presignOnce.Do(func() {
		endpoint := s.cfg.MinioPublicEndpoint
		if endpoint == "" || endpoint == s.cfg.MinioEndpoint {
			s.presignClient = s.storage
			return
		}
		// Region must be set explicitly: this client's endpoint resolves to the
		// browser host, which the API container cannot reach, so presigning must
		// not trigger a bucket-region lookup. us-east-1 is MinIO's default.
		c, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s.cfg.MinioAccessKey, s.cfg.MinioSecretKey, ""),
			Secure: s.cfg.MinioUseSSL,
			Region: "us-east-1",
		})
		if err != nil {
			s.presignClient = s.storage
			return
		}
		s.presignClient = c
	})
	return s.presignClient
}

func (s *Service) publicObjectURL(bucket, key string) string {
	endpoint := s.cfg.MinioPublicEndpoint
	if endpoint == "" {
		endpoint = s.cfg.MinioEndpoint
	}
	scheme := "http"
	if s.cfg.MinioUseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, bucket, key)
}

func (s *Service) GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string) (*PresignedUploadURL, error) {
	if s.storage == nil {
		return nil, errors.New("storage not configured")
	}
	if userID == "" || filename == "" {
		return nil, errors.New("user_id and filename are required")
	}

	bucket := s.cfg.MinioBucketName
	exists, err := s.storage.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := s.storage.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	// Uploaded objects are served directly via <img src>, so the bucket needs
	// anonymous read. Idempotent — safe to re-apply.
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
	if err := s.storage.SetBucketPolicy(ctx, bucket, policy); err != nil {
		return nil, err
	}

	key := fmt.Sprintf("avatars/%s/%s-%s", userID, uuid.New().String(), filename)
	presigned, err := s.presignStorage().PresignedPutObject(ctx, bucket, key, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return &PresignedUploadURL{
		URL:       presigned.String(),
		Method:    "PUT",
		Fields:    map[string]string{},
		Key:       key,
		PublicURL: s.publicObjectURL(bucket, key),
	}, nil
}

func (s *Service) UpdatePhoto(ctx context.Context, userID, photoURL string) (*model.User, error) {
	if photoURL == "" {
		return nil, errors.New("photo_url is required")
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if err := s.repo.UpdateUserPhoto(ctx, userID, photoURL); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}
	return updated, nil
}
