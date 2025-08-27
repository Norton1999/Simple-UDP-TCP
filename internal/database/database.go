package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// DB manages the SQLite database connection
type DB struct {
	conn *sql.DB
}

// New creates and initializes a new SQLite database
func New(filePath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %v", err)
	}

	// Create tables if not exist
	if err := createTables(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &DB{conn: conn}, nil
}

// createTables creates the users and messages tables if they don't exist
func createTables(conn *sql.DB) error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL
	);`
	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_username TEXT,
		to_username TEXT,
		content TEXT NOT NULL,
		timestamp TEXT NOT NULL
	);`

	_, err := conn.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}
	_, err = conn.Exec(messagesTable)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %v", err)
	}
	return nil
}

// SaveUser saves a user with hashed password
func (db *DB) SaveUser(username, passwordHash string) error {
	_, err := db.conn.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to save user: %v", err)
	}
	return nil
}

// GetUserPassword retrieves the hashed password for a user
func (db *DB) GetUserPassword(username string) (string, bool, error) {
	var passwordHash string
	err := db.conn.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get user password: %v", err)
	}
	return passwordHash, true, nil
}

// SaveMessage saves a message to the database
func (db *DB) SaveMessage(from, to, content, timestamp string) error {
	_, err := db.conn.Exec("INSERT INTO messages (from_username, to_username, content, timestamp) VALUES (?, ?, ?, ?)", from, to, content, timestamp)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}
	return nil
}

// LoadRecentMessages loads the recent N messages from the database
func (db *DB) LoadRecentMessages(limit int) ([]string, error) {
	rows, err := db.conn.Query("SELECT from_username, to_username, content, timestamp FROM messages ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load recent messages: %v", err)
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var from, to, content, timestamp string
		if err := rows.Scan(&from, &to, &content, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		// Format message string (e.g., "[timestamp] [from] to [to]: content")
		formatted := fmt.Sprintf("[%s] %s", timestamp, content)
		messages = append([]string{formatted}, messages...)  // Reverse to chronological order
	}
	return messages, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}