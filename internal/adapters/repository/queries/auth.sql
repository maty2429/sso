-- name: CreateUser :one
INSERT INTO users (
    rut, dv, email, first_name, last_name, password_hash, must_change_password
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByRut :one
SELECT * FROM users
WHERE rut = $1 LIMIT 1;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2, must_change_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, device_info, ip_address, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetRefreshTokenByID :one
SELECT * FROM refresh_tokens
WHERE id = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET is_revoked = TRUE
WHERE id = $1;

-- name: CreateProjectMember :one
INSERT INTO project_members (
    user_id, project_id, is_active
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: AssignRoleToMember :exec
INSERT INTO project_member_roles (
    member_id, role_code
) VALUES (
    $1, $2
);

-- name: GetMemberRoles :many
SELECT pmr.role_code, rd.name as role_name, rd.description
FROM project_member_roles pmr
JOIN role_definitions rd ON rd.code = pmr.role_code
WHERE pmr.member_id = $1;

SELECT pm.*
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1 
  AND p.project_code = $2 
  AND pm.is_active = TRUE
LIMIT 1;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (user_id, project_id, action, description, ip_address, meta_data)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateProject :one
INSERT INTO projects (
    name, project_code, description, frontend_url, is_active
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetProjectByCode :one
SELECT * FROM projects
WHERE project_code = $1 LIMIT 1;
