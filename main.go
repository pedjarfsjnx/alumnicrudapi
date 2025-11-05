package main

import (
	"alumni-crud-api/app/repository"
	"alumni-crud-api/app/service"
	"alumni-crud-api/config"
	"alumni-crud-api/database"
	"alumni-crud-api/route"
	"fmt"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	database.ConnectDB()
	defer database.CloseDB()

	// Setup Fiber app
	fiberApp := config.SetupApp()

	// Initialize repositories
	alumniRepo := repository.NewAlumniRepository()
	pekerjaanRepo := repository.NewPekerjaanRepository()
	authRepo := repository.NewAuthRepository()

	// Initialize services
	alumniService := service.NewAlumniService(alumniRepo)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo, alumniRepo)
	authService := service.NewAuthService(authRepo)

	// Setup routes
	route.SetupRoutes(fiberApp, alumniService, pekerjaanService, authService)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(fiberApp.Listen(serverAddr))
}
