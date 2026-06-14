package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"akademi-bimbel/internal/repository"
)

type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
}

func (s *Service) GoogleLogin(ctx context.Context, idToken string) (pendingToken string, otpRequired bool, accessToken string, refreshToken string, err error) {
	info, err := s.verifyGoogleToken(ctx, idToken)
	if err != nil {
		return "", false, "", "", err
	}
	if info.Aud != s.cfg.GoogleClientID || info.EmailVerified != "true" {
		return "", false, "", "", ErrInvalidToken
	}

	email := normalizeEmail(info.Email)
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", false, "", "", err
	}

	if user == nil {
		newUser := &repository.User{
			Email:      &email,
			Role:       RoleStudent,
			Name:       info.Name,
			Status:     "active",
			OTPEnabled: false,
		}
		if err := s.repo.CreateUser(ctx, newUser); err != nil {
			return "", false, "", "", err
		}
		user = newUser
	} else if user.Status != "active" {
		return "", false, "", "", ErrAccountDeactivated
	}

	if !user.OTPEnabled {
		access, refresh, err := s.mintSession(ctx, user)
		if err != nil {
			return "", false, "", "", err
		}
		return "", false, access, refresh, nil
	}

	pending, err := s.startOTPChallenge(ctx, user)
	if err != nil {
		return "", false, "", "", err
	}
	return pending, true, "", "", nil
}

func (s *Service) verifyGoogleToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	endpoint := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: tokeninfo status %d", ErrInvalidToken, resp.StatusCode)
	}
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
