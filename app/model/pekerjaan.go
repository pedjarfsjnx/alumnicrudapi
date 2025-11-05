package model

import "time"

type PekerjaanAlumni struct {
	ID                  int        `json:"id"`
	AlumniID            int        `json:"alumni_id"`
	NamaPerusahaan      string     `json:"nama_perusahaan"`
	PosisiJabatan       string     `json:"posisi_jabatan"`
	BidangIndustri      string     `json:"bidang_industri"`
	LokasiKerja         string     `json:"lokasi_kerja"`
	GajiRange           *string    `json:"gaji_range"`
	TanggalMulaiKerja   string     `json:"tanggal_mulai_kerja"`
	TanggalSelesaiKerja *string    `json:"tanggal_selesai_kerja"`
	StatusPekerjaan     string     `json:"status_pekerjaan"`
	DeskripsiPekerjaan  *string    `json:"deskripsi_pekerjaan"`
	IsDeleted           bool       `json:"is_deleted"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
	DeletedBy           *int       `json:"deleted_by,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type CreatePekerjaanRequest struct {
	AlumniID            int     `json:"alumni_id" validate:"required"`
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
