package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	authRepo *repository.AuthRepository
}

func NewAuthService(authRepo *repository.AuthRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// Get user from database
	user, passwordHash, err := s.authRepo.GetUserByUsernameOrEmail(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("username atau password salah")
		}
		return nil, errors.New("error database: " + err.Error())
	}

	// Check password
	if !helper.CheckPassword(req.Password, passwordHash) {
		return nil, errors.New("username atau password salah")
	}

	// Generate JWT token
	token, err := helper.GenerateToken(*user)
	if err != nil {
		return nil, errors.New("gagal generate token")
	}

	response := &model.LoginResponse{
		User:  *user,
		Token: token,
	}

	return response, nil
}

func (s *AuthService) GetProfile(userID int) (*model.User, error) {
	return s.authRepo.GetUserByID(userID)
}

func (s *AuthService) HandleLogin(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, 400, "Request body tidak valid")
	}

	if req.Username == "" || req.Password == "" {
		return helper.ErrorResponse(c, 400, "Username dan password harus diisi")
	}

	response, err := s.Login(req)
	if err != nil {
		return helper.ErrorResponse(c, 401, err.Error())
	}

	return helper.SuccessResponse(c, "Login berhasil", response)
}

func (s *AuthService) HandleGetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	user, err := s.GetProfile(userID)
	if err != nil {
		return helper.ErrorResponse(c, 404, "User tidak ditemukan")
	}

	return helper.SuccessResponse(c, "Profile berhasil diambil", user)
}
