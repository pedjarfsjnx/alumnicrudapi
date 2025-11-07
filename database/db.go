package database

import (
	"alumni-crud-api/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Kita tidak akan menyimpan variabel DB global,
// kita akan mengembalikannya dan menginjeksikannya di main.go
// var DB *mongo.Database

func ConnectMongo() *mongo.Database {
	cfg := config.LoadConfig()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Koneksi ke MongoDB gagal: %v", err)
	}

	// Cek koneksi (Ping)
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Ping ke MongoDB gagal: %v", err)
	}

	fmt.Println("Berhasil terhubung ke MongoDB!")
	// Kembalikan database yang akan digunakan
	return client.Database(cfg.DatabaseName)
}

// Hapus fungsi CloseDB() karena koneksi dikelola oleh driver.
