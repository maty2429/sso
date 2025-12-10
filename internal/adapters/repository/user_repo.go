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

func (r *PostgresRepo) Save(ctx context.Context, u *domain.User) (*domain.User, error) {
	var pwdHash pgtype.Text
	if u.PasswordHash != nil {
		pwdHash = pgtype.Text{String: *u.PasswordHash, Valid: true}
	}

	params := dbrepo.CreateUserParams{
		Rut:                int32(u.Rut),
		Dv:                 u.Dv,
		Email:              u.Email,
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		PasswordHash:       pwdHash,
		MustChangePassword: pgtype.Bool{Bool: u.MustChangePassword, Valid: true},
	}

	row, err := r.Q.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapUser(row), nil
}

func (r *PostgresRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.Q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapUser(row), nil
}

func (r *PostgresRepo) FindByRut(ctx context.Context, rut int) (*domain.User, error) {
	row, err := r.Q.GetUserByRut(ctx, int32(rut))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapUser(row), nil
}

func (r *PostgresRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, rut, dv, email, first_name, last_name, password_hash, must_change_password, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`
	var row dbrepo.User
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&row.ID,
		&row.Rut,
		&row.Dv,
		&row.Email,
		&row.FirstName,
		&row.LastName,
		&row.PasswordHash,
		&row.MustChangePassword,
		&row.IsActive,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapUser(row), nil
}

func (r *PostgresRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string, mustChange bool) error {
	params := dbrepo.UpdateUserPasswordParams{
		ID:                 pgtype.UUID{Bytes: userID, Valid: true},
		PasswordHash:       pgtype.Text{String: passwordHash, Valid: true},
		MustChangePassword: pgtype.Bool{Bool: mustChange, Valid: true},
	}
	_, err := r.Q.UpdateUserPassword(ctx, params)
	return err
}

func mapUser(row dbrepo.User) *domain.User {
	var id uuid.UUID
	if row.ID.Valid {
		id = row.ID.Bytes
	}

	var pwdHash *string
	if row.PasswordHash.Valid {
		s := row.PasswordHash.String
		pwdHash = &s
	}

	return &domain.User{
		ID:                 id,
		Rut:                int(row.Rut),
		Dv:                 row.Dv,
		Email:              row.Email,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		PasswordHash:       pwdHash,
		MustChangePassword: row.MustChangePassword.Bool,
		IsActive:           row.IsActive.Bool,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}
