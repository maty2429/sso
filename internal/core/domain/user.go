package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Rut                 int        `json:"rut"`
	Dv                  string     `json:"dv"`
	Email               string     `json:"email"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	PasswordHash        *string    `json:"-"` // Hidden
	MustChangePassword  bool       `json:"must_change_password"`
	PasswordChangedAt   *time.Time `json:"password_changed_at"`
	RecoveryToken       *string    `json:"-"` // Hidden
	RecoveryTokenExpiry *time.Time `json:"-"` // Hidden
	FailedAttempts      int        `json:"failed_attempts"`
	LockedUntil         *time.Time `json:"locked_until"`
	IsActive            bool       `json:"is_active"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	DeviceInfo string
	IPAddress  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	IsRevoked  bool
}
