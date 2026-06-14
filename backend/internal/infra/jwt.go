package infra

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub          string   `json:"sub"`
	Role         string   `json:"role"`
	SchoolID     *string  `json:"school_id,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	jwt.RegisteredClaims
}

type JWTSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTSigner(secret string, ttl time.Duration) *JWTSigner {
	return &JWTSigner{secret: []byte(secret), ttl: ttl}
}

func (s *JWTSigner) SignAccess(sub, role string, schoolID *string, capabilities []string) (tokenString, jti string, err error) {
	jti = fmt.Sprintf("%d", time.Now().UnixNano())
	now := time.Now()
	claims := Claims{
		Sub:          sub,
		Role:         role,
		SchoolID:     schoolID,
		Capabilities: capabilities,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(s.secret)
	return tokenString, jti, err
}

func (s *JWTSigner) ParseAccess(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
