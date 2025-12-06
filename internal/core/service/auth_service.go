package service

import (
	"context"
	"errors"
	"time"

	"sso/internal/core/domain"
	"sso/internal/core/ports"
	"sso/internal/utils"

	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    ports.UserRepository
	tokenRepo   ports.TokenRepository
	projectRepo ports.ProjectRepository
	auditRepo   ports.AuditRepository
	jwtSecret   []byte
}

func NewAuthService(userRepo ports.UserRepository, tokenRepo ports.TokenRepository, projectRepo ports.ProjectRepository, auditRepo ports.AuditRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		projectRepo: projectRepo,
		auditRepo:   auditRepo,
		jwtSecret:   []byte(jwtSecret),
	}
}

func (s *AuthService) Login(ctx context.Context, rut, password, projectCode string) (string, string, *domain.User, []int, string, error) {
	// 1. Parsear RUT
	rutInt, _, err := utils.ParseRut(rut)
	if err != nil {
		return "", "", nil, nil, "", errors.New("invalid rut format")
	}

	// 2. Buscar usuario
	user, err := s.userRepo.FindByRut(ctx, rutInt)
	if err != nil {
		return "", "", nil, nil, "", err
	}
	if user == nil {
		return "", "", nil, nil, "", errors.New("invalid credentials")
	}

	// 2. Verificar contraseña
	if user.PasswordHash == nil {
		return "", "", nil, nil, "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, nil, "", errors.New("invalid credentials")
	}

	// 3. Verificar si debe cambiar contraseña
	if user.MustChangePassword {
		return "", "", nil, nil, "", errors.New("PASSWORD_CHANGE_REQUIRED")
	}

	// 3.1 Obtener proyecto para incluir su frontend_url
	project, err := s.projectRepo.GetProjectByCode(ctx, projectCode)
	if err != nil {
		return "", "", nil, nil, "", err
	}
	if project == nil {
		return "", "", nil, nil, "", errors.New("project not found")
	}

	// 4. Verificar membresía en el proyecto y obtener roles
	roles, err := s.projectRepo.GetMemberRoles(ctx, user.ID.String(), projectCode)
	if err != nil {
		return "", "", nil, nil, "", errors.New("no access to this project")
	}

	// 5. Generar Access Token (JWT) con Roles
	accessToken, err := s.generateAccessToken(user, roles)
	if err != nil {
		return "", "", nil, nil, "", err
	}

	// 6. Generar Refresh Token (Opaco)
	refreshTokenID := uuid.New()
	refreshTokenStr := refreshTokenID.String()
	refreshTokenHash := hashRefreshToken(refreshTokenStr)

	refreshToken := &domain.RefreshToken{
		ID:        refreshTokenID,
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(24 * 7 * time.Hour), // 7 días
	}

	// 7. Guardar Refresh Token
	if err := s.tokenRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return "", "", nil, nil, "", err
	}

	s.logAuditAsync(domain.AuditLog{
		UserID:      &user.ID,
		Action:      "LOGIN_SUCCESS",
		Description: "Login exitoso",
	})

	return accessToken, refreshTokenStr, user, roles, project.FrontendURL, nil
}

func (s *AuthService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	// 1. Generar password inicial (primeros 4 dígitos del RUT)
	// Asumimos que user.Rut tiene el RUT numérico (ej: 12345678)
	rutStr := strconv.Itoa(user.Rut)
	if len(rutStr) < 4 {
		return nil, errors.New("invalid rut length")
	}
	initialPassword := rutStr[:4]

	// 2. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedPwdStr := string(hashedPassword)
	user.PasswordHash = &hashedPwdStr
	user.MustChangePassword = true

	// 3. Guardar usuario
	return s.userRepo.Save(ctx, user)
}

func (s *AuthService) ChangePassword(ctx context.Context, rut int, oldPassword, newPassword string) error {
	// 1. Buscar usuario
	user, err := s.userRepo.FindByRut(ctx, rut)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 2. Verificar contraseña actual
	if user.PasswordHash == nil {
		return errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid credentials")
	}

	// 3. Hash nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 4. Actualizar usuario
	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword), false); err != nil {
		return err
	}

	s.logAuditAsync(domain.AuditLog{
		UserID:      &user.ID,
		Action:      "PASSWORD_CHANGED",
		Description: "Contraseña cambiada",
	})

	return nil
}

func (s *AuthService) ValidateToken(tokenString string) (*domain.User, []int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, nil, errors.New("invalid subject")
	}

	var roles []int
	if rolesClaim, ok := claims["roles"].([]interface{}); ok {
		for _, r := range rolesClaim {
			if rFloat, ok := r.(float64); ok {
				roles = append(roles, int(rFloat))
			}
		}
	}

	// Aquí podrías buscar el usuario en la BD si necesitas más datos o verificar si sigue activo
	// Por ahora devolvemos un usuario parcial con el ID
	uid, _ := uuid.Parse(userIDStr)
	return &domain.User{ID: uid}, roles, nil
}

func (s *AuthService) generateAccessToken(user *domain.User, roles []int) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"roles": roles,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) GetUserWithProjects(ctx context.Context, rut int) (*domain.UserWithProjects, error) {
	// 1. Get User
	user, err := s.userRepo.FindByRut(ctx, rut)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 2. Get Projects and Roles
	projects, err := s.projectRepo.GetUserProjectsWithRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 3. Construct response
	return &domain.UserWithProjects{
		User:     user,
		Projects: projects,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, projectCode string) (string, string, error) {
	tokenID, err := uuid.Parse(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	stored, err := s.tokenRepo.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if stored.IsRevoked || time.Now().After(stored.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}

	if stored.TokenHash != hashRefreshToken(refreshToken) {
		return "", "", errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil || user == nil {
		return "", "", errors.New("user not found")
	}

	roles, err := s.projectRepo.GetMemberRoles(ctx, user.ID.String(), projectCode)
	if err != nil {
		return "", "", errors.New("no access to this project")
	}

	accessToken, err := s.generateAccessToken(user, roles)
	if err != nil {
		return "", "", err
	}

	// Rotate refresh token
	newRefreshID := uuid.New()
	newRefreshStr := newRefreshID.String()
	newHash := hashRefreshToken(newRefreshStr)

	if err := s.tokenRepo.SaveRefreshToken(ctx, &domain.RefreshToken{
		ID:        newRefreshID,
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(24 * 7 * time.Hour),
	}); err != nil {
		return "", "", err
	}

	// Revoke old token (best effort)
	_ = s.tokenRepo.RevokeRefreshToken(ctx, tokenID)

	return accessToken, newRefreshStr, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenID, err := uuid.Parse(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	stored, err := s.tokenRepo.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return errors.New("invalid refresh token")
	}
	if stored.TokenHash != hashRefreshToken(refreshToken) {
		return errors.New("invalid refresh token")
	}

	return s.tokenRepo.RevokeRefreshToken(ctx, tokenID)
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *AuthService) logAuditAsync(entry domain.AuditLog) {
	if s.auditRepo == nil {
		return
	}
	go s.auditRepo.InsertAuditLog(context.Background(), &entry)
}
