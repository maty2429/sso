package repository

import (
	"context"
	"net/netip"

	"sso/internal/adapters/repository/dbrepo"
	"sso/internal/core/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepo) InsertAuditLog(ctx context.Context, entry *domain.AuditLog) error {
	var userID pgtype.UUID
	if entry.UserID != nil {
		userID = pgtype.UUID{Bytes: *entry.UserID, Valid: true}
	}

	var projectID pgtype.Int4
	if entry.ProjectID != nil {
		projectID = pgtype.Int4{Int32: *entry.ProjectID, Valid: true}
	}

	var ipAddr *netip.Addr
	if entry.IPAddress != "" {
		if parsed, err := netip.ParseAddr(entry.IPAddress); err == nil {
			ipAddr = &parsed
		}
	}

	params := dbrepo.InsertAuditLogParams{
		UserID:      userID,
		ProjectID:   projectID,
		Action:      entry.Action,
		Description: pgtype.Text{String: entry.Description, Valid: entry.Description != ""},
		IpAddress:   ipAddr,
		MetaData:    entry.MetaData,
	}

	return r.Q.InsertAuditLog(ctx, params)
}
