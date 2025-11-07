package repository

import (
	"alumni-crud-api/app/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Definisikan Interface untuk Dependency Injection
type AuthRepository interface {
	GetUserByUsernameOrEmail(identifier string) (*model.User, string, error)
	GetUserByID(id string) (*model.User, error)
}

type authRepository struct {
	collection *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) AuthRepository {
	return &authRepository{
		collection: db.Collection("users"),
	}
}

func (r *authRepository) GetUserByUsernameOrEmail(identifier string) (*model.User, string, error) {
	var user model.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mencari berdasarkan username atau email
	filter := bson.M{
		"$or": []bson.M{
			{"username": identifier},
			{"email": identifier},
		},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", err
		}
		return nil, "", err
	}

	return &user, user.PasswordHash, nil
}

func (r *authRepository) GetUserByID(id string) (*model.User, error) {
	var user model.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err // ID tidak valid
	}

	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err // Termasuk mongo.ErrNoDocuments
	}

	// Hapus password hash untuk keamanan
	user.PasswordHash = ""

	return &user, nil
}
