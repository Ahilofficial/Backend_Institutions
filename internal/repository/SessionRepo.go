package repository

import (
	"backend_institutions/internal/model"

	"gorm.io/gorm"
)

// SessionRepository handles database operations for active user sessions and tokens
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository instantiates a new SessionRepository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession inserts an active session record with session ID and JWT tokens
func (r *SessionRepository) CreateSession(session *model.Session) error {
	// 1. Prepare session insertion query
	query := `
		INSERT INTO sessions (user_id, is_active, platform, session_id, access_token, refresh_token)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// 2. Execute insert
	return r.db.Exec(
		query,
		session.UserID,
		session.IsActive,
		session.Platform,
		session.SessionID,
		session.AccessToken,
		session.RefreshToken,
	).Error
}
