package repository

import (
	"alumni-crud-api/app/model"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PekerjaanRepository interface {
	GetAll() ([]model.PekerjaanAlumni, error)
	GetByID(id string) (*model.PekerjaanAlumni, error)
	GetByAlumniID(alumniID string) ([]model.PekerjaanAlumni, error)
	Create(pekerjaan *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	Update(id string, pekerjaan *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	Delete(id string) error // Hard delete (admin)
	GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.PekerjaanAlumni, error)
	CountWithSearch(search string) (int, error)
	SoftDelete(id string, deleterID primitive.ObjectID) error
	ListTrashAdmin(search string, limit, offset int) ([]model.PekerjaanAlumni, error)
	ListTrashUser(alumniID primitive.ObjectID, search string, limit, offset int) ([]model.PekerjaanAlumni, error)
	Restore(id string) error
	HardDeleteAdmin(id string) error
	HardDeleteUser(id string, alumniID primitive.ObjectID) error
	GetByIDWithDeleted(id string) (*model.PekerjaanAlumni, error) // Untuk restore/harddelete
}

type pekerjaanRepository struct {
	collection *mongo.Collection
}

func NewPekerjaanRepository(db *mongo.Database) PekerjaanRepository {
	return &pekerjaanRepository{
		collection: db.Collection("pekerjaan_alumni"),
	}
}

// buildPekerjaanSearchFilter adalah helper internal
func (r *pekerjaanRepository) buildPekerjaanSearchFilter(search string) bson.M {
	if search == "" {
		return bson.M{}
	}
	searchRegex := bson.M{"$regex": search, "$options": "i"}
	return bson.M{
		"$or": []bson.M{
			{"nama_perusahaan": searchRegex},
			{"posisi_jabatan": searchRegex},
			{"bidang_industri": searchRegex},
			{"lokasi_kerja": searchRegex},
			{"status_pekerjaan": searchRegex},
		},
	}
}

func (r *pekerjaanRepository) GetAll() ([]model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var pekerjaan []model.PekerjaanAlumni
	filter := bson.M{"is_deleted": false}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &pekerjaan); err != nil {
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) GetByID(id string) (*model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	var p model.PekerjaanAlumni
	filter := bson.M{"_id": objID, "is_deleted": false}
	if err := r.collection.FindOne(ctx, filter).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pekerjaan tidak ditemukan")
		}
		return nil, err
	}
	return &p, nil
}

func (r *pekerjaanRepository) GetByIDWithDeleted(id string) (*model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	var p model.PekerjaanAlumni
	filter := bson.M{"_id": objID} // Ambil meskipun is_deleted = true
	if err := r.collection.FindOne(ctx, filter).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pekerjaan tidak ditemukan")
		}
		return nil, err
	}
	return &p, nil
}

func (r *pekerjaanRepository) GetByAlumniID(alumniID string) ([]model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	alumniObjID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, fmt.Errorf("Alumni ID tidak valid: %v", err)
	}

	var pekerjaan []model.PekerjaanAlumni
	filter := bson.M{"alumni_id": alumniObjID, "is_deleted": false}
	opts := options.Find().SetSort(bson.D{{"tanggal_mulai_kerja", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &pekerjaan); err != nil {
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) Create(req *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alumniObjID, err := primitive.ObjectIDFromHex(req.AlumniID)
	if err != nil {
		return nil, fmt.Errorf("Alumni ID tidak valid: %v", err)
	}

	now := time.Now()
	newPekerjaan := model.PekerjaanAlumni{
		AlumniID:            alumniObjID,
		NamaPerusahaan:      req.NamaPerusahaan,
		PosisiJabatan:       req.PosisiJabatan,
		BidangIndustri:      req.BidangIndustri,
		LokasiKerja:         req.LokasiKerja,
		GajiRange:           req.GajiRange,
		TanggalMulaiKerja:   req.TanggalMulaiKerja,
		TanggalSelesaiKerja: req.TanggalSelesaiKerja,
		StatusPekerjaan:     req.StatusPekerjaan,
		DeskripsiPekerjaan:  req.DeskripsiPekerjaan,
		IsDeleted:           false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result, err := r.collection.InsertOne(ctx, newPekerjaan)
	if err != nil {
		return nil, err
	}

	newPekerjaan.ID = result.InsertedID.(primitive.ObjectID)
	return &newPekerjaan, nil
}

func (r *pekerjaanRepository) Update(id string, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID tidak valid: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"nama_perusahaan":       req.NamaPerusahaan,
			"posisi_jabatan":        req.PosisiJabatan,
			"bidang_industri":       req.BidangIndustri,
			"lokasi_kerja":          req.LokasiKerja,
			"gaji_range":            req.GajiRange,
			"tanggal_mulai_kerja":   req.TanggalMulaiKerja,
			"tanggal_selesai_kerja": req.TanggalSelesaiKerja,
			"status_pekerjaan":      req.StatusPekerjaan,
			"deskripsi_pekerjaan":   req.DeskripsiPekerjaan,
			"updated_at":            time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedPekerjaan model.PekerjaanAlumni
	filter := bson.M{"_id": objID, "is_deleted": false}
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedPekerjaan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pekerjaan tidak ditemukan")
		}
		return nil, err
	}

	return &updatedPekerjaan, nil
}

func (r *pekerjaanRepository) Delete(id string) error {
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
		return fmt.Errorf("pekerjaan tidak ditemukan")
	}

	return nil
}

