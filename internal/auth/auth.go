package auth

import (
	"chat/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// AuthManager manages user authentication
type AuthManager struct {
	db *database.DB
}

// New creates a new authentication manager with database
func New(db *database.DB) *AuthManager {
	return &AuthManager{
		db: db,
	}
}

// Authenticate verifies user credentials
func (a *AuthManager) Authenticate(username, password string) bool {
	storedHash, exists, err := a.db.GetUserPassword(username)
	if err != nil {
		return false
	}
	if !exists {
		// Register new user
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return false
		}
		err = a.db.SaveUser(username, string(hash))
		return err == nil
	}
	// Verify password
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
}