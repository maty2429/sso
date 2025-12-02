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
}

func NewProjectService(projectRepo ports.ProjectRepository, userRepo ports.UserRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
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

	// 3. Add Member
	err = s.projectRepo.AddMember(ctx, user.ID, project.ID)
	if err != nil {
		return err
	}

	// 4. Get Member ID
	// We need to fetch the member ID to assign roles.
	// Since we just added it, we can try to fetch it.
	// We need to update the repo to support this.
	memberID, err := s.projectRepo.GetMemberID(ctx, user.ID, project.ID)
	if err != nil {
		return err
	}

	// 5. Assign Roles
	for _, role := range roles {
		if err := s.projectRepo.AssignRole(ctx, memberID, role); err != nil {
			return err
		}
	}

	return nil
}
