// CHQ: Gemini AI generated
package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv" // Import the godotenv library
)

func main() {
	// Load environment variables from .env file
	// This makes the environment variables available to the os.Getenv() calls
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the database URL from the environment.
	dbURL := os.Getenv("AVIEN_DB_CONNECTION")
	if dbURL == "" {
		log.Fatal("Error: AVIEN_DB_CONNECTION environment variable not set.")
	}

	// Create a new migrator instance.
	// The first argument is the source driver (file://path/to/migrations)
	// The second is the database driver (postgres://...)
	// m, err := migrate.New(
	// 	"file://migrations", // Points to your migrations folder
	// 	dbURL,
	// )
		m, err := migrate.New(
			"file://../migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Could not create new migrator: %v", err)
	}
	defer m.Close() // Make sure to close the migrator instance

	// Apply all available migrations.
	// This will look for any .up.sql files that haven't been applied yet.
	log.Println("Applying migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
