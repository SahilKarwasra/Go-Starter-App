package main

import (
	"api/cmd/server/handler"
	"api/cmd/server/repository"
	"api/cmd/server/routes"
	"api/cmd/server/services"
	"config"
	"database"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Repositories
	userRepo := repository.NewUserRepository(db)

	// Services
	authService := services.NewAuthService(
		userRepo,
		cfg.JwtSecret,
		cfg.JwtRefreshSecret,
		cfg.AccessTokenDuration,
		cfg.RefreshTokenDuration,
	)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)

	// Router
	router := routes.SetupRouter(cfg.JwtSecret, authHandler)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on :%s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
