package repository

import (
	"alumni-crud-api/app/model"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlumniRepository interface {
	GetAll() ([]model.Alumni, error)
	GetByID(id string) (*model.Alumni, error)
	GetByUserID(userID string) (*model.Alumni, error) // Penting untuk otorisasi
	Create(alumni *model.CreateAlumniRequest, userID primitive.ObjectID) (*model.Alumni, error)
	Update(id string, alumni *model.UpdateAlumniRequest) (*model.Alumni, error)
	Delete(id string) error
	GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.Alumni, error)
	CountWithSearch(search string) (int, error)
}

type alumniRepository struct {
	collection *mongo.Collection
}

func NewAlumniRepository(db *mongo.Database) AlumniRepository {
	return &alumniRepository{
		collection: db.Collection("alumni"),
	}
}

func (r *alumniRepository) GetAll() ([]model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var alumni []model.Alumni
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &alumni); err != nil {
		return nil, err
	}
	return alumni, nil
}

func (r *alumniRepository) GetByID(id string) (*model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	var a model.Alumni
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&a); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("alumni tidak ditemukan")
		}
		return nil, err
	}
	return &a, nil
}

func (r *alumniRepository) GetByUserID(userID string) (*model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("ID user tidak valid: %v", err)
	}

	var a model.Alumni
	if err := r.collection.FindOne(ctx, bson.M{"user_id": userObjID}).Decode(&a); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("alumni tidak ditemukan")
		}
		return nil, err
	}
	return &a, nil
}

func (r *alumniRepository) Create(req *model.CreateAlumniRequest, userID primitive.ObjectID) (*model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	newAlumni := model.Alumni{
		UserID:     userID, // Diambil dari service
		NIM:        req.NIM,
		Nama:       req.Nama,
		Jurusan:    req.Jurusan,
		Angkatan:   req.Angkatan,
		TahunLulus: req.TahunLulus,
		Email:      req.Email,
		NoTelepon:  req.NoTelepon,
		Alamat:     req.Alamat,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result, err := r.collection.InsertOne(ctx, newAlumni)
	if err != nil {
		return nil, err
	}

	newAlumni.ID = result.InsertedID.(primitive.ObjectID)
	return &newAlumni, nil
}

func (r *alumniRepository) Update(id string, req *model.UpdateAlumniRequest) (*model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"nama":        req.Nama,
			"jurusan":     req.Jurusan,
			"angkatan":    req.Angkatan,
			"tahun_lulus": req.TahunLulus,
			"email":       req.Email,
			"no_telepon":  req.NoTelepon,
			"alamat":      req.Alamat,
			"updated_at":  time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedAlumni model.Alumni
	err = r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updatedAlumni)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("alumni tidak ditemukan")
		}
		return nil, err
	}

	return &updatedAlumni, nil
}

func (r *alumniRepository) Delete(id string) error {
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
		return fmt.Errorf("alumni tidak ditemukan")
	}

	return nil
}

func (r *alumniRepository) buildSearchFilter(search string) bson.M {
	if search == "" {
		return bson.M{}
	}
	searchRegex := bson.M{"$regex": search, "$options": "i"}
	return bson.M{
		"$or": []bson.M{
			{"nama": searchRegex},
			{"nim": searchRegex},
			{"jurusan": searchRegex},
			{"email": searchRegex},
		},
	}
}

func (r *alumniRepository) GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.Alumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := r.buildSearchFilter(search)

	validSortColumns := map[string]bool{"id": true, "nim": true, "nama": true, "jurusan": true, "angkatan": true, "tahun_lulus": true, "email": true, "created_at": true}
	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}
	if sortBy == "id" {
		sortBy = "_id"
	}

	sortOrder := 1 // asc
	if order == "desc" {
		sortOrder = -1 // desc
	}

	opts := options.Find().
		SetSort(bson.D{{sortBy, sortOrder}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var alumni []model.Alumni
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &alumni); err != nil {
		return nil, err
	}
	return alumni, nil
}

func (r *alumniRepository) CountWithSearch(search string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := r.buildSearchFilter(search)
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
