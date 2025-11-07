package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlumniService interface {
	GetAllAlumni() ([]model.Alumni, error)
	GetAlumniByID(id string) (*model.Alumni, error)
	CreateAlumni(req *model.CreateAlumniRequest, userID string) (*model.Alumni, error)
	UpdateAlumni(id string, req *model.UpdateAlumniRequest) (*model.Alumni, error)
	DeleteAlumni(id string) error
	GetAlumniWithPagination(search, sortBy, order string, page, limit int) (*model.AlumniResponse, error)

	HandleGetAllAlumni(c *fiber.Ctx) error
	HandleGetAlumniByID(c *fiber.Ctx) error
	HandleCreateAlumni(c *fiber.Ctx) error
	HandleUpdateAlumni(c *fiber.Ctx) error
	HandleDeleteAlumni(c *fiber.Ctx) error
}

type alumniService struct {
	alumniRepo repository.AlumniRepository
}

func NewAlumniService(alumniRepo repository.AlumniRepository) AlumniService {
	return &alumniService{
		alumniRepo: alumniRepo,
	}
}

func (s *alumniService) GetAllAlumni() ([]model.Alumni, error) {
	return s.alumniRepo.GetAll()
}

func (s *alumniService) GetAlumniByID(id string) (*model.Alumni, error) {
	alumni, err := s.alumniRepo.GetByID(id)
	if err != nil {
		if err.Error() == "alumni tidak ditemukan" || err == mongo.ErrNoDocuments {
			return nil, errors.New("alumni tidak ditemukan")
		}
		log.Printf("[ERROR] GetAlumniByID service: %v", err)
		return nil, errors.New("gagal mengambil data alumni")
	}
	return alumni, nil
}

func (s *alumniService) CreateAlumni(req *model.CreateAlumniRequest, userID string) (*model.Alumni, error) {
	if err := helper.ValidateCreateAlumni(req.NIM, req.Nama, req.Jurusan, req.Email, req.Angkatan, req.TahunLulus); err != nil {
		return nil, err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("UserID tidak valid")
	}

	// Cek apakah user ini sudah punya data alumni
	_, err = s.alumniRepo.GetByUserID(userID)
	if err == nil {
		return nil, errors.New("user ini sudah terdaftar sebagai alumni")
	}

	return s.alumniRepo.Create(req, userObjID)
}

func (s *alumniService) UpdateAlumni(id string, req *model.UpdateAlumniRequest) (*model.Alumni, error) {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(id)
	if err != nil {
		return nil, err // "alumni tidak ditemukan" atau error lain
	}

	if err := helper.ValidateUpdateAlumni(req.Nama, req.Jurusan, req.Email, req.Angkatan, req.TahunLulus); err != nil {
		return nil, err
	}

	return s.alumniRepo.Update(id, req)
}

func (s *alumniService) DeleteAlumni(id string) error {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(id)
	if err != nil {
		return err // "alumni tidak ditemukan" atau error lain
	}

	return s.alumniRepo.Delete(id)
}

func (s *alumniService) GetAlumniWithPagination(search, sortBy, order string, page, limit int) (*model.AlumniResponse, error) {
	offset := (page - 1) * limit

	alumni, err := s.alumniRepo.GetAllWithPagination(search, sortBy, order, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.alumniRepo.CountWithSearch(search)
	if err != nil {
		return nil, err
	}

	pages := (total + limit - 1) / limit
	if total == 0 {
		pages = 0
	}

	response := &model.AlumniResponse{
		Data: alumni,
		Meta: model.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  total,
			Pages:  pages,
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	}

	return response, nil
}

// --- Handlers ---

// HandleGetAllAlumni godoc
// @Summary Get Semua Alumni
// @Description Mengambil daftar semua alumni, mendukung pagination via query params.
// @Tags Alumni
// @Produce json
// @Security ApiKeyAuth[]
// @Param page query int false "Nomor Halaman"
// @Param limit query int false "Jumlah item per halaman"
// @Param sortBy query string false "Field untuk sorting (contoh: nama, nim, tahun_lulus)"
// @Param order query string false "Urutan sorting (asc atau desc)"
// @Param search query string false "Teks pencarian (mencari di nama, nim, jurusan, email)"
// @Success 200 {object} helper.Response{data=model.AlumniResponse} "Data alumni berhasil diambil"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 500 {object} helper.Response "Gagal mengambil data alumni"
// @Router /alumni [get]

func (s *alumniService) HandleGetAllAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)
	log.Printf("User %s (%s) accessing GET /alumni", username, role)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	if c.Query("page") != "" || c.Query("limit") != "" || c.Query("search") != "" || c.Query("sortBy") != "" || c.Query("order") != "" {
		response, err := s.GetAlumniWithPagination(search, sortBy, order, page, limit)
		if err != nil {
			log.Printf("[ERROR] Alumni pagination service error: %v", err)
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data alumni")
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Alumni data retrieved successfully",
			"data":    response.Data,
			"meta":    response.Meta,
		})
	}

	alumni, err := s.GetAllAlumni()
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data alumni")
	}
	return helper.SuccessResponse(c, "Alumni data retrieved successfully", alumni)
}

