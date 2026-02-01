package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/nicolaananda/catatuang/internal/config"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if direction == "up" {
		if err := migrateUp(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ Migration completed successfully")
	} else {
		log.Println("Down migrations not implemented")
	}
}

func migrateUp(db *sql.DB) error {
	// Read migration file
	file, err := os.Open("migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to open migration file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
