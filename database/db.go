package database

import (
    "alumni-crud-api/config"
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
    cfg := config.LoadConfig()
    
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

    var err error
    DB, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    fmt.Println("Successfully connected to PostgreSQL database")
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}
