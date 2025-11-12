package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes the database and creates the schema
func InitDB() error {
	// Try multiple possible paths relative to different execution contexts
	var dataDir string
	possiblePaths := []string{
		"../../../data", // From backend/cmd/server
		"../../data",    // From backend
		"../data",       // From project root
		"data",          // Current directory
	}

	for _, path := range possiblePaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(filepath.Dir(absPath)); err == nil {
			dataDir = absPath
			break
		}
	}

	if dataDir == "" {
		// Default to project root/data
		dataDir, _ = filepath.Abs("../../../data")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "shifts.db")
	var err error
	// modernc.org/sqlite uses "sqlite" as driver name, not "sqlite3"
	DB, err = sql.Open("sqlite", dbPath+"?_foreign_keys=1")
	if err != nil {
		return err
	}

	// Connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	if err := DB.Ping(); err != nil {
		return err
	}

	return createSchema()
}

// createSchema creates database tables
func createSchema() error {
	// Users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Members table
	createMembersTable := `
	CREATE TABLE IF NOT EXISTS members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Shifts table
	createShiftsTable := `
	CREATE TABLE IF NOT EXISTS shifts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		member_id INTEGER NOT NULL,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		is_long_shift BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
	);`

	// Sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Indexes
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_members_user_id ON members(user_id);
	CREATE INDEX IF NOT EXISTS idx_shifts_user_id ON shifts(user_id);
	CREATE INDEX IF NOT EXISTS idx_shifts_member_id ON shifts(member_id);
	CREATE INDEX IF NOT EXISTS idx_shifts_start_date ON shifts(start_date);
	CREATE INDEX IF NOT EXISTS idx_shifts_end_date ON shifts(end_date);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	`

	if _, err := DB.Exec(createUsersTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createMembersTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createShiftsTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createSessionsTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createIndexes); err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
