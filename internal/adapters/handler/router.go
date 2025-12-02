package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter define todas las rutas y retorna el motor de Gin listo para usar
func NewRouter(authHandler *AuthHandler, projectHandler *ProjectHandler) *gin.Engine {
	r := gin.Default()

	// Configuración CORS (Vital para que el Frontend no falle)
	r.Use(cors.Default())

	// Grupo de rutas para la API v1 (Buenas prácticas de versionado)
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/change-password", authHandler.ChangePassword)
			auth.GET("/users/:rut", authHandler.GetUser)
			// Aquí añadirías el refresh: auth.POST("/refresh", authHandler.RefreshToken)
		}

		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.POST("/:projectCode/members", projectHandler.AddMember)
		}
	}

	return r
}
