package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Definisikan konstanta untuk validasi
const (
	MaxFotoSize       int64 = 1 * 1024 * 1024 // 1MB
	MaxSertifikatSize int64 = 2 * 1024 * 1024 // 2MB
)

var (
	AllowedFotoTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}
	AllowedSertifikatTypes = map[string]bool{
		"application/pdf": true,
	}
)

type FileService interface {
	HandleUpload(c *fiber.Ctx) error
	HandleGetFilesByAlumni(c *fiber.Ctx) error
	HandleDeleteFile(c *fiber.Ctx) error
}

type fileService struct {
	fileRepo   repository.FileRepository
	alumniRepo repository.AlumniRepository // Kita butuh ini untuk otorisasi
}

// Kita tidak menyimpan uploadPath di service, kita tentukan di handler
func NewFileService(fileRepo repository.FileRepository, alumniRepo repository.AlumniRepository) FileService {
	return &fileService{
		fileRepo:   fileRepo,
		alumniRepo: alumniRepo,
	}
}

// --- Handler Utama ---

// HandleUpload adalah satu fungsi yang menangani FOTO dan SERTIFIKAT
// HandleUpload adalah satu fungsi yang menangani FOTO dan SERTIFIKAT
// HandleUpload godoc
// @Summary Upload File (Foto atau Sertifikat)
// @Description Upload foto (jpg/png, max 1MB) ke /upload/foto atau sertifikat (pdf, max 2MB) ke /upload/sertifikat.
// @Description User hanya bisa upload untuk diri sendiri. Admin bisa upload untuk alumni lain dengan menyertakan 'alumni_id'.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth[]
// @Param file formData file true "File yang akan di-upload (jpg/png/pdf)"
// @Param alumni_id formData string false "Alumni ID (Wajib untuk Admin, diabaikan untuk User)"
// @Success 201 {object} helper.Response{data=model.File} "File berhasil di-upload"
// @Failure 400 {object} helper.Response "Request tidak valid (file hilang, tipe/ukuran salah, dll)"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 403 {object} helper.Response "Akses ditolak (user tidak punya profil alumni)"
// @Failure 500 {object} helper.Response "Server error (gagal simpan file/db)"
// @Router /upload/foto [post]
// @Router /upload/sertifikat [post]

func (s *fileService) HandleUpload(c *fiber.Ctx) error {
	// 1. Dapatkan file dari form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "File 'file' tidak ditemukan di form-data")
	}

	// 2. Tentukan Tipe Upload (Foto vs Sertifikat) dari path
	var fileCategory string
	var uploadPath string
	var allowedTypes map[string]bool
	var maxSize int64

	// --- INI BAGIAN YANG DIPERBAIKI ---
	// Kita ubah dari "/api/v1/..." menjadi "/alumni-crud-api/..."

	if c.Path() == "/alumni-crud-api/upload/foto" {
		fileCategory = "foto"
		uploadPath = "./uploads/foto"
		allowedTypes = AllowedFotoTypes
		maxSize = MaxFotoSize
	} else if c.Path() == "/alumni-crud-api/upload/sertifikat" {
		fileCategory = "sertifikat"
		uploadPath = "./uploads/sertifikat"
		allowedTypes = AllowedSertifikatTypes
		maxSize = MaxSertifikatSize
	} else {
		// Error ini seharusnya tidak terjadi jika rute sudah benar
		errMsg := fmt.Sprintf("Path rute tidak dikenal: %s", c.Path())
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, errMsg)
	}
	// --- AKHIR BAGIAN PERBAIKAN ---

	// 3. Validasi (sesuai Tugas)
	// Validasi Ukuran
	if fileHeader.Size > maxSize {
		msg := fmt.Sprintf("Ukuran file melebihi batas. Maks: %d MB", maxSize/(1024*1024))
		return helper.ErrorResponse(c, fiber.StatusBadRequest, msg)
	}

	// Validasi Tipe
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Tipe file tidak diizinkan untuk kategori ini")
	}

	// 4. Otorisasi (sesuai Tugas)
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	// Admin bisa mengupload untuk siapa saja, user hanya untuk diri sendiri
	var targetAlumniID primitive.ObjectID

	// Admin HARUS menyertakan 'alumni_id' di form-data
	if role == "admin" {
		alumniIDStr := c.FormValue("alumni_id")
		if alumniIDStr == "" {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Admin harus menyertakan 'alumni_id' di form-data")
		}
		targetAlumniID, err = primitive.ObjectIDFromHex(alumniIDStr)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "Format 'alumni_id' tidak valid")
		}
		// Pastikan alumni-nya ada
		if _, err := s.alumniRepo.GetByID(alumniIDStr); err != nil {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni dengan ID tersebut tidak ditemukan")
		}
	} else {
		// User HANYA bisa untuk diri sendiri
		alumni, err := s.alumniRepo.GetByUserID(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Profil alumni Anda tidak ditemukan, tidak dapat mengupload file")
		}
		targetAlumniID = alumni.ID
	}

	// 5. Simpan File ke Disk (sesuai PDF)
	// Buat nama file unik
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext

	// Pastikan direktori ada
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat direktori upload")
	}

	// Simpan file
	filePath := filepath.Join(uploadPath, newFileName)
	if err := c.SaveFile(fileHeader, filePath); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan file ke server")
	}

	// 6. Simpan Metadata ke MongoDB (sesuai PDF + modifikasi)
	fileModel := &model.File{
		AlumniID:     targetAlumniID,
		FileCategory: fileCategory,
		FileName:     newFileName,
		OriginalName: fileHeader.Filename,
		FilePath:     filePath,
		FileSize:     fileHeader.Size,
		FileType:     contentType,
	}

	savedFile, err := s.fileRepo.Create(fileModel)
	if err != nil {
		// Rollback: Hapus file jika gagal simpan ke DB
		os.Remove(filePath)
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan metadata file ke database")
	}

	return helper.CreatedResponse(c, "File berhasil di-upload", savedFile)
}

