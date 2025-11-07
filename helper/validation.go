package helper

import (
	"fmt"
	"strings"
)

func ValidateCreateAlumni(nim, nama, jurusan, email string, angkatan, tahunLulus int) error {
	var errors []string

	if nim == "" {
		errors = append(errors, "NIM is required")
	}
	if nama == "" {
		errors = append(errors, "Nama is required")
	}
	if jurusan == "" {
		errors = append(errors, "Jurusan is required")
	}
	if email == "" {
		errors = append(errors, "Email is required")
	}
	if angkatan <= 0 {
		errors = append(errors, "Angkatan must be greater than 0")
	}
	if tahunLulus <= 0 {
		errors = append(errors, "Tahun lulus must be greater than 0")
	}
	if tahunLulus < angkatan {
		errors = append(errors, "Tahun lulus cannot be earlier than angkatan")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

// --- FUNGSI YANG DIPERBAIKI ADA DI BAWAH INI ---

func ValidateCreatePekerjaan(
	alumniID string, // <-- Diubah dari int ke string
	namaPerusahaan, posisiJabatan, bidangIndustri, lokasiKerja, tanggalMulaiKerja, statusPekerjaan string,
) error {
	var errors []string

	// Diubah dari <= 0 menjadi pengecekan string kosong
	if alumniID == "" {
		errors = append(errors, "Alumni ID is required")
	}
	if namaPerusahaan == "" {
		errors = append(errors, "Nama perusahaan is required")
	}
	if posisiJabatan == "" {
		errors = append(errors, "Posisi jabatan is required")
	}
	if bidangIndustri == "" {
		errors = append(errors, "Bidang industri is required")
	}
	if lokasiKerja == "" {
		errors = append(errors, "Lokasi kerja is required")
	}
	if tanggalMulaiKerja == "" {
		errors = append(errors, "Tanggal mulai kerja is required")
	}
	if statusPekerjaan == "" {
		errors = append(errors, "Status pekerjaan is required")
	}
	if statusPekerjaan != "aktif" && statusPekerjaan != "selesai" && statusPekerjaan != "resigned" {
		errors = append(errors, "Status pekerjaan must be 'aktif', 'selesai', or 'resigned'")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

func ValidateUpdateAlumni(nama, jurusan, email string, angkatan, tahunLulus int) error {
	var errors []string

	if nama == "" {
		errors = append(errors, "Nama is required")
	}
	if jurusan == "" {
		errors = append(errors, "Jurusan is required")
	}
	if email == "" {
		errors = append(errors, "Email is required")
	}
	if angkatan <= 0 {
		errors = append(errors, "Angkatan must be greater than 0")
	}
	if tahunLulus <= 0 {
		errors = append(errors, "Tahun lulus must be greater than 0")
	}
	if tahunLulus < angkatan {
		errors = append(errors, "Tahun lulus cannot be earlier than angkatan")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

func ValidateUpdatePekerjaan(namaPerusahaan, posisiJabatan, bidangIndustri, lokasiKerja, tanggalMulaiKerja, statusPekerjaan string) error {
	var errors []string

	if namaPerusahaan == "" {
		errors = append(errors, "Nama perusahaan is required")
	}
	if posisiJabatan == "" {
		errors = append(errors, "Posisi jabatan is required")
	}
	if bidangIndustri == "" {
		errors = append(errors, "Bidang industri is required")
	}
	if lokasiKerja == "" {
		errors = append(errors, "Lokasi kerja is required")
	}
	if tanggalMulaiKerja == "" {
		errors = append(errors, "Tanggal mulai kerja is required")
	}
	if statusPekerjaan == "" {
		errors = append(errors, "Status pekerjaan is required")
	}
	if statusPekerjaan != "aktif" && statusPekerjaan != "selesai" && statusPekerjaan != "resigned" {
		errors = append(errors, "Status pekerjaan must be 'aktif', 'selesai', or 'resigned'")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}
