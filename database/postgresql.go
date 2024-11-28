package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver cho database/sql
)

var db *sql.DB

// GetDB trả về đối tượng database
func GetDB() *sql.DB {
	return db
}

// StartPostgreSQL khởi tạo kết nối với PostgreSQL và tạo bảng nếu chưa tồn tại
func StartPostgreSQL() error {
	uri := os.Getenv("POSTGRESQL_URI")
	if uri == "" {
		return errors.New("you must set your 'POSTGRESQL_URI' environmental variable")
	}

	var err error
	db, err = sql.Open("pgx", uri)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		return fmt.Errorf("cannot connect to PostgreSQL: %w", err)
	}

	fmt.Println("Connected to PostgreSQL successfully")

	// Tạo bảng nếu chưa tồn tại
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables tạo bảng nếu chưa tồn tại
func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS todos (
		id VARCHAR(50) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT FALSE,
		description TEXT,
		date VARCHAR(50),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sensors (
		id SERIAL PRIMARY KEY,
		topic TEXT NOT NULL,
		message TEXT NOT NULL,
		received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	fmt.Println("Tables created or already exist")
	return nil
}

// ClosePostgreSQL đóng kết nối với PostgreSQL
func ClosePostgreSQL() {
	if db != nil {
		err := db.Close()
		if err != nil {
			panic(err)
		}
		fmt.Println("Database connection closed")
	}
}
