package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// requiredTables lists tables we expect after migrations
var requiredTables = []string{"users", "sessions", "messages"}

func InitDB() {
	var err error

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./socialnetwork.db"
	}

	// convert to absolute path for clarity
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve DB path: %v", err))
	}
	log.Printf("Opening SQLite DB at %s", absPath)

	// Enable WAL journal mode and a busy timeout to reduce lock contention.
	// The DSN parameters are appended to the file path.
	dsn := absPath + "?_busy_timeout=5000&_journal_mode=WAL"
	DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}

	// enforce foreign key constraints globally for this connection
	_, fkErr := DB.Exec("PRAGMA foreign_keys = ON;")
	if fkErr != nil {
		log.Printf("Warning: failed to enable foreign key enforcement: %v", fkErr)
	} else {
		log.Println("SQLite foreign key enforcement enabled")
	}

	log.Println("Database connection established")

	// Note: Migrations are handled by entrypoint.sh before app starts
	// Simple schema check: ensure required tables exist
	for _, t := range requiredTables {
		var name string
		row := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", t)
		if err := row.Scan(&name); err != nil {
			log.Printf("Warning: expected table '%s' not found in DB (%s)", t, absPath)
		}
	}
}

/* package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	// // Register the file source for golang-migrate (allows file:// URLs)
	// _ "github.com/golang-migrate/migrate/v4/source/file"

	// "github.com/golang-migrate/migrate/v4"
	// "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

var DB *sql.DB

// requiredTables lists tables we expect after migrations
var requiredTables = []string{"users", "sessions", "messages"}

func InitDB() {
	var err error

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./socialnetwork.db"
	}

	// convert to absolute path for clarity
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve DB path: %v", err))
	}
	log.Printf("Opening SQLite DB at %s", absPath)

	// Enable WAL journal mode and a busy timeout to reduce lock contention.
	// The DSN parameters are appended to the file path.
	dsn := absPath + "?_busy_timeout=5000&_journal_mode=WAL"
	DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}

	// enforce foreign key constraints globally for this connection
	_, fkErr := DB.Exec("PRAGMA foreign_keys = ON;")
	if fkErr != nil {
		log.Printf("Warning: failed to enable foreign key enforcement: %v", fkErr)
	} else {
		log.Println("SQLite foreign key enforcement enabled")
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		panic(fmt.Sprintf("Migration driver error: %v", err))
	}

	// Determine migrations directory. Prefer MIGRATIONS_PATH env, then common locations.
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		candidates := []string{
			"db/migrations/sqlite",
			"./db/migrations/sqlite",
			"../backend/db/migrations/sqlite",
			"./migrations/sqlite",
		}
		for _, c := range candidates {
			if _, statErr := os.Stat(c); statErr == nil {
				migrationsPath = c
				break
			}
		}
	}
	if migrationsPath == "" {
		// fallback to original relative path
		migrationsPath = "backend/db/migrations/sqlite"
	}
	absMigrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve migrations path: %v", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absMigrationsPath,
		"sqlite3", driver)
	if err != nil {
		panic(fmt.Sprintf("Migration setup error: %v", err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if strings.Contains(err.Error(), "Dirty database version") {
			var version int
			if _, scanErr := fmt.Sscanf(err.Error(), "Dirty database version %d", &version); scanErr == nil {
				log.Printf("Detected dirty migration at version %d; forcing clean state and retrying", version)
				if forceErr := m.Force(version); forceErr != nil {
					panic(fmt.Sprintf("Migration force failed: %v", forceErr))
				}
				if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
					panic(fmt.Sprintf("Migration retry failed: %v", retryErr))
				}
			} else {
				panic(fmt.Sprintf("Migration failed and version couldn't be parsed: %v", err))
			}
		} else {
			panic(fmt.Sprintf("Migration failed: %v", err))
		}
	}

	// simple schema check: ensure required tables exist
	for _, t := range requiredTables {
		var name string
		row := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", t)
		if err := row.Scan(&name); err != nil {
			log.Printf("Warning: expected table '%s' not found in DB (%s)", t, absPath)
		}
	}
}
*/
