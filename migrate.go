package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- STRUCT UNTUK DATA LAMA (POSTGRESQL) ---

type PGUser struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PGAlumni struct {
	ID         int
	UserID     sql.NullInt64 // Asumsi ada user_id
	NIM        string
	Nama       string
	Jurusan    string
	Angkatan   int
	TahunLulus int
	Email      string
	NoTelepon  sql.NullString
	Alamat     sql.NullString
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PGPekerjaan struct {
	ID                  int
	AlumniID            int
	NamaPerusahaan      string
	PosisiJabatan       string
	BidangIndustri      string
	LokasiKerja         string
	GajiRange           sql.NullString
	TanggalMulaiKerja   string
	TanggalSelesaiKerja sql.NullString
	StatusPekerjaan     string
	DeskripsiPekerjaan  sql.NullString
	IsDeleted           bool
	DeletedAt           sql.NullTime
	DeletedBy           sql.NullInt64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// --- STRUCT UNTUK DATA BARU (MONGODB) ---
// (Ini harus cocok dengan model baru Anda)

type MongoUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `bson:"username"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"password_hash"`
	Role         string             `bson:"role"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

type MongoAlumni struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id,omitempty"`
	NIM        string             `bson:"nim"`
	Nama       string             `bson:"nama"`
	Jurusan    string             `bson:"jurusan"`
	Angkatan   int                `bson:"angkatan"`
	TahunLulus int                `bson:"tahun_lulus"`
	Email      string             `bson:"email"`
	NoTelepon  *string            `bson:"no_telepon,omitempty"`
	Alamat     *string            `bson:"alamat,omitempty"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

type MongoPekerjaan struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty"`
	AlumniID            primitive.ObjectID  `bson:"alumni_id"`
	NamaPerusahaan      string              `bson:"nama_perusahaan"`
	PosisiJabatan       string              `bson:"posisi_jabatan"`
	BidangIndustri      string              `bson:"bidang_industri"`
	LokasiKerja         string              `bson:"lokasi_kerja"`
	GajiRange           *string             `bson:"gaji_range,omitempty"`
	TanggalMulaiKerja   string              `bson:"tanggal_mulai_kerja"`
	TanggalSelesaiKerja *string             `bson:"tanggal_selesai_kerja,omitempty"`
	StatusPekerjaan     string              `bson:"status_pekerjaan"`
	DeskripsiPekerjaan  *string             `bson:"deskripsi_pekerjaan,omitempty"`
	IsDeleted           bool                `bson:"is_deleted"`
	DeletedAt           *time.Time          `bson:"deleted_at,omitempty"`
	DeletedBy           *primitive.ObjectID `bson:"deleted_by,omitempty"`
	CreatedAt           time.Time           `bson:"created_at"`
	UpdatedAt           time.Time           `bson:"updated_at"`
}

// --- FUNGSI KONEKSI ---

func connectPostgres() *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "alumnidb"),
		getEnv("DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal konek ke Postgres:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Gagal ping Postgres:", err)
	}
	log.Println("Berhasil terhubung ke PostgreSQL")
	return db
}

func connectMongo() *mongo.Database {
	uri := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("DATABASE_NAME", "alumnidb")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Gagal konek ke Mongo:", err)
	}
	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatal("Gagal ping Mongo:", err)
	}
	log.Println("Berhasil terhubung ke MongoDB")
	return client.Database(dbName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// --- MAIN ---

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan")
	}

	pgDB := connectPostgres()
	defer pgDB.Close()

	mongoDB := connectMongo()

	// Hapus data lama di MongoDB agar tidak duplikat
	log.Println("Menghapus data lama di MongoDB...")
	mongoDB.Collection("users").DeleteMany(context.Background(), bson.M{})
	mongoDB.Collection("alumni").DeleteMany(context.Background(), bson.M{})
	mongoDB.Collection("pekerjaan_alumni").DeleteMany(context.Background(), bson.M{})

	log.Println("Memulai migrasi...")

	userMap, err := migrateUsers(pgDB, mongoDB)
	if err != nil {
		log.Fatal("Gagal migrasi users: ", err)
	}

	alumniMap, err := migrateAlumni(pgDB, mongoDB, userMap)
	if err != nil {
		log.Fatal("Gagal migrasi alumni: ", err)
	}

	err = migratePekerjaan(pgDB, mongoDB, alumniMap, userMap)
	if err != nil {
		log.Fatal("Gagal migrasi pekerjaan: ", err)
	}

	log.Println("🎉 MIGRASI SELESAI! 🎉")
}

