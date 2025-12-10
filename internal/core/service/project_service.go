package service

import (
	"context"
	"errors"

	"sso/internal/core/domain"
	"sso/internal/core/ports"
)

type ProjectService struct {
	projectRepo ports.ProjectRepository
	userRepo    ports.UserRepository
	auditRepo   ports.AuditRepository
}

func NewProjectService(projectRepo ports.ProjectRepository, userRepo ports.UserRepository, auditRepo ports.AuditRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, name, code, description, frontendUrl string) (*domain.Project, error) {
	// 1. Check if project code exists
	existing, err := s.projectRepo.GetProjectByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("project code already exists")
	}

	// 2. Create Project
	project := &domain.Project{
		Name:        name,
		ProjectCode: code,
		Description: description,
		FrontendURL: frontendUrl,
		IsActive:    true,
	}

	return s.projectRepo.CreateProject(ctx, project)
}

func (s *ProjectService) AddMember(ctx context.Context, projectCode string, rut int, roles []int) error {
	// 1. Get Project
	project, err := s.projectRepo.GetProjectByCode(ctx, projectCode)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.New("project not found")
	}

	// 2. Get User
	user, err := s.userRepo.FindByRut(ctx, rut)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 3. Add Member + Roles atomically
	if err := s.projectRepo.AddMemberWithRoles(ctx, user.ID, project.ID, roles); err != nil {
		return err
	}

	// 4. Audit (async)
	if s.auditRepo != nil {
		go s.auditRepo.InsertAuditLog(context.Background(), &domain.AuditLog{
			UserID:      &user.ID,
			ProjectID:   &project.ID,
			Action:      "MEMBER_ADDED",
			Description: "Miembro agregado al proyecto " + project.ProjectCode,
		})
	}

	return nil
}
