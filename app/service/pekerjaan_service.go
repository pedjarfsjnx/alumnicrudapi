package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"database/sql"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PekerjaanService interface {
	GetAllPekerjaan() ([]model.PekerjaanAlumni, error)
	GetPekerjaanByID(id int) (*model.PekerjaanAlumni, error)
	GetPekerjaanByAlumniID(alumniID int) ([]model.PekerjaanAlumni, error)
	CreatePekerjaan(req *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	UpdatePekerjaan(id int, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	DeletePekerjaan(id int) error
	GetPekerjaanWithPagination(search, sortBy, order string, page, limit int) (*model.PekerjaanResponse, error)
	SoftDeletePekerjaan(id int, userID int, role string) error
	ListTrash(search string, page, limit int, userID int, role string) ([]model.PekerjaanAlumni, error)
	RestorePekerjaan(id int, userID int, role string) error
	HardDeletePekerjaan(id int, userID int, role string) error

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

func (s *pekerjaanService) GetPekerjaanByID(id int) (*model.PekerjaanAlumni, error) {
	pekerjaan, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("pekerjaan not found")
		}
		return nil, err
	}
	return pekerjaan, nil
}

func (s *pekerjaanService) GetPekerjaanByAlumniID(alumniID int) ([]model.PekerjaanAlumni, error) {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(alumniID)
	if err != nil {
		if err == sql.ErrNoRows {
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
		if err == sql.ErrNoRows {
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

func (s *pekerjaanService) UpdatePekerjaan(id int, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	// Check if pekerjaan exists
	_, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("pekerjaan not found")
		}
		return nil, err
	}

	if err := helper.ValidateUpdatePekerjaan(req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja, req.TanggalMulaiKerja, req.StatusPekerjaan); err != nil {
		return nil, err
	}

	return s.pekerjaanRepo.Update(id, req)
}

func (s *pekerjaanService) DeletePekerjaan(id int) error {
	// Check if pekerjaan exists
	_, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pekerjaan not found")
		}
		return err
	}

	return s.pekerjaanRepo.Delete(id)
}

func (s *pekerjaanService) SoftDeletePekerjaan(id int, userID int, role string) error {
	// Check if pekerjaan exists
	_, err := s.pekerjaanRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pekerjaan not found")
		}
		return err
	}

	// Authorization check
	if role != "admin" {
		// Regular user can only delete their own pekerjaan
		ownerUserID, err := s.pekerjaanRepo.GetOwnerUserID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("pekerjaan not found or no owner")
			}
			return err
		}

		if ownerUserID != userID {
			return errors.New("access denied: you can only delete your own pekerjaan")
		}
	}
	// Admin can delete any pekerjaan, so no additional check needed

	return s.pekerjaanRepo.SoftDelete(id, userID)
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

func (s *pekerjaanService) ListTrash(search string, page, limit int, userID int, role string) ([]model.PekerjaanAlumni, error) {
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
	return s.pekerjaanRepo.ListTrashUser(userID, search, limit, offset)
}

func (s *pekerjaanService) RestorePekerjaan(id int, userID int, role string) error {
	// check ownership for non-admin
	if role != "admin" {
		ownerID, err := s.pekerjaanRepo.GetOwnerUserID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("pekerjaan not found")
			}
			return err
		}
		if ownerID != userID {
			return errors.New("access denied: you can only restore your own pekerjaan")
		}
	}
	return s.pekerjaanRepo.Restore(id)
}

func (s *pekerjaanService) HardDeletePekerjaan(id int, userID int, role string) error {
	if role == "admin" {
		return s.pekerjaanRepo.HardDeleteAdmin(id)
	}
	// user only own
	return s.pekerjaanRepo.HardDeleteUser(id, userID)
}

func (s *pekerjaanService) HandleGetAllPekerjaan(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "id")
	order := c.Query("order", "asc")
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
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get pekerjaan data")
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
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get pekerjaan data")
	}
	return helper.SuccessResponse(c, "Pekerjaan data retrieved successfully", pekerjaan)
}

func (s *pekerjaanService) HandleGetPekerjaanByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	pekerjaan, err := s.GetPekerjaanByID(id)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get pekerjaan data")
	}

	return helper.SuccessResponse(c, "Pekerjaan data retrieved successfully", pekerjaan)
}

func (s *pekerjaanService) HandleGetPekerjaanByAlumniID(c *fiber.Ctx) error {
	alumniID, err := strconv.Atoi(c.Params("alumni_id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid alumni ID parameter")
	}

	pekerjaan, err := s.GetPekerjaanByAlumniID(alumniID)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get pekerjaan data")
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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	err = s.DeletePekerjaan(id)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan deleted successfully", nil)
}

func (s *pekerjaanService) HandleSoftDeletePekerjaan(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	err = s.SoftDeletePekerjaan(id, userID, role)
	if err != nil {
		if err.Error() == "pekerjaan not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found")
		}
		if err.Error() == "access denied: you can only delete your own pekerjaan" {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Access denied: you can only delete your own pekerjaan")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan soft deleted successfully", nil)
}

func (s *pekerjaanService) HandleListTrash(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(int)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	list, err := s.ListTrash(search, page, limit, userID, role)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get trash data")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Trash data retrieved successfully",
		"data":    list,
	})
}

func (s *pekerjaanService) HandleRestorePekerjaan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(int)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	if err := s.RestorePekerjaan(id, userID, role); err != nil {
		switch err.Error() {
		case "pekerjaan not found":
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found in trash")
		case "access denied: you can only restore your own pekerjaan":
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Access denied: you can only restore your own pekerjaan")
		default:
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to restore pekerjaan")
		}
	}

	return helper.SuccessResponse(c, "Pekerjaan restored successfully", nil)
}

func (s *pekerjaanService) HandleHardDeletePekerjaan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(int)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	if err := s.HardDeletePekerjaan(id, userID, role); err != nil {
		if err == sql.ErrNoRows {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Pekerjaan not found in trash or access denied")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to hard delete pekerjaan")
	}

	return helper.SuccessResponse(c, "Pekerjaan permanently deleted", nil)
}
