package service

import (
	"context"
	"errors"
	"fmt"

	"akademi-bimbel/internal/repository"

	"github.com/google/uuid"
)

// ParticipantSelector specifies how to select participants for a bulk exam order.
// Exactly one of StudentIDs, Grade, or All should be set.
type ParticipantSelector struct {
	StudentIDs []string
	Grade      *int
	All        bool
}

var (
	// ErrCrossSchoolStudent is returned when a student_id in the selector
	// does not belong to the specified school (FR-BULK-03).
	ErrCrossSchoolStudent = errors.New("student does not belong to this school")
	// ErrEmptySelector is returned when no selection criteria are provided.
	ErrEmptySelector = errors.New("empty participant selector")
	// ErrDuplicateParticipant is returned when the same student_id appears more
	// than once in a participant selector (FR-BULK-03).
	ErrDuplicateParticipant = errors.New("duplicate student_id in participant selector")
)

// ResolveSchoolParticipantSet resolves a ParticipantSelector against a single school.
// Individual StudentIDs are validated to belong to schoolID (cross-school ids fail).
// Grade selects all students of that grade in the school.
// All selects every non-deleted student in the school (bounded by bulkAllRowCap).
func (s *Service) ResolveSchoolParticipantSet(ctx context.Context, schoolID string, selector ParticipantSelector) ([]uuid.UUID, error) {
	switch {
	case len(selector.StudentIDs) > 0:
		return s.resolveByStudentIDs(ctx, schoolID, selector.StudentIDs)
	case selector.Grade != nil:
		return s.resolveByGrade(ctx, schoolID, *selector.Grade)
	case selector.All:
		return s.resolveAll(ctx, schoolID)
	default:
		return nil, ErrEmptySelector
	}
}

// resolveByStudentIDs validates each student_id belongs to the given school.
// Any cross-school id causes the entire call to fail (FR-BULK-03).
// Duplicates are rejected before any DB calls.
func (s *Service) resolveByStudentIDs(ctx context.Context, schoolID string, studentIDs []string) ([]uuid.UUID, error) {
	seen := make(map[string]bool, len(studentIDs))
	for _, id := range studentIDs {
		if seen[id] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateParticipant, id)
		}
		seen[id] = true
	}

	result := make([]uuid.UUID, 0, len(studentIDs))
	for _, id := range studentIDs {
		student, err := s.storeRepo.GetStudentByID(ctx, id, schoolID)
		if err != nil {
			return nil, err
		}
		if student == nil {
			return nil, fmt.Errorf("%w: student %s not found in school %s", ErrCrossSchoolStudent, id, schoolID)
		}
		result = append(result, uuid.MustParse(id))
	}
	return result, nil
}

// resolveByGrade paginates all students of a specific grade in the school.
func (s *Service) resolveByGrade(ctx context.Context, schoolID string, grade int) ([]uuid.UUID, error) {
	var allIDs []string
	cursor := ""
	for {
		page, nextCursor, err := s.storeRepo.ListStudentsBySchool(ctx, schoolID, repository.StudentFilter{
			Cursor: cursor,
			Limit:  100,
			Grade:  &grade,
		})
		if err != nil {
			return nil, err
		}
		for _, r := range page {
			allIDs = append(allIDs, r.ID)
			if len(allIDs) > bulkAllRowCap {
				return nil, ErrRowLimitExceeded
			}
		}
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return toUUIDs(allIDs), nil
}

// resolveAll paginates every non-deleted student in the school (bounded).
func (s *Service) resolveAll(ctx context.Context, schoolID string) ([]uuid.UUID, error) {
	ids, err := s.collectAllStudentIDs(ctx, schoolID)
	if err != nil {
		return nil, err
	}
	return toUUIDs(ids), nil
}

// toUUIDs converts a []string of UUIDs to []uuid.UUID. Panics if any
// string is not a valid UUID — callers must pass pre-validated data.
func toUUIDs(ids []string) []uuid.UUID {
	result := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		result[i] = uuid.MustParse(id)
	}
	return result
}
