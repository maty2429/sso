package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter define todas las rutas y retorna el motor de Gin listo para usar
func NewRouter(authHandler *AuthHandler, projectHandler *ProjectHandler) *gin.Engine {
	r := gin.Default()

	// Configuración CORS abierta para desarrollo
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}))

	// Grupo de rutas para la API v1 (Buenas prácticas de versionado)
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/change-password", authHandler.ChangePassword)
			auth.GET("/users/:rut", authHandler.GetUser)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.POST("/:projectCode/members", projectHandler.AddMember)
		}
	}

	return r
}
