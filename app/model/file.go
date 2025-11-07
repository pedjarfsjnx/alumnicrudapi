package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// File merepresentasikan metadata file yang di-upload
type File struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AlumniID     primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`         // Relasi ke Alumni
	FileCategory string             `bson:"file_category" json:"file_category"` // "foto" atau "sertifikat"
	FileName     string             `bson:"file_name" json:"file_name"`         // Nama unik (UUID)
	OriginalName string             `bson:"original_name" json:"original_name"` // Nama file asli
	FilePath     string             `bson:"file_path" json:"file_path"`         // Path di server, misal: "uploads/foto/uuid.png"
	FileSize     int64              `bson:"file_size" json:"file_size"`
	FileType     string             `bson:"file_type" json:"file_type"` // MIME Type, misal: "image/png"
	UploadedAt   time.Time          `bson:"uploaded_at" json:"uploaded_at"`
}
