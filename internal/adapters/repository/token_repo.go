package repository

import (
	"context"
	"net/netip"

	"sso/internal/adapters/repository/dbrepo"
	"sso/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepo) SaveRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	var ip *netip.Addr
	if t.IPAddress != "" {
		parsedIP, err := netip.ParseAddr(t.IPAddress)
		if err == nil {
			ip = &parsedIP
		}
	}

	params := dbrepo.CreateRefreshTokenParams{
		ID:         pgtype.UUID{Bytes: t.ID, Valid: true},
		UserID:     pgtype.UUID{Bytes: t.UserID, Valid: true},
		TokenHash:  t.TokenHash,
		DeviceInfo: pgtype.Text{String: t.DeviceInfo, Valid: t.DeviceInfo != ""},
		IpAddress:  ip,
		ExpiresAt:  pgtype.Timestamp{Time: t.ExpiresAt, Valid: true},
	}

	_, err := r.Q.CreateRefreshToken(ctx, params)
	return err
}

func (r *PostgresRepo) GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*domain.RefreshToken, error) {
	rt, err := r.Q.GetRefreshTokenByID(ctx, pgtype.UUID{Bytes: tokenID, Valid: true})
	if err != nil {
		return nil, err
	}
	var ipStr string
	if rt.IpAddress != nil {
		ipStr = rt.IpAddress.String()
	}
	return &domain.RefreshToken{
		ID:         rt.ID.Bytes,
		UserID:     rt.UserID.Bytes,
		TokenHash:  rt.TokenHash,
		DeviceInfo: rt.DeviceInfo.String,
		IPAddress:  ipStr,
		ExpiresAt:  rt.ExpiresAt.Time,
		CreatedAt:  rt.CreatedAt.Time,
		IsRevoked:  rt.IsRevoked.Bool,
	}, nil
}

func (r *PostgresRepo) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	return r.Q.RevokeRefreshToken(ctx, pgtype.UUID{Bytes: tokenID, Valid: true})
}
