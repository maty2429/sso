package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sso/config"
	"sso/internal/adapters/handler"
	"sso/internal/adapters/repository"
	"sso/internal/core/service"
	"sso/pkg/db"
)

func main() {
	// 1. Cargar Configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Conectar a Base de Datos (Pool de conexiones)
	dbPool, err := db.Connect(cfg.DBSource)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}
	defer dbPool.Close()

	// 3. Inicializar Capas (Inyección de Dependencias)
	// CAPA DE DATOS: Repositorio (SQLC)
	// Nota: PostgresRepo implementa tanto UserRepository como TokenRepository
	repo := repository.NewPostgresRepo(dbPool)

	// CAPA DE LÓGICA: Servicio
	// Inyectamos el repo tres veces porque cumple las tres interfaces (User, Token, Project)
	// Nota: Necesitamos actualizar NewAuthService para aceptar ProjectRepository
	authService := service.NewAuthService(repo, repo, repo, repo, cfg.JWTSecret)
	projectService := service.NewProjectService(repo, repo, repo)
	authMiddleware := handler.NewAuthMiddleware(authService)

	// CAPA DE TRANSPORTE: Handler
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)

	// 4. Inicializar Router (Aquí es donde limpiamos el main)
	r := handler.NewRouter(authHandler, projectHandler, authMiddleware)

	// 5. Configuración del Servidor HTTP con apagado elegante
	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Esperar señal de interrupción (Ctrl+C o Docker Stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Dar 5 segundos para terminar peticiones vivas
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
