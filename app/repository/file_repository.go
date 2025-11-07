package repository

import (
	"alumni-crud-api/app/model"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileRepository interface {
	Create(file *model.File) (*model.File, error)
	GetByID(id string) (*model.File, error)
	GetByAlumniID(alumniID primitive.ObjectID) ([]model.File, error)
	Delete(id string) error
}

type fileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(db *mongo.Database) FileRepository {
	return &fileRepository{
		collection: db.Collection("files"), // Koleksi baru bernama "files"
	}
}

func (r *fileRepository) Create(file *model.File) (*model.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	file.UploadedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, file)
	if err != nil {
		return nil, err
	}

	file.ID = result.InsertedID.(primitive.ObjectID)
	return file, nil
}

func (r *fileRepository) GetByID(id string) (*model.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	var file model.File
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&file); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("file tidak ditemukan")
		}
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetByAlumniID(alumniID primitive.ObjectID) ([]model.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var files []model.File
	filter := bson.M{"alumni_id": alumniID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &files); err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID tidak valid: %v", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("file tidak ditemukan")
	}
	return nil
}
