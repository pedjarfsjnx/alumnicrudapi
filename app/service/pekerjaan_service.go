package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PekerjaanService interface {
	GetAllPekerjaan() ([]model.PekerjaanAlumni, error)
	GetPekerjaanByID(id string) (*model.PekerjaanAlumni, error)
	GetPekerjaanByAlumniID(alumniID string) ([]model.PekerjaanAlumni, error)
	CreatePekerjaan(req *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	UpdatePekerjaan(id string, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	DeletePekerjaan(id string) error // Ini adalah hard delete (admin only)
	GetPekerjaanWithPagination(search, sortBy, order string, page, limit int) (*model.PekerjaanResponse, error)
	SoftDeletePekerjaan(id string, userID string, role string) error
	ListTrash(search string, page, limit int, userID string, role string) ([]model.PekerjaanAlumni, error)
	RestorePekerjaan(id string, userID string, role string) error
	HardDeletePekerjaan(id string, userID string, role string) error

	HandleGetAllPekerjaan(c *fiber.Ctx) error
	HandleGetPekerjaanByID(c *fiber.Ctx) error
	HandleGetPekerjaanByAlumniID(c *fiber.Ctx) error
	HandleCreatePekerjaan(c *fiber.Ctx) error
	HandleUpdatePekerjaan(c *fiber.Ctx) error
	HandleDeletePekerjaan(c *fiber.Ctx) error
	HandleSoftDeletePekerjaan(c *fiber.Ctx) error
	HandleListTrash(c *fiber.Ctx) error
	HandleRestorePekerjaan(c *fiber.Ctx) error
	HandleHardDeletePekerjaan(c *fiber.Ctx) error
}

type pekerjaanService struct {
	pekerjaanRepo repository.PekerjaanRepository
	alumniRepo    repository.AlumniRepository
}

func NewPekerjaanService(pekerjaanRepo repository.PekerjaanRepository, alumniRepo repository.AlumniRepository) PekerjaanService {
	return &pekerjaanService{
		pekerjaanRepo: pekerjaanRepo,
		alumniRepo:    alumniRepo,
	}
}

func (s *pekerjaanService) GetAllPekerjaan() ([]model.PekerjaanAlumni, error) {
	return s.pekerjaanRepo.GetAll()
}

func (s *pekerjaanService) GetPekerjaanByID(id string) (*model.PekerjaanAlumni, error) {
	pekerjaan, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments || err.Error() == "pekerjaan tidak ditemukan" {
			return nil, errors.New("pekerjaan not found")
		}
		return nil, err
	}
	return pekerjaan, nil
}

func (s *pekerjaanService) GetPekerjaanByAlumniID(alumniID string) ([]model.PekerjaanAlumni, error) {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(alumniID)
	if err != nil {
		if err == mongo.ErrNoDocuments || err.Error() == "alumni tidak ditemukan" {
			return nil, errors.New("alumni not found")
		}
		return nil, err
	}

	return s.pekerjaanRepo.GetByAlumniID(alumniID)
}

func (s *pekerjaanService) CreatePekerjaan(req *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(req.AlumniID)
	if err != nil {
		if err == mongo.ErrNoDocuments || err.Error() == "alumni tidak ditemukan" {
			return nil, errors.New("alumni not found")
		}
		return nil, err
	}

	// Validate input
	if err := helper.ValidateCreatePekerjaan(req.AlumniID, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja, req.TanggalMulaiKerja, req.StatusPekerjaan); err != nil {
		return nil, err
	}

	return s.pekerjaanRepo.Create(req)
}

func (s *pekerjaanService) UpdatePekerjaan(id string, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	// Check if pekerjaan exists
	_, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("pekerjaan not found")
	}

	if err := helper.ValidateUpdatePekerjaan(req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja, req.TanggalMulaiKerja, req.StatusPekerjaan); err != nil {
		return nil, err
	}

	return s.pekerjaanRepo.Update(id, req)
}

func (s *pekerjaanService) DeletePekerjaan(id string) error {
	// Check if pekerjaan exists
	_, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		return errors.New("pekerjaan not found")
	}
	// Ini adalah hard delete, hanya untuk admin
	return s.pekerjaanRepo.Delete(id)
}

func (s *pekerjaanService) SoftDeletePekerjaan(id string, userID string, role string) error {
	deleterObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user ID tidak valid")
	}

	pekerjaan, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		return errors.New("pekerjaan not found")
	}

	if role != "admin" {
		alumni, err := s.alumniRepo.GetByUserID(userID)
		if err != nil {
			return errors.New("profil alumni tidak ditemukan untuk user ini")
		}
		if pekerjaan.AlumniID != alumni.ID {
			return errors.New("access denied: you can only delete your own pekerjaan")
		}
	}

	return s.pekerjaanRepo.SoftDelete(id, deleterObjID)
}

