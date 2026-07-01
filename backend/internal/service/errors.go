package service

import "errors"

var (
	ErrAlreadyAttempted     = errors.New("already attempted")
	ErrExamNotStarted       = errors.New("exam not started")
	ErrDeviceMismatch       = errors.New("device mismatch")
	ErrCheckinWindowClosed  = errors.New("check-in window closed")
	ErrNotCheckedIn         = errors.New("not checked in")
	ErrAlreadySubmitted     = errors.New("already submitted")
	ErrSessionNotFound      = errors.New("session not found")
	ErrInvalidViolationType = errors.New("invalid violation type")
)
