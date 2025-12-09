package handler

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter define todas las rutas y retorna el motor de Gin listo para usar
func NewRouter(authHandler *AuthHandler, projectHandler *ProjectHandler, authMiddleware *AuthMiddleware) *gin.Engine {
	r := gin.Default()

	// Rate limiting simple por IP
	r.Use(RateLimitMiddleware())

	// Configuración CORS para producción y pruebas locales
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://mi-frontend-app.com",
			"http://localhost:4200",
			"http://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Grupo de rutas para la API v1 (Buenas prácticas de versionado)
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.Refresh)

			// Rutas protegidas
			auth.GET("/me", authMiddleware.RequireRole(0), authHandler.Me)
			auth.POST("/change-password", authMiddleware.RequireRole(0), authHandler.ChangePassword)
			auth.GET("/users/:rut", authMiddleware.RequireRole(0), authHandler.GetUser)
			auth.POST("/logout", authMiddleware.RequireRole(0), authHandler.Logout)
		}

		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.POST("/:projectCode/members", projectHandler.AddMember)
		}
	}

	return r
}
