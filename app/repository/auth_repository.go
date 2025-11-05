package repository

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/database"
)

type AuthRepository struct{}

func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}

func (r *AuthRepository) GetUserByUsernameOrEmail(identifier string) (*model.User, string, error) {
	var user model.User
	var passwordHash string

	err := database.DB.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users 
		WHERE username = $1 OR email = $1
	`, identifier).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, "", err
	}

	return &user, passwordHash, nil
}

func (r *AuthRepository) GetUserByID(id int) (*model.User, error) {
	var user model.User

	err := database.DB.QueryRow(`
		SELECT id, username, email, role, created_at, updated_at
		FROM users 
		WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
