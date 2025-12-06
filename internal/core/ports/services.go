package ports

import (
	"context"

	"sso/internal/core/domain"
)

type AuthService interface {
	Login(ctx context.Context, rut, password, projectCode string) (string, string, *domain.User, []int, string, error)
	Register(ctx context.Context, user *domain.User) (*domain.User, error)
	ChangePassword(ctx context.Context, rut int, oldPassword, newPassword string) error
	ValidateToken(tokenString string) (*domain.User, []int, error)
	GetUserWithProjects(ctx context.Context, rut int) (*domain.UserWithProjects, error)
	Refresh(ctx context.Context, refreshToken string, projectCode string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
}
