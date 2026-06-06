package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims stored in access tokens.
type Claims struct {
	UserID  uuid.UUID `json:"uid"`
	Role    string    `json:"role"`
	IsAdmin bool      `json:"is_admin"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWT access tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, accessTTL time.Duration) *Manager {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	return &Manager{secret: []byte(secret), ttl: accessTTL}
}

func (m *Manager) IssueAccess(userID uuid.UUID, role string, isAdmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Role:    role,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *Manager) ParseAccess(token string) (Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return claims, nil
}
