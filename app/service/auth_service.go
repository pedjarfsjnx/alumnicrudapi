package service

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/app/repository"
	"alumni-crud-api/helper"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// Get user from database
	user, passwordHash, err := s.authRepo.GetUserByUsernameOrEmail(req.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("username atau password salah")
		}
		log.Printf("[ERROR] AuthService Login: %v", err)
		return nil, errors.New("error database")
	}

	// Check password
	if !helper.CheckPassword(req.Password, passwordHash) {
		return nil, errors.New("username atau password salah")
	}

	// Generate JWT token
	token, err := helper.GenerateToken(*user)
	if err != nil {
		log.Printf("[ERROR] AuthService GenerateToken: %v", err)
		return nil, errors.New("gagal generate token")
	}

	// Hapus hash sebelum mengirim response
	user.PasswordHash = ""
	response := &model.LoginResponse{
		User:  *user,
		Token: token,
	}

	return response, nil
}

func (s *AuthService) GetProfile(userID string) (*model.User, error) {
	user, err := s.authRepo.GetUserByID(userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
	}
	return user, nil
}

// HandleLogin godoc
// @Summary Login Pengguna
// @Description Autentikasi pengguna dan dapatkan token JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Kredensial Login"
// @Success 200 {object} helper.Response{data=model.LoginResponse} "Login Berhasil"
// @Failure 400 {object} helper.Response "Request body tidak valid"
// @Failure 401 {object} helper.Response "Username atau password salah"
// @Router /auth/login [post]

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

// HandleGetProfile godoc
// @Summary Get Profil Pengguna
// @Description Mengambil profil pengguna yang sedang login (berdasarkan token).
// @Tags Auth
// @Produce json
// @Security ApiKeyAuth[]
// @Success 200 {object} helper.Response{data=model.User} "Profil berhasil diambil"
// @Failure 401 {object} helper.Response "Token tidak valid"
// @Failure 404 {object} helper.Response "User tidak ditemukan"
// @Router /auth/profile [get]

func (s *AuthService) HandleGetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string) // Diubah ke string

	user, err := s.GetProfile(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return helper.ErrorResponse(c, 404, "User tidak ditemukan")
		}
		log.Printf("[ERROR] HandleGetProfile: %v", err)
		return helper.ErrorResponse(c, 500, "Gagal mengambil profil")
	}

	return helper.SuccessResponse(c, "Profile berhasil diambil", user)
}