// HandleGetAllAlumni godoc
// @Summary Get Semua Alumni
// @Description Mengambil daftar semua alumni, mendukung pagination via query params.
// @Tags Alumni
// @Produce json
// @Security ApiKeyAuth[]
// @Param page query int false "Nomor Halaman"
// @Param limit query int false "Jumlah item per halaman"
// @Param sortBy query string false "Field untuk sorting (contoh: nama, nim, tahun_lulus)"
// @Param order query string false "Urutan sorting (asc atau desc)"
// @Param search query string false "Teks pencarian (mencari di nama, nim, jurusan, email)"
// @Success 200 {object} helper.Response{data=model.AlumniResponse} "Data alumni berhasil diambil"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 500 {object} helper.Response "Gagal mengambil data alumni"
// @Router /alumni [get]

func (s *alumniService) HandleGetAlumniByID(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)
	id := c.Params("id") // ID sekarang string

	log.Printf("User %s (%s) accessing GET /alumni/%s", username, role, id)

	alumni, err := s.GetAlumniByID(id)
	if err != nil {
		if err.Error() == "alumni tidak ditemukan" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni tidak ditemukan")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.SuccessResponse(c, "Alumni data retrieved successfully", alumni)
}

// HandleGetAllAlumni godoc
// @Summary Get Semua Alumni
// @Description Mengambil daftar semua alumni, mendukung pagination via query params.
// @Tags Alumni
// @Produce json
// @Security ApiKeyAuth[]
// @Param page query int false "Nomor Halaman"
// @Param limit query int false "Jumlah item per halaman"
// @Param sortBy query string false "Field untuk sorting (contoh: nama, nim, tahun_lulus)"
// @Param order query string false "Urutan sorting (asc atau desc)"
// @Param search query string false "Teks pencarian (mencari di nama, nim, jurusan, email)"
// @Success 200 {object} helper.Response{data=model.AlumniResponse} "Data alumni berhasil diambil"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 500 {object} helper.Response "Gagal mengambil data alumni"
// @Router /alumni [get]

func (s *alumniService) HandleCreateAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	log.Printf("Admin %s creating new alumni", username)

	var req model.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	userIDStr := c.Locals("user_id").(string)

	alumni, err := s.CreateAlumni(&req, userIDStr)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.CreatedResponse(c, "Alumni created successfully", alumni)
}

func (s *alumniService) HandleUpdateAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id := c.Params("id") // ID sekarang string

	log.Printf("Admin %s updating alumni ID %s", username, id)

	var req model.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	alumni, err := s.UpdateAlumni(id, &req)
	if err != nil {
		if err.Error() == "alumni tidak ditemukan" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni tidak ditemukan")
		}
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Alumni updated successfully", alumni)
}

func (s *alumniService) HandleDeleteAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id := c.Params("id") // ID sekarang string

	log.Printf("Admin %s deleting alumni ID %s", username, id)

	err := s.DeleteAlumni(id)
	if err != nil {
		if err.Error() == "alumni tidak ditemukan" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni tidak ditemukan")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus alumni")
	}

	return helper.SuccessResponse(c, "Alumni deleted successfully", nil)
}