// HandleUpload godoc
// @Summary Upload File (Foto atau Sertifikat)
// @Description Upload foto (jpg/png, max 1MB) ke /upload/foto atau sertifikat (pdf, max 2MB) ke /upload/sertifikat.
// @Description User hanya bisa upload untuk diri sendiri. Admin bisa upload untuk alumni lain dengan menyertakan 'alumni_id'.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth[]
// @Param file formData file true "File yang akan di-upload (jpg/png/pdf)"
// @Param alumni_id formData string false "Alumni ID (Wajib untuk Admin, diabaikan untuk User)"
// @Success 201 {object} helper.Response{data=model.File} "File berhasil di-upload"
// @Failure 400 {object} helper.Response "Request tidak valid (file hilang, tipe/ukuran salah, dll)"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 403 {object} helper.Response "Akses ditolak (user tidak punya profil alumni)"
// @Failure 500 {object} helper.Response "Server error (gagal simpan file/db)"
// @Router /upload/foto [post]
// @Router /upload/sertifikat [post]

func (s *fileService) HandleGetFilesByAlumni(c *fiber.Ctx) error {
	alumniIDStr := c.Params("alumni_id")
	alumniID, err := primitive.ObjectIDFromHex(alumniIDStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Format Alumni ID tidak valid")
	}

	// Otorisasi: Admin bisa lihat siapa saja, user hanya bisa lihat punya sendiri
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	if role != "admin" {
		alumni, err := s.alumniRepo.GetByUserID(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Profil alumni Anda tidak ditemukan")
		}
		if alumni.ID != alumniID {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Anda hanya dapat melihat file Anda sendiri")
		}
	}

	files, err := s.fileRepo.GetByAlumniID(alumniID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data file")
	}

	return helper.SuccessResponse(c, "Data file berhasil diambil", files)
}

// HandleUpload godoc
// @Summary Upload File (Foto atau Sertifikat)
// @Description Upload foto (jpg/png, max 1MB) ke /upload/foto atau sertifikat (pdf, max 2MB) ke /upload/sertifikat.
// @Description User hanya bisa upload untuk diri sendiri. Admin bisa upload untuk alumni lain dengan menyertakan 'alumni_id'.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth[]
// @Param file formData file true "File yang akan di-upload (jpg/png/pdf)"
// @Param alumni_id formData string false "Alumni ID (Wajib untuk Admin, diabaikan untuk User)"
// @Success 201 {object} helper.Response{data=model.File} "File berhasil di-upload"
// @Failure 400 {object} helper.Response "Request tidak valid (file hilang, tipe/ukuran salah, dll)"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 403 {object} helper.Response "Akses ditolak (user tidak punya profil alumni)"
// @Failure 500 {object} helper.Response "Server error (gagal simpan file/db)"
// @Router /upload/foto [post]
// @Router /upload/sertifikat [post]

func (s *fileService) HandleDeleteFile(c *fiber.Ctx) error {
	fileID := c.Params("id")

	// 1. Dapatkan metadata file dari DB
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "File tidak ditemukan di database")
	}

	// 2. Otorisasi
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	if role != "admin" {
		alumni, err := s.alumniRepo.GetByUserID(userID)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Profil alumni Anda tidak ditemukan")
		}
		if alumni.ID != file.AlumniID {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Anda hanya dapat menghapus file Anda sendiri")
		}
	}

	// 3. Hapus file dari server
	if err := os.Remove(file.FilePath); err != nil {
		log.Printf("Peringatan: Gagal menghapus file fisik '%s' dari server: %v", file.FilePath, err)
	}

	// 4. Hapus metadata dari DB
	if err := s.fileRepo.Delete(fileID); err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus metadata file dari database")
	}

	return helper.SuccessResponse(c, "File berhasil dihapus", nil)
}
