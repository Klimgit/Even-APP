package domain

import (
	"time"

	"github.com/google/uuid"
)

type DemoNote struct {
	ID        uuid.UUID
	Text      string
	CreatedAt time.Time
}
