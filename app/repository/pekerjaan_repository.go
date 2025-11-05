package repository

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/database"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PekerjaanRepository interface {
	GetAll() ([]model.PekerjaanAlumni, error)
	GetByID(id int) (*model.PekerjaanAlumni, error)
	GetByAlumniID(alumniID int) ([]model.PekerjaanAlumni, error)
	Create(pekerjaan *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	Update(id int, pekerjaan *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error)
	Delete(id int) error
	GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.PekerjaanAlumni, error)
	CountWithSearch(search string) (int, error)
	SoftDelete(id int, deleterID int) error
	GetOwnerUserID(pekerjaanID int) (int, error)
	ListTrashAdmin(search string, limit, offset int) ([]model.PekerjaanAlumni, error)
	ListTrashUser(userID int, search string, limit, offset int) ([]model.PekerjaanAlumni, error)
	Restore(id int) error
	HardDeleteAdmin(id int) error
	HardDeleteUser(id int, userID int) error
}

type pekerjaanRepository struct{}

func NewPekerjaanRepository() PekerjaanRepository {
	return &pekerjaanRepository{}
}

func (r *pekerjaanRepository) GetAll() ([]model.PekerjaanAlumni, error) {
	query := `
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
               gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
               deskripsi_pekerjaan, is_deleted, created_at, updated_at
        FROM pekerjaan_alumni
        WHERE is_deleted = FALSE
        ORDER BY created_at DESC
    `

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pekerjaan []model.PekerjaanAlumni
	for rows.Next() {
		var p model.PekerjaanAlumni
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		pekerjaan = append(pekerjaan, p)
	}

	return pekerjaan, nil
}