// --- FUNGSI MIGRASI ---

// userMap adalah [postgres_int_id] -> [mongo_object_id]
var userMap = make(map[int]primitive.ObjectID)

func migrateUsers(pg *sql.DB, mongo *mongo.Database) (map[int]primitive.ObjectID, error) {
	log.Println("Migrasi Users...")
	rows, err := pg.Query("SELECT id, username, email, password_hash, role, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newUsers []interface{}
	for rows.Next() {
		var pgUser PGUser
		if err := rows.Scan(
			&pgUser.ID, &pgUser.Username, &pgUser.Email, &pgUser.PasswordHash,
			&pgUser.Role, &pgUser.CreatedAt, &pgUser.UpdatedAt,
		); err != nil {
			return nil, err
		}

		newID := primitive.NewObjectID()
		userMap[pgUser.ID] = newID // Simpan pemetaan ID

		newUser := MongoUser{
			ID:           newID,
			Username:     pgUser.Username,
			Email:        pgUser.Email,
			PasswordHash: pgUser.PasswordHash,
			Role:         pgUser.Role,
			CreatedAt:    pgUser.CreatedAt,
			UpdatedAt:    pgUser.UpdatedAt,
		}
		newUsers = append(newUsers, newUser)
	}

	if len(newUsers) > 0 {
		_, err = mongo.Collection("users").InsertMany(context.Background(), newUsers)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("  -> %d users berhasil dimigrasi.\n", len(newUsers))
	return userMap, nil
}

// alumniMap adalah [postgres_int_id] -> [mongo_object_id]
var alumniMap = make(map[int]primitive.ObjectID)

func migrateAlumni(pg *sql.DB, mongo *mongo.Database, userMap map[int]primitive.ObjectID) (map[int]primitive.ObjectID, error) {
	log.Println("Migrasi Alumni...")
	// Pastikan nama kolom 'user_id' di PG benar
	rows, err := pg.Query("SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at FROM alumni")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newAlumni []interface{}
	for rows.Next() {
		var pgAlumni PGAlumni
		if err := rows.Scan(
			&pgAlumni.ID, &pgAlumni.UserID, &pgAlumni.NIM, &pgAlumni.Nama, &pgAlumni.Jurusan,
			&pgAlumni.Angkatan, &pgAlumni.TahunLulus, &pgAlumni.Email, &pgAlumni.NoTelepon,
			&pgAlumni.Alamat, &pgAlumni.CreatedAt, &pgAlumni.UpdatedAt,
		); err != nil {
			return nil, err
		}

		newID := primitive.NewObjectID()
		alumniMap[pgAlumni.ID] = newID // Simpan pemetaan ID

		newMongo := MongoAlumni{
			ID:         newID,
			NIM:        pgAlumni.NIM,
			Nama:       pgAlumni.Nama,
			Jurusan:    pgAlumni.Jurusan,
			Angkatan:   pgAlumni.Angkatan,
			TahunLulus: pgAlumni.TahunLulus,
			Email:      pgAlumni.Email,
			CreatedAt:  pgAlumni.CreatedAt,
			UpdatedAt:  pgAlumni.UpdatedAt,
		}

		// Konversi relasi UserID
		if pgAlumni.UserID.Valid {
			newMongo.UserID = userMap[int(pgAlumni.UserID.Int64)]
		}
		if pgAlumni.NoTelepon.Valid {
			newMongo.NoTelepon = &pgAlumni.NoTelepon.String
		}
		if pgAlumni.Alamat.Valid {
			newMongo.Alamat = &pgAlumni.Alamat.String
		}

		newAlumni = append(newAlumni, newMongo)
	}

	if len(newAlumni) > 0 {
		_, err = mongo.Collection("alumni").InsertMany(context.Background(), newAlumni)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("  -> %d alumni berhasil dimigrasi.\n", len(newAlumni))
	return alumniMap, nil
}

func migratePekerjaan(pg *sql.DB, mongo *mongo.Database, alumniMap map[int]primitive.ObjectID, userMap map[int]primitive.ObjectID) error {
	log.Println("Migrasi Pekerjaan...")
	rows, err := pg.Query("SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, is_deleted, deleted_at, deleted_by, created_at, updated_at FROM pekerjaan_alumni")
	if err != nil {
		return err
	}
	defer rows.Close()

	var newPekerjaan []interface{}
	for rows.Next() {
		var pgPekerjaan PGPekerjaan
		if err := rows.Scan(
			&pgPekerjaan.ID, &pgPekerjaan.AlumniID, &pgPekerjaan.NamaPerusahaan, &pgPekerjaan.PosisiJabatan,
			&pgPekerjaan.BidangIndustri, &pgPekerjaan.LokasiKerja, &pgPekerjaan.GajiRange, &pgPekerjaan.TanggalMulaiKerja,
			&pgPekerjaan.TanggalSelesaiKerja, &pgPekerjaan.StatusPekerjaan, &pgPekerjaan.DeskripsiPekerjaan,
			&pgPekerjaan.IsDeleted, &pgPekerjaan.DeletedAt, &pgPekerjaan.DeletedBy, &pgPekerjaan.CreatedAt, &pgPekerjaan.UpdatedAt,
		); err != nil {
			return err
		}

		newMongo := MongoPekerjaan{
			ID:                primitive.NewObjectID(),
			AlumniID:          alumniMap[pgPekerjaan.AlumniID], // Konversi relasi AlumniID
			NamaPerusahaan:    pgPekerjaan.NamaPerusahaan,
			PosisiJabatan:     pgPekerjaan.PosisiJabatan,
			BidangIndustri:    pgPekerjaan.BidangIndustri,
			LokasiKerja:       pgPekerjaan.LokasiKerja,
			TanggalMulaiKerja: pgPekerjaan.TanggalMulaiKerja,
			StatusPekerjaan:   pgPekerjaan.StatusPekerjaan,
			IsDeleted:         pgPekerjaan.IsDeleted,
			CreatedAt:         pgPekerjaan.CreatedAt,
			UpdatedAt:         pgPekerjaan.UpdatedAt,
		}

		// Konversi Tipe data Nullable
		if pgPekerjaan.GajiRange.Valid {
			newMongo.GajiRange = &pgPekerjaan.GajiRange.String
		}
		if pgPekerjaan.TanggalSelesaiKerja.Valid {
			newMongo.TanggalSelesaiKerja = &pgPekerjaan.TanggalSelesaiKerja.String
		}
		if pgPekerjaan.DeskripsiPekerjaan.Valid {
			newMongo.DeskripsiPekerjaan = &pgPekerjaan.DeskripsiPekerjaan.String
		}
		if pgPekerjaan.DeletedAt.Valid {
			newMongo.DeletedAt = &pgPekerjaan.DeletedAt.Time
		}
		if pgPekerjaan.DeletedBy.Valid {
			deletedByID := userMap[int(pgPekerjaan.DeletedBy.Int64)] // Konversi relasi DeletedBy
			newMongo.DeletedBy = &deletedByID
		}

		newPekerjaan = append(newPekerjaan, newMongo)
	}

	if len(newPekerjaan) > 0 {
		_, err = mongo.Collection("pekerjaan_alumni").InsertMany(context.Background(), newPekerjaan)
		if err != nil {
			return err
		}
	}

	log.Printf("  -> %d pekerjaan berhasil dimigrasi.\n", len(newPekerjaan))
	return nil
}