func (s *pekerjaanService) GetPekerjaanWithPagination(search, sortBy, order string, page, limit int) (*model.PekerjaanResponse, error) {
	offset := (page - 1) * limit

	pekerjaan, err := s.pekerjaanRepo.GetAllWithPagination(search, sortBy, order, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.pekerjaanRepo.CountWithSearch(search)
	if err != nil {
		return nil, err
	}

	pages := (total + limit - 1) / limit
	if total == 0 {
		pages = 0
	}

	response := &model.PekerjaanResponse{
		Data: pekerjaan,
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

func (s *pekerjaanService) ListTrash(search string, page, limit int, userID string, role string) ([]model.PekerjaanAlumni, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	if role == "admin" {
		return s.pekerjaanRepo.ListTrashAdmin(search, limit, offset)
	}

	alumni, err := s.alumniRepo.GetByUserID(userID)
	if err != nil {
		return []model.PekerjaanAlumni{}, nil // Kembalikan trash kosong
	}
	return s.pekerjaanRepo.ListTrashUser(alumni.ID, search, limit, offset)
}

func (s *pekerjaanService) RestorePekerjaan(id string, userID string, role string) error {
	pekerjaan, err := s.pekerjaanRepo.GetByIDWithDeleted(id)
	if err != nil {
		return errors.New("pekerjaan not found")
	}

	if !pekerjaan.IsDeleted {
		return errors.New("pekerjaan not found in trash")
	}

	if role != "admin" {
		alumni, err := s.alumniRepo.GetByUserID(userID)
		if err != nil {
			return errors.New("profil alumni tidak ditemukan untuk user ini")
		}
		if pekerjaan.AlumniID != alumni.ID {
			return errors.New("access denied: you can only restore your own pekerjaan")
		}
	}
	return s.pekerjaanRepo.Restore(id)
}

func (s *pekerjaanService) HardDeletePekerjaan(id string, userID string, role string) error {
	pekerjaan, err := s.pekerjaanRepo.GetByIDWithDeleted(id)
	if err != nil {
		return errors.New("pekerjaan not found")
	}

	if !pekerjaan.IsDeleted {
		return errors.New("pekerjaan not found in trash")
	}

	if role == "admin" {
		return s.pekerjaanRepo.HardDeleteAdmin(id)
	}

	alumni, err := s.alumniRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("profil alumni tidak ditemukan untuk user ini")
	}
	if pekerjaan.AlumniID != alumni.ID {
		return errors.New("access denied: you can only delete your own pekerjaan")
	}

	return s.pekerjaanRepo.HardDeleteUser(id, alumni.ID)
}

// --- Handlers ---

func (s *pekerjaanService) HandleGetAllPekerjaan(c *fiber.Ctx) error {
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

	if c.Query("page") != "" || c.Query("limit") != "" || c.Query("search") != "" || c.Query("sortBy") != "" {
		response, err := s.GetPekerjaanWithPagination(search, sortBy, order, page, limit)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pekerjaan")
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Pekerjaan data retrieved successfully",
			"data":    response.Data,
			"meta":    response.Meta,
		})
	}

	pekerjaan, err := s.GetAllPekerjaan()
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pekerjaan")
	}
	return helper.SuccessResponse(c, "Pekerjaan data retrieved successfully", pekerjaan)
}

func (s *pekerjaanService) HandleGetPekerjaanByID(c *fiber.Ctx) error {
	id := c.Params("id")
	pekerjaan, err := s.GetPekerjaanByID(id)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan data retrieved successfully", pekerjaan)
}

func (s *pekerjaanService) HandleGetPekerjaanByAlumniID(c *fiber.Ctx) error {
	alumniID := c.Params("alumni_id")
	pekerjaan, err := s.GetPekerjaanByAlumniID(alumniID)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan data retrieved successfully", pekerjaan)
}

func (s *pekerjaanService) HandleCreatePekerjaan(c *fiber.Ctx) error {
	var req model.CreatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	pekerjaan, err := s.CreatePekerjaan(&req)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.CreatedResponse(c, "Pekerjaan created successfully", pekerjaan)
}

func (s *pekerjaanService) HandleUpdatePekerjaan(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.UpdatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	pekerjaan, err := s.UpdatePekerjaan(id, &req)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Pekerjaan updated successfully", pekerjaan)
}

func (s *pekerjaanService) HandleDeletePekerjaan(c *fiber.Ctx) error {
	id := c.Params("id")
	err := s.DeletePekerjaan(id)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan deleted successfully", nil)
}

func (s *pekerjaanService) HandleSoftDeletePekerjaan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)
	id := c.Params("id")

	err := s.SoftDeletePekerjaan(id, userID, role)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		if strings.Contains(err.Error(), "access denied") {
			return helper.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus pekerjaan: "+err.Error())
	}

	return helper.SuccessResponse(c, "Pekerjaan soft deleted successfully", nil)
}

func (s *pekerjaanService) HandleListTrash(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	list, err := s.ListTrash(search, page, limit, userID, role)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data trash")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Trash data retrieved successfully",
		"data":    list,
	})
}

func (s *pekerjaanService) HandleRestorePekerjaan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	if err := s.RestorePekerjaan(id, userID, role); err != nil {
		switch err.Error() {
		case "pekerjaan not found":
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		case "pekerjaan not found in trash":
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found in trash")
		case "access denied: you can only restore your own pekerjaan":
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Access denied: you can only restore your own pekerjaan")
		default:
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal restore pekerjaan: "+err.Error())
		}
	}

	return helper.SuccessResponse(c, "Pekerjaan restored successfully", nil)
}

func (s *pekerjaanService) HandleHardDeletePekerjaan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	if err := s.HardDeletePekerjaan(id, userID, role); err != nil {
		switch err.Error() {
		case "pekerjaan not found":
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		case "pekerjaan not found in trash":
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found in trash")
		case "access denied: you can only delete your own pekerjaan":
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Access denied: you can only delete your own pekerjaan")
		default:
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal hard delete pekerjaan: "+err.Error())
		}
	}

	return helper.SuccessResponse(c, "Pekerjaan permanently deleted", nil)
}