func (r *pekerjaanRepository) SoftDelete(id string, deleterID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID tidak valid: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": time.Now(),
			"deleted_by": deleterID,
			"updated_at": time.Now(),
		},
	}
	filter := bson.M{"_id": objID, "is_deleted": false}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("pekerjaan tidak ditemukan atau sudah dihapus")
	}
	return nil
}

func (r *pekerjaanRepository) GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchFilter := r.buildPekerjaanSearchFilter(search)
	mainFilter := bson.M{"is_deleted": false}
	if search != "" {
		mainFilter = bson.M{"is_deleted": false, "$and": []bson.M{searchFilter}}
	}

	validSortColumns := map[string]bool{"id": true, "alumni_id": true, "nama_perusahaan": true, "posisi_jabatan": true, "bidang_industri": true, "lokasi_kerja": true, "status_pekerjaan": true, "tanggal_mulai_kerja": true, "created_at": true}
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

	var pekerjaan []model.PekerjaanAlumni
	cursor, err := r.collection.Find(ctx, mainFilter, opts)
	if err != nil {
		log.Printf("[ERROR] Query execution failed: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &pekerjaan); err != nil {
		log.Printf("[ERROR] Row scan failed: %v", err)
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) CountWithSearch(search string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searchFilter := r.buildPekerjaanSearchFilter(search)
	mainFilter := bson.M{"is_deleted": false}
	if search != "" {
		mainFilter = bson.M{"is_deleted": false, "$and": []bson.M{searchFilter}}
	}

	count, err := r.collection.CountDocuments(ctx, mainFilter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *pekerjaanRepository) ListTrashAdmin(search string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchFilter := r.buildPekerjaanSearchFilter(search)
	mainFilter := bson.M{"is_deleted": true}
	if search != "" {
		mainFilter = bson.M{"is_deleted": true, "$and": []bson.M{searchFilter}}
	}

	opts := options.Find().
		SetSort(bson.D{{"deleted_at", -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var pekerjaan []model.PekerjaanAlumni
	cursor, err := r.collection.Find(ctx, mainFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &pekerjaan); err != nil {
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) ListTrashUser(alumniID primitive.ObjectID, search string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchFilter := r.buildPekerjaanSearchFilter(search)
	mainFilter := bson.M{
		"is_deleted": true,
		"alumni_id":  alumniID,
	}

	if search != "" {
		mainFilter["$and"] = []bson.M{searchFilter}
	}

	opts := options.Find().
		SetSort(bson.D{{"deleted_at", -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var pekerjaan []model.PekerjaanAlumni
	cursor, err := r.collection.Find(ctx, mainFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &pekerjaan); err != nil {
		return nil, err
	}
	return pekerjaan, nil
}

func (r *pekerjaanRepository) Restore(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID tidak valid: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": false,
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"deleted_at": "",
			"deleted_by": "",
		},
	}
	filter := bson.M{"_id": objID, "is_deleted": true}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("pekerjaan tidak ditemukan di trash")
	}
	return nil
}

func (r *pekerjaanRepository) HardDeleteAdmin(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID tidak valid: %v", err)
	}

	filter := bson.M{"_id": objID, "is_deleted": true}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("pekerjaan tidak ditemukan di trash")
	}
	return nil
}

func (r *pekerjaanRepository) HardDeleteUser(id string, alumniID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID tidak valid: %v", err)
	}

	filter := bson.M{
		"_id":        objID,
		"is_deleted": true,
		"alumni_id":  alumniID,
	}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("pekerjaan tidak ditemukan di trash atau Anda tidak memiliki akses")
	}
	return nil
}
