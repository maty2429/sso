package ports

import (
	"context"

	"sso/internal/core/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByRut(ctx context.Context, rut int) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string, mustChange bool) error
}

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error
}

type ProjectRepository interface {
	GetMemberRoles(ctx context.Context, userID string, projectCode string) ([]int, error)
	CreateProject(ctx context.Context, project *domain.Project) (*domain.Project, error)
	GetProjectByCode(ctx context.Context, projectCode string) (*domain.Project, error)
	AddMember(ctx context.Context, userID uuid.UUID, projectID int32) error
	GetMemberID(ctx context.Context, userID uuid.UUID, projectID int32) (uuid.UUID, error)
	AssignRole(ctx context.Context, memberID uuid.UUID, roleCode int) error
	GetUserProjectsWithRoles(ctx context.Context, userID uuid.UUID) ([]domain.UserProject, error)
	AddMemberWithRoles(ctx context.Context, userID uuid.UUID, projectID int32, roles []int) error
}

type AuditRepository interface {
	InsertAuditLog(ctx context.Context, entry *domain.AuditLog) error
}
