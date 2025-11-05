package repository

import (
	"alumni-crud-api/app/model"
	"alumni-crud-api/database"
	"database/sql"
	"fmt"
	"time"
)

type AlumniRepository interface {
	GetAll() ([]model.Alumni, error)
	GetByID(id int) (*model.Alumni, error)
	Create(alumni *model.CreateAlumniRequest) (*model.Alumni, error)
	Update(id int, alumni *model.UpdateAlumniRequest) (*model.Alumni, error)
	Delete(id int) error
	GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.Alumni, error)
	CountWithSearch(search string) (int, error)
}

type alumniRepository struct{}

func NewAlumniRepository() AlumniRepository {
	return &alumniRepository{}
}

func (r *alumniRepository) GetAll() ([]model.Alumni, error) {
	query := `
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
        FROM alumni
        ORDER BY created_at DESC
    `

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alumni []model.Alumni
	for rows.Next() {
		var a model.Alumni
		err := rows.Scan(
			&a.ID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
			&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		alumni = append(alumni, a)
	}

	return alumni, nil
}

func (r *alumniRepository) GetByID(id int) (*model.Alumni, error) {
	query := `
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
        FROM alumni
        WHERE id = $1
    `

	var a model.Alumni
	err := database.DB.QueryRow(query, id).Scan(
		&a.ID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
		&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *alumniRepository) Create(req *model.CreateAlumniRequest) (*model.Alumni, error) {
	query := `
        INSERT INTO alumni (nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
    `

	now := time.Now()
	var a model.Alumni
	err := database.DB.QueryRow(
		query, req.NIM, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus,
		req.Email, req.NoTelepon, req.Alamat, now, now,
	).Scan(
		&a.ID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
		&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *alumniRepository) Update(id int, req *model.UpdateAlumniRequest) (*model.Alumni, error) {
	query := `
        UPDATE alumni 
        SET nama = $1, jurusan = $2, angkatan = $3, tahun_lulus = $4, email = $5, no_telepon = $6, alamat = $7, updated_at = $8
        WHERE id = $9
        RETURNING id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
    `

	now := time.Now()
	var a model.Alumni
	err := database.DB.QueryRow(
		query, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus,
		req.Email, req.NoTelepon, req.Alamat, now, id,
	).Scan(
		&a.ID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
		&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *alumniRepository) Delete(id int) error {
	query := "DELETE FROM alumni WHERE id = $1"
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

func (r *alumniRepository) GetAllWithPagination(search, sortBy, order string, limit, offset int) ([]model.Alumni, error) {
	fmt.Printf("[DEBUG] Alumni repo GetAllWithPagination - search: '%s', sortBy: %s, order: %s, limit: %d, offset: %d\n",
		search, sortBy, order, limit, offset)

	// Validate sortBy to prevent SQL injection
	validSortColumns := map[string]bool{
		"id": true, "nim": true, "nama": true, "jurusan": true,
		"angkatan": true, "tahun_lulus": true, "email": true, "created_at": true,
	}
	if !validSortColumns[sortBy] {
		sortBy = "id"
	}

	// Validate order
	if order != "desc" {
		order = "asc"
	}

	query := fmt.Sprintf(`
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at
        FROM alumni
        WHERE (nama ILIKE $1 OR nim ILIKE $1 OR jurusan ILIKE $1 OR email ILIKE $1)
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)

	searchParam := "%" + search + "%"
	fmt.Printf("[DEBUG] Alumni SQL query: %s\n", query)
	fmt.Printf("[DEBUG] Alumni SQL params: search='%s', limit=%d, offset=%d\n", searchParam, limit, offset)

	rows, err := database.DB.Query(query, searchParam, limit, offset)
	if err != nil {
		fmt.Printf("[ERROR] Alumni query error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var alumni []model.Alumni
	for rows.Next() {
		var a model.Alumni
		err := rows.Scan(
			&a.ID, &a.NIM, &a.Nama, &a.Jurusan, &a.Angkatan, &a.TahunLulus,
			&a.Email, &a.NoTelepon, &a.Alamat, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("[ERROR] Alumni scan error: %v\n", err)
			return nil, err
		}
		alumni = append(alumni, a)
	}

	fmt.Printf("[DEBUG] Alumni repo found %d records\n", len(alumni))
	if len(alumni) > 0 {
		fmt.Printf("[DEBUG] First alumni record: %+v\n", alumni[0])
	}

	return alumni, nil
}

func (r *alumniRepository) CountWithSearch(search string) (int, error) {
	searchParam := "%" + search + "%"
	fmt.Printf("[DEBUG] Alumni count query with search: '%s'\n", searchParam)

	var total int
	countQuery := `
        SELECT COUNT(*) 
        FROM alumni 
        WHERE (nama ILIKE $1 OR nim ILIKE $1 OR jurusan ILIKE $1 OR email ILIKE $1)
    `
	err := database.DB.QueryRow(countQuery, searchParam).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("[ERROR] Alumni count error: %v\n", err)
		return 0, err
	}

	fmt.Printf("[DEBUG] Alumni total count: %d\n", total)

	return total, nil
}
