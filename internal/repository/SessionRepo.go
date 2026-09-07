package repository

import (
	"backend_institutions/internal/model"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(session *model.Session) error {

	query := `
		INSERT INTO sessions (user_id, is_active, platform, session_id, access_token, refresh_token)
		VALUES (?, ?, ?, ?, ?, ?)
	`

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
