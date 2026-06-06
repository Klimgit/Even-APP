package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  *string
	Role         string
	IsAdmin      bool
	CreatedAt    time.Time
}
