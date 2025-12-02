package handler

import (
	"net/http"
	"sso/internal/core/domain"
	"sso/internal/core/ports"
	"sso/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type loginRequest struct {
	Rut         string `json:"rut" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ProjectCode string `json:"project_code" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, roles, err := h.authService.Login(c.Request.Context(), req.Rut, req.Password, req.ProjectCode)
	if err != nil {
		if err.Error() == "PASSWORD_CHANGE_REQUIRED" {
			c.JSON(http.StatusForbidden, gin.H{"error": "PASSWORD_CHANGE_REQUIRED", "message": "You must change your password"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
		"roles":         roles,
	})
}

type registerRequest struct {
	Rut       int    `json:"rut" binding:"required"`
	Dv        string `json:"dv" binding:"required,len=1"`
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Rut:       req.Rut,
		Dv:        utils.NormalizeDv(req.Dv),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	createdUser, err := h.authService.Register(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    createdUser.ID,
		"email": createdUser.Email,
	})
}

type changePasswordRequest struct {
	Rut         interface{} `json:"rut" binding:"required"` // Can be int or string
	OldPassword string      `json:"old_password" binding:"required"`
	NewPassword string      `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rutInt int
	var err error

	// Handle RUT input (int or string)
	switch v := req.Rut.(type) {
	case float64: // JSON numbers are float64
		rutInt = int(v)
	case string:
		rutInt, _, err = utils.ParseRut(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rut format: " + err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rut format"})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), rutInt, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	rutParam := c.Param("rut")
	rutInt, _, err := utils.ParseRut(rutParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rut format: " + err.Error()})
		return
	}

	userWithProjects, err := h.authService.GetUserWithProjects(c.Request.Context(), rutInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userWithProjects)
}
