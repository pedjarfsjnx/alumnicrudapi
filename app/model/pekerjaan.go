package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PekerjaanAlumni struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	AlumniID            primitive.ObjectID  `bson:"alumni_id" json:"alumni_id"`
	NamaPerusahaan      string              `bson:"nama_perusahaan" json:"nama_perusahaan"`
	PosisiJabatan       string              `bson:"posisi_jabatan" json:"posisi_jabatan"`
	BidangIndustri      string              `bson:"bidang_industri" json:"bidang_industri"`
	LokasiKerja         string              `bson:"lokasi_kerja" json:"lokasi_kerja"`
	GajiRange           *string             `bson:"gaji_range,omitempty" json:"gaji_range,omitempty"`
	TanggalMulaiKerja   string              `bson:"tanggal_mulai_kerja" json:"tanggal_mulai_kerja"`
	TanggalSelesaiKerja *string             `bson:"tanggal_selesai_kerja,omitempty" json:"tanggal_selesai_kerja,omitempty"`
	StatusPekerjaan     string              `bson:"status_pekerjaan" json:"status_pekerjaan"`
	DeskripsiPekerjaan  *string             `bson:"deskripsi_pekerjaan,omitempty" json:"deskripsi_pekerjaan,omitempty"`
	IsDeleted           bool                `bson:"is_deleted" json:"is_deleted"`
	DeletedAt           *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	DeletedBy           *primitive.ObjectID `bson:"deleted_by,omitempty" json:"deleted_by,omitempty"`
	CreatedAt           time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `bson:"updated_at" json:"updated_at"`
}

type CreatePekerjaanRequest struct {
	AlumniID            string  `json:"alumni_id" validate:"required"` // Terima sebagai string
	NamaPerusahaan      string  `json:"nama_perusahaan" validate:"required"`
	PosisiJabatan       string  `json:"posisi_jabatan" validate:"required"`
	BidangIndustri      string  `json:"bidang_industri" validate:"required"`
	LokasiKerja         string  `json:"lokasi_kerja" validate:"required"`
	GajiRange           *string `json:"gaji_range"`
	TanggalMulaiKerja   string  `json:"tanggal_mulai_kerja" validate:"required"`
	TanggalSelesaiKerja *string `json:"tanggal_selesai_kerja"`
	StatusPekerjaan     string  `json:"status_pekerjaan" validate:"required"`
	DeskripsiPekerjaan  *string `json:"deskripsi_pekerjaan"`
}

type UpdatePekerjaanRequest struct {
	NamaPerusahaan      string  `json:"nama_perusahaan" validate:"required"`
	PosisiJabatan       string  `json:"posisi_jabatan" validate:"required"`
	BidangIndustri      string  `json:"bidang_industri" validate:"required"`
	LokasiKerja         string  `json:"lokasi_kerja" validate:"required"`
	GajiRange           *string `json:"gaji_range"`
	TanggalMulaiKerja   string  `json:"tanggal_mulai_kerja" validate:"required"`
	TanggalSelesaiKerja *string `json:"tanggal_selesai_kerja"`
	StatusPekerjaan     string  `json:"status_pekerjaan" validate:"required"`
	DeskripsiPekerjaan  *string `json:"deskripsi_pekerjaan"`
}

type SoftDeletePekerjaanRequest struct {
	Reason string `json:"reason,omitempty"` // Optional reason for deletion
}
