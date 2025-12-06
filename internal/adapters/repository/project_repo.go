package repository

import (
	"context"
	"errors"

	"sso/internal/adapters/repository/dbrepo"
	"sso/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepo) GetMemberRoles(ctx context.Context, userID string, projectCode string) ([]int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// 1. Get member ID for the user in the project
	var memberID pgtype.UUID
	err = r.DB.QueryRow(ctx, `
		SELECT pm.id
		FROM project_members pm
		JOIN projects p ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND p.project_code = $2 AND pm.is_active = true
		LIMIT 1
	`, uid, projectCode).Scan(&memberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("member not found in project")
		}
		return nil, err
	}

	// 2. Get Roles
	rolesRows, err := r.Q.GetMemberRoles(ctx, memberID)
	if err != nil {
		return nil, err
	}

	var roles []int
	for _, row := range rolesRows {
		roles = append(roles, int(row.RoleCode))
	}

	return roles, nil
}

func (r *PostgresRepo) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	query := `
		INSERT INTO projects (name, project_code, description, frontend_url, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, project_code, description, frontend_url, is_active, created_at, updated_at
	`

	var row domain.Project
	var desc, frontend pgtype.Text
	var isActive pgtype.Bool

	err := r.DB.QueryRow(ctx, query,
		p.Name,
		p.ProjectCode,
		p.Description,
		p.FrontendURL,
		p.IsActive,
	).Scan(
		&row.ID,
		&row.Name,
		&row.ProjectCode,
		&desc,
		&frontend,
		&isActive,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	row.Description = desc.String
	row.FrontendURL = frontend.String
	row.IsActive = isActive.Bool

	return &row, nil
}

func (r *PostgresRepo) GetProjectByCode(ctx context.Context, projectCode string) (*domain.Project, error) {
	query := `
		SELECT id, name, project_code, description, frontend_url, is_active, created_at, updated_at
		FROM projects
		WHERE project_code = $1
	`

	var row domain.Project
	var desc, frontend pgtype.Text
	var isActive pgtype.Bool

	err := r.DB.QueryRow(ctx, query, projectCode).Scan(
		&row.ID,
		&row.Name,
		&row.ProjectCode,
		&desc,
		&frontend,
		&isActive,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	row.Description = desc.String
	row.FrontendURL = frontend.String
	row.IsActive = isActive.Bool

	return &row, nil
}

func (r *PostgresRepo) AddMember(ctx context.Context, userID uuid.UUID, projectID int32) error {
	query := `
		INSERT INTO project_members (user_id, project_id, is_active)
		VALUES ($1, $2, $3)
	`
	_, err := r.DB.Exec(ctx, query, userID, projectID, true)
	return err
}

func (r *PostgresRepo) GetMemberID(ctx context.Context, userID uuid.UUID, projectID int32) (uuid.UUID, error) {
	query := `
		SELECT id FROM project_members 
		WHERE user_id = $1 AND project_id = $2 AND is_active = true
		LIMIT 1
	`
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, query, userID, projectID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, errors.New("member not found")
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (r *PostgresRepo) AssignRole(ctx context.Context, memberID uuid.UUID, roleCode int) error {
	query := `
		INSERT INTO project_member_roles (member_id, role_code)
		VALUES ($1, $2)
	`
	_, err := r.DB.Exec(ctx, query, memberID, roleCode)
	return err
}

func (r *PostgresRepo) AddMemberWithRoles(ctx context.Context, userID uuid.UUID, projectID int32, roles []int) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	qtx := dbrepo.New(tx)

	member, err := qtx.CreateProjectMember(ctx, dbrepo.CreateProjectMemberParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ProjectID: projectID,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	for _, role := range roles {
		if err := qtx.AssignRoleToMember(ctx, dbrepo.AssignRoleToMemberParams{
			MemberID: member.ID,
			RoleCode: int16(role),
		}); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepo) GetUserProjectsWithRoles(ctx context.Context, userID uuid.UUID) ([]domain.UserProject, error) {
	// Query to get projects and roles for a user
	// We need to join projects, project_members, project_member_roles, and role_definitions
	query := `
		SELECT 
			p.id, p.project_code, p.name,
			rd.code, rd.name
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		LEFT JOIN project_member_roles pmr ON pmr.member_id = pm.id
		LEFT JOIN role_definitions rd ON rd.code = pmr.role_code
		WHERE pm.user_id = $1 AND pm.is_active = true
		ORDER BY p.id, rd.code
	`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to store projects by ID to aggregate roles
	projectMap := make(map[int32]*domain.UserProject)
	var projects []domain.UserProject
	// To keep order, we can maintain a slice of IDs or just append to slice if new

	for rows.Next() {
		var pID int32
		var pCode, pName string
		var rCode pgtype.Int4
		var rName pgtype.Text

		if err := rows.Scan(&pID, &pCode, &pName, &rCode, &rName); err != nil {
			return nil, err
		}

		if _, exists := projectMap[pID]; !exists {
			proj := &domain.UserProject{
				ProjectID:   pID,
				ProjectCode: pCode,
				ProjectName: pName,
				Roles:       []domain.ProjectRole{},
			}
			projectMap[pID] = proj
		}

		if rCode.Valid {
			role := domain.ProjectRole{
				RoleCode: int(rCode.Int32),
				RoleName: rName.String,
			}
			projectMap[pID].Roles = append(projectMap[pID].Roles, role)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice
	for _, p := range projectMap {
		projects = append(projects, *p)
	}

	return projects, nil
}