func (r *pekerjaanRepository) GetByID(id int) (*model.PekerjaanAlumni, error) {
	query := `
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
               gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
               deskripsi_pekerjaan, is_deleted, created_at, updated_at
        FROM pekerjaan_alumni
        WHERE id = $1 AND is_deleted = FALSE
    `

	var p model.PekerjaanAlumni
	err := database.DB.QueryRow(query, id).Scan(
		&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
		&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
		&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *pekerjaanRepository) GetByAlumniID(alumniID int) ([]model.PekerjaanAlumni, error) {
	query := `
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
               gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
               deskripsi_pekerjaan, is_deleted, created_at, updated_at
        FROM pekerjaan_alumni
        WHERE alumni_id = $1 AND is_deleted = FALSE
        ORDER BY tanggal_mulai_kerja DESC
    `

	rows, err := database.DB.Query(query, alumniID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pekerjaan []model.PekerjaanAlumni
	for rows.Next() {
		var p model.PekerjaanAlumni
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		pekerjaan = append(pekerjaan, p)
	}

	return pekerjaan, nil
}

func (r *pekerjaanRepository) Create(req *model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	query := `
        INSERT INTO pekerjaan_alumni (alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
                                     gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
                                     deskripsi_pekerjaan, is_deleted, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
                  gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
                  deskripsi_pekerjaan, is_deleted, created_at, updated_at
    `

	now := time.Now()
	var p model.PekerjaanAlumni
	err := database.DB.QueryRow(
		query, req.AlumniID, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri,
		req.LokasiKerja, req.GajiRange, req.TanggalMulaiKerja, req.TanggalSelesaiKerja,
		req.StatusPekerjaan, req.DeskripsiPekerjaan, false, now, now,
	).Scan(
		&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
		&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
		&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *pekerjaanRepository) Update(id int, req *model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	query := `
        UPDATE pekerjaan_alumni 
        SET nama_perusahaan = $1, posisi_jabatan = $2, bidang_industri = $3, lokasi_kerja = $4,
            gaji_range = $5, tanggal_mulai_kerja = $6, tanggal_selesai_kerja = $7, status_pekerjaan = $8,
            deskripsi_pekerjaan = $9, updated_at = $10
        WHERE id = $11 AND is_deleted = FALSE
        RETURNING id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
                  gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
                  deskripsi_pekerjaan, is_deleted, created_at, updated_at
    `

	now := time.Now()
	var p model.PekerjaanAlumni
	err := database.DB.QueryRow(
		query, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri, req.LokasiKerja,
		req.GajiRange, req.TanggalMulaiKerja, req.TanggalSelesaiKerja, req.StatusPekerjaan,
		req.DeskripsiPekerjaan, now, id,
	).Scan(
		&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
		&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
		&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *pekerjaanRepository) Delete(id int) error {
	query := "DELETE FROM pekerjaan_alumni WHERE id = $1"
	result, err := database.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return nil // No rows affected, but not an error
	}

	return nil
}

func (r *pekerjaanRepository) SoftDelete(id int, deleterID int) error {
	query := `
        UPDATE pekerjaan_alumni
        SET is_deleted = TRUE, deleted_at = NOW(), deleted_by = $2, updated_at = NOW()
        WHERE id = $1 AND is_deleted = FALSE
    `
	result, err := database.DB.Exec(query, id, deleterID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pekerjaanRepository) GetOwnerUserID(pekerjaanID int) (int, error) {
	query := `
        SELECT a.user_id
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.id = $1
    `
	var userID int
	err := database.DB.QueryRow(query, pekerjaanID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *pekerjaanRepository) GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	// Validate sortBy to prevent SQL injection
	validSortColumns := map[string]bool{
		"id": true, "alumni_id": true, "nama_perusahaan": true, "posisi_jabatan": true,
		"bidang_industri": true, "lokasi_kerja": true, "status_pekerjaan": true,
		"tanggal_mulai_kerja": true, "created_at": true,
	}
	if !validSortColumns[sortBy] {
		sortBy = "id"
	}

	// Validate order
	if order != "desc" {
		order = "asc"
	}

	query := fmt.Sprintf(`
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
               gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
               deskripsi_pekerjaan, is_deleted, created_at, updated_at
        FROM pekerjaan_alumni
        WHERE (nama_perusahaan ILIKE $1 OR posisi_jabatan ILIKE $1 OR bidang_industri ILIKE $1 OR lokasi_kerja ILIKE $1 OR status_pekerjaan ILIKE $1)
        AND is_deleted = FALSE
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)

	searchPattern := "%" + search + "%"
	log.Printf("[DEBUG] Executing query with search: '%s', limit: %d, offset: %d", searchPattern, limit, offset)
	log.Printf("[DEBUG] SQL Query: %s", query)

	rows, err := database.DB.Query(query, searchPattern, limit, offset)
	if err != nil {
		log.Printf("[ERROR] Query execution failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var pekerjaan []model.PekerjaanAlumni
	for rows.Next() {
		var p model.PekerjaanAlumni
		err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri,
			&p.LokasiKerja, &p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja,
			&p.StatusPekerjaan, &p.DeskripsiPekerjaan, &p.IsDeleted, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			log.Printf("[ERROR] Row scan failed: %v", err)
			return nil, err
		}
		pekerjaan = append(pekerjaan, p)
	}

	log.Printf("[DEBUG] Retrieved %d pekerjaan records", len(pekerjaan))

	return pekerjaan, nil
}

func (r *pekerjaanRepository) CountWithSearch(search string) (int, error) {
	var total int
	countQuery := `
        SELECT COUNT(*) 
        FROM pekerjaan_alumni 
        WHERE (nama_perusahaan ILIKE $1 OR posisi_jabatan ILIKE $1 OR bidang_industri ILIKE $1 OR lokasi_kerja ILIKE $1 OR status_pekerjaan ILIKE $1)
        AND is_deleted = FALSE
    `

	searchPattern := "%" + search + "%"
	log.Printf("[DEBUG] Count query with search: '%s'", searchPattern)

	err := database.DB.QueryRow(countQuery, searchPattern).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[ERROR] Count query failed: %v", err)
		return 0, err
	}

	log.Printf("[DEBUG] Total count: %d", total)

	return total, nil
}

func (r *pekerjaanRepository) ListTrashAdmin(search string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	q := `
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja,
               gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan,
               deskripsi_pekerjaan, is_deleted, deleted_at, deleted_by, created_at, updated_at
        FROM pekerjaan_alumni
        WHERE is_deleted = TRUE
          AND ($1 = '' OR nama_perusahaan ILIKE '%'||$1||'%' OR posisi_jabatan ILIKE '%'||$1||'%' OR bidang_industri ILIKE '%'||$1||'%' OR lokasi_kerja ILIKE '%'||$1||'%')
        ORDER BY deleted_at DESC NULLS LAST
        LIMIT $2 OFFSET $3
    `
	rows, err := database.DB.Query(q, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.PekerjaanAlumni
	for rows.Next() {
		var p model.PekerjaanAlumni
		if err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri, &p.LokasiKerja,
			&p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja, &p.StatusPekerjaan,
			&p.DeskripsiPekerjaan, &p.IsDeleted, &p.DeletedAt, &p.DeletedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *pekerjaanRepository) ListTrashUser(userID int, search string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	q := `
        SELECT pa.id, pa.alumni_id, pa.nama_perusahaan, pa.posisi_jabatan, pa.bidang_industri, pa.lokasi_kerja,
               pa.gaji_range, pa.tanggal_mulai_kerja, pa.tanggal_selesai_kerja, pa.status_pekerjaan,
               pa.deskripsi_pekerjaan, pa.is_deleted, pa.deleted_at, pa.deleted_by, pa.created_at, pa.updated_at
        FROM pekerjaan_alumni pa
        JOIN alumni a ON a.id = pa.alumni_id
        WHERE pa.is_deleted = TRUE
          AND a.user_id = $1
          AND ($2 = '' OR pa.nama_perusahaan ILIKE '%'||$2||'%' OR pa.posisi_jabatan ILIKE '%'||$2||'%' OR pa.bidang_industri ILIKE '%'||$2||'%' OR pa.lokasi_kerja ILIKE '%'||$2||'%')
        ORDER BY pa.deleted_at DESC NULLS LAST
        LIMIT $3 OFFSET $4
    `
	rows, err := database.DB.Query(q, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.PekerjaanAlumni
	for rows.Next() {
		var p model.PekerjaanAlumni
		if err := rows.Scan(
			&p.ID, &p.AlumniID, &p.NamaPerusahaan, &p.PosisiJabatan, &p.BidangIndustri, &p.LokasiKerja,
			&p.GajiRange, &p.TanggalMulaiKerja, &p.TanggalSelesaiKerja, &p.StatusPekerjaan,
			&p.DeskripsiPekerjaan, &p.IsDeleted, &p.DeletedAt, &p.DeletedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *pekerjaanRepository) Restore(id int) error {
	q := `
        UPDATE pekerjaan_alumni
        SET is_deleted = FALSE, deleted_at = NULL, deleted_by = NULL, updated_at = NOW()
        WHERE id = $1 AND is_deleted = TRUE
    `
	res, err := database.DB.Exec(q, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pekerjaanRepository) HardDeleteAdmin(id int) error {
	q := `DELETE FROM pekerjaan_alumni WHERE id = $1 AND is_deleted = TRUE`
	res, err := database.DB.Exec(q, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *pekerjaanRepository) HardDeleteUser(id int, userID int) error {
	q := `
        DELETE FROM pekerjaan_alumni pa
        USING alumni a
        WHERE pa.id = $1
          AND pa.is_deleted = TRUE
          AND a.id = pa.alumni_id
          AND a.user_id = $2
    `
	res, err := database.DB.Exec(q, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
