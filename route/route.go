package route

import (
	"alumni-crud-api/app/service"
	"alumni-crud-api/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(fiberApp *fiber.App, alumniService service.AlumniService, pekerjaanService service.PekerjaanService, authService *service.AuthService) {
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Alumni CRUD API Server",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"auth": fiber.Map{
					"POST /alumni-crud-api/auth/login":  "Login to get access token",
					"GET /alumni-crud-api/auth/profile": "Get user profile (requires auth)",
				},
				"alumni": fiber.Map{
					"GET /alumni-crud-api/alumni":        "Get all alumni (requires auth)",
					"GET /alumni-crud-api/alumni/:id":    "Get alumni by ID (requires auth)",
					"POST /alumni-crud-api/alumni":       "Create new alumni (admin only)",
					"PUT /alumni-crud-api/alumni/:id":    "Update alumni (admin only)",
					"DELETE /alumni-crud-api/alumni/:id": "Delete alumni (admin only)",
				},
				"pekerjaan": fiber.Map{
					"GET /alumni-crud-api/pekerjaan":                    "Get all pekerjaan (requires auth)",
					"GET /alumni-crud-api/pekerjaan/:id":                "Get pekerjaan by ID (requires auth)",
					"GET /alumni-crud-api/pekerjaan/alumni/:alumni_id":  "Get pekerjaan by alumni ID (admin only)",
					"POST /alumni-crud-api/pekerjaan":                   "Create new pekerjaan (admin only)",
					"PUT /alumni-crud-api/pekerjaan/:id":                "Update pekerjaan (admin only)",
					"DELETE /alumni-crud-api/pekerjaan/:id":             "Delete pekerjaan (admin only)",
					"PATCH /alumni-crud-api/pekerjaan/:id/soft-delete":  "Soft delete pekerjaan (admin or owner)",
					"GET /alumni-crud-api/pekerjaan/trash":              "List trash (admin: all, user: own)",
					"PATCH /alumni-crud-api/pekerjaan/:id/restore":      "Restore from trash (admin or owner)",
					"DELETE /alumni-crud-api/pekerjaan/:id/hard-delete": "Hard delete from trash (admin or owner)",
				},
			},
		})
	})

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"status":  "healthy",
			"message": "Server is running properly",
		})
	})

	api := fiberApp.Group("/alumni-crud-api")

	auth := api.Group("/auth")
	auth.Post("/login", authService.HandleLogin)

	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/auth/profile", authService.HandleGetProfile)

	alumni := protected.Group("/alumni")
	alumni.Get("/", alumniService.HandleGetAllAlumni)
	alumni.Get("/:id", alumniService.HandleGetAlumniByID)
	alumni.Post("/", middleware.AdminOnly(), alumniService.HandleCreateAlumni)
	alumni.Put("/:id", middleware.AdminOnly(), alumniService.HandleUpdateAlumni)
	alumni.Delete("/:id", middleware.AdminOnly(), alumniService.HandleDeleteAlumni)

	pekerjaan := protected.Group("/pekerjaan")
	pekerjaan.Get("/", pekerjaanService.HandleGetAllPekerjaan)
	pekerjaan.Get("/:id", pekerjaanService.HandleGetPekerjaanByID)
	pekerjaan.Get("/alumni/:alumni_id", middleware.AdminOnly(), pekerjaanService.HandleGetPekerjaanByAlumniID)
	pekerjaan.Post("/", middleware.AdminOnly(), pekerjaanService.HandleCreatePekerjaan)
	pekerjaan.Put("/:id", middleware.AdminOnly(), pekerjaanService.HandleUpdatePekerjaan)
	pekerjaan.Delete("/:id", middleware.AdminOnly(), pekerjaanService.HandleDeletePekerjaan)
	pekerjaan.Patch("/:id/soft-delete", pekerjaanService.HandleSoftDeletePekerjaan)
	pekerjaan.Get("/trash", pekerjaanService.HandleListTrash)
	pekerjaan.Patch("/:id/restore", pekerjaanService.HandleRestorePekerjaan)
	pekerjaan.Delete("/:id/hard-delete", pekerjaanService.HandleHardDeletePekerjaan)
}
