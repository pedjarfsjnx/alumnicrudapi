// @title Alumni CRUD API
// @version 1.0
// @description API untuk mengelola data alumni, pekerjaan, dan file upload.
// @description Ini adalah proyek praktikum Pemrograman Backend Lanjut.
// @host localhost:3000
// @BasePath /alumni-crud-api
// @schemes http
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Format: Bearer {token}"

package main

import (
	"alumni-crud-api/app/repository"
	"alumni-crud-api/app/service"
	"alumni-crud-api/config"
	"alumni-crud-api/database"
	_ "alumni-crud-api/docs"
	"alumni-crud-api/route"
	"fmt"
	"log"
	// Diperlukan untuk BodyLimit
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database (MongoDB)
	db := database.ConnectMongo()

	// Setup Fiber app
	fiberApp := config.SetupApp()

	// TAMBAHKAN INI: Sajikan folder 'uploads' secara statis
	// Ini memungkinkan URL "http://.../uploads/foto/namafile.png" diakses
	fiberApp.Static("/uploads", "./uploads")

	// Initialize repositories
	alumniRepo := repository.NewAlumniRepository(db)
	pekerjaanRepo := repository.NewPekerjaanRepository(db)
	authRepo := repository.NewAuthRepository(db)
	fileRepo := repository.NewFileRepository(db) // BARU: Tambahkan file repo

	// Initialize services
	alumniService := service.NewAlumniService(alumniRepo)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo, alumniRepo)
	authService := service.NewAuthService(authRepo)
	// BARU: Tambahkan file service (membutuhkan alumniRepo untuk otorisasi)
	fileService := service.NewFileService(fileRepo, alumniRepo)

	// Setup routes
	// BARU: Tambahkan fileService ke dalam pemanggilan SetupRoutes
	route.SetupRoutes(fiberApp, alumniService, pekerjaanService, authService, fileService)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(fiberApp.Listen(serverAddr))
}
