package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          int32
	ProjectCode string
	Name        string
	Description string
	FrontendURL string // Added FrontendURL
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectMember struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProjectID int32
	Roles     []int
	IsActive  bool
	JoinedAt  time.Time
}

type Role struct {
	Code        int
	Name        string
	Description string
}

type ProjectRole struct {
	RoleCode int    `json:"role_code"`
	RoleName string `json:"role_name"`
}

type UserProject struct {
	ProjectID   int32         `json:"project_id"`
	ProjectCode string        `json:"project_code"`
	ProjectName string        `json:"project_name"`
	Roles       []ProjectRole `json:"roles"`
}

type UserWithProjects struct {
	User     *User         `json:"user"`
	Projects []UserProject `json:"projects"`
}
