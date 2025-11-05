package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AlumniService interface {
	GetAllAlumni() ([]model.Alumni, error)
	GetAlumniByID(id int) (*model.Alumni, error)
	CreateAlumni(req *model.CreateAlumniRequest) (*model.Alumni, error)
	UpdateAlumni(id int, req *model.UpdateAlumniRequest) (*model.Alumni, error)
	DeleteAlumni(id int) error
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

func (s *alumniService) GetAlumniByID(id int) (*model.Alumni, error) {
	alumni, err := s.alumniRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("alumni not found")
		}
		return nil, err
	}
	return alumni, nil
}

func (s *alumniService) CreateAlumni(req *model.CreateAlumniRequest) (*model.Alumni, error) {
	// Validate input
	if err := helper.ValidateCreateAlumni(req.NIM, req.Nama, req.Jurusan, req.Email, req.Angkatan, req.TahunLulus); err != nil {
		return nil, err
	}

	return s.alumniRepo.Create(req)
}

func (s *alumniService) UpdateAlumni(id int, req *model.UpdateAlumniRequest) (*model.Alumni, error) {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("alumni not found")
		}
		return nil, err
	}

	if err := helper.ValidateUpdateAlumni(req.Nama, req.Jurusan, req.Email, req.Angkatan, req.TahunLulus); err != nil {
		return nil, err
	}

	return s.alumniRepo.Update(id, req)
}

func (s *alumniService) DeleteAlumni(id int) error {
	// Check if alumni exists
	_, err := s.alumniRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("alumni not found")
		}
		return err
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

func (s *alumniService) HandleGetAllAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)
	log.Printf("User %s (%s) accessing GET /alumni", username, role)

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
		response, err := s.GetAlumniWithPagination(search, sortBy, order, page, limit)
		if err != nil {
			log.Printf("[ERROR] Alumni pagination service error: %v", err)
			return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get alumni data")
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
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get alumni data")
	}
	return helper.SuccessResponse(c, "Alumni data retrieved successfully", alumni)
}

func (s *alumniService) HandleGetAlumniByID(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	log.Printf("User %s (%s) accessing GET /alumni/%d", username, role, id)

	alumni, err := s.GetAlumniByID(id)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get alumni data")
	}

	return helper.SuccessResponse(c, "Alumni data retrieved successfully", alumni)
}

func (s *alumniService) HandleCreateAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	log.Printf("Admin %s creating new alumni", username)

	var req model.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	alumni, err := s.CreateAlumni(&req)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.CreatedResponse(c, "Alumni created successfully", alumni)
}

func (s *alumniService) HandleUpdateAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	log.Printf("Admin %s updating alumni ID %d", username, id)

	var req model.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	alumni, err := s.UpdateAlumni(id, &req)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Alumni updated successfully", alumni)
}

func (s *alumniService) HandleDeleteAlumni(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	log.Printf("Admin %s deleting alumni ID %d", username, id)

	err = s.DeleteAlumni(id)
	if err != nil {
		if err.Error() == "alumni not found" {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "Alumni not found")
		}
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete alumni")
	}

	return helper.SuccessResponse(c, "Alumni deleted successfully", nil)
}
