package route

import (
	"alumni-crud-api/app/service"
	_ "alumni-crud-api/docs" // Impor folder docs yang di-generate
	"alumni-crud-api/middleware"

	"github.com/gofiber/fiber/v2"

	// INI YANG MEMPERBAIKI ERROR:
	// Kita mengimpor paket dan memberinya nama alias "fiberSwagger"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// Pastikan "fileService service.FileService" ada di sini
func SetupRoutes(
	fiberApp *fiber.App,
	alumniService service.AlumniService,
	pekerjaanService service.PekerjaanService,
	authService *service.AuthService,
	fileService service.FileService,
) {

	// TAMBAHKAN INI: Handler untuk Swagger UI
	// Sekarang "fiberSwagger" sudah dikenali
	fiberApp.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Endpoint publik (tidak berubah)
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Alumni CRUD API Server (MongoDB + File Upload)",
			"version": "2.0.0",
		})
	})

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"status":  "healthy",
			"message": "Server is running properly",
		})
	})

	// Grup API utama Anda
	api := fiberApp.Group("/alumni-crud-api")

	// Rute Autentikasi (tidak berubah)
	auth := api.Group("/auth")
	auth.Post("/login", authService.HandleLogin)

	// Grup rute yang dilindungi (memerlukan token)
	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/auth/profile", authService.HandleGetProfile)

	// Rute Alumni (tidak berubah)
	alumni := protected.Group("/alumni")
	alumni.Get("/", alumniService.HandleGetAllAlumni)
	alumni.Get("/:id", alumniService.HandleGetAlumniByID)
	alumni.Post("/", middleware.AdminOnly(), alumniService.HandleCreateAlumni)
	alumni.Put("/:id", middleware.AdminOnly(), alumniService.HandleUpdateAlumni)
	alumni.Delete("/:id", middleware.AdminOnly(), alumniService.HandleDeleteAlumni)

	// Rute Pekerjaan (tidak berubah)
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

	// Rute File Upload (tidak berubah)
	upload := protected.Group("/upload")
	upload.Post("/foto", fileService.HandleUpload)
	upload.Post("/sertifikat", fileService.HandleUpload)
	upload.Get("/alumni/:alumni_id", fileService.HandleGetFilesByAlumni)
	upload.Delete("/:id", fileService.HandleDeleteFile)
}
