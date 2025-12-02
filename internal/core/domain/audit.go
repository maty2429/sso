package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	UserID      *uuid.UUID
	ProjectID   *int32
	Action      string
	Description string
	IPAddress   string
	MetaData    []byte
	CreatedAt   time.Time
}
