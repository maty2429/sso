package handler

import (
	"net/http"
	"strings"

	"sso/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService ports.AuthService
}

func NewAuthMiddleware(authService ports.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) RequireRole(requiredRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		user, roles, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if user has the required role (or higher/equivalent logic if needed)
		// Requirement: "¿Tiene el usuario ALGUNO de sus roles con nivel suficiente?"
		// "Si la ruta pide nivel 50, y el usuario tiene [10, 99], el middleware revisa la lista, encuentra el 99 y lo deja pasar."
		// So we check if any role >= requiredRole.
		
		hasAccess := false
		for _, role := range roles {
			if role >= requiredRole {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Set("user_id", user.ID)
		c.Set("roles", roles)
		c.Next()
	}
}
