package service

import (
	"context"
	"errors"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

type DashboardView struct {
	EnrolledCourses []DashboardCourseSummary `json:"enrolled_courses"`
	PendingOrder    *DashboardPendingOrder   `json:"pending_order,omitempty"`
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
		courses = append(courses, DashboardCourseSummary{
			ID:           course.ID.String(),
			Title:        course.Title,
			Progress:     progress,
			TotalLessons: total,
			DoneLessons:  done,
		})
	}

	pending, err := s.getPendingOrder(ctx, sID)
	if err != nil {
		return nil, err
	}

	return &DashboardView{
		EnrolledCourses: courses,
		PendingOrder:    pending,
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

func (s *Service) UpdateProfile(ctx context.Context, userID string, name, email, username, phone, address, targetExam *string) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	var normalizedEmail *string
	if email != nil {
		e := normalizeEmail(*email)
		normalizedEmail = &e
	}

	if err := s.repo.UpdateUserProfile(ctx, userID, name, normalizedEmail, username, phone, address, targetExam); err != nil {
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