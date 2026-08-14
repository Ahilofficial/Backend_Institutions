package repository

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) IsSuperAdmin(userID uint) (bool, error) {
	var isSuper bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			LEFT JOIN role_permissions rp ON rp.role_id = r.id
			LEFT JOIN permissions p ON p.id = rp.permission_id
			WHERE u.id = ? 
			  AND (
				LOWER(r.name) = 'super admin'
				OR LOWER(r.name) = 'super_admin'
				OR LOWER(r.name) = 'superadmin'
				OR p.name = 'ADMIN_PERMISSION'
			  )
		)
	`
	err := r.db.Raw(query, userID).Scan(&isSuper).Error
	if err != nil {
		return false, err
	}
	return isSuper, nil
}

func (r *UserRepository) HasAnyRoleAssigned(userID uint) (bool, error) {
	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
		return true, nil
	}

	isInstAdmin, err := r.IsInstitutionAdmin(userID)
	if err == nil && isInstAdmin {
		return true, nil
	}

	var count int64
	err = r.db.Table("user_roles").Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) AssignRoleByName(userID uint, roleName string) error {
	var roleID uint
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return errors.New("role name is required")
	}

	err := r.db.Raw("SELECT id FROM roles WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL LIMIT 1", roleName).Scan(&roleID).Error
	if err != nil || roleID == 0 {
		res := r.db.Exec("INSERT INTO roles (name) VALUES (?)", roleName)
		if res.Error == nil {
			_ = r.db.Raw("SELECT id FROM roles WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL LIMIT 1", roleName).Scan(&roleID)
		}
	}

	if roleID == 0 {
		return errors.New("failed to resolve role id for role: " + roleName)
	}

	return r.db.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

func (r *UserRepository) GetUserInstitutionID(userID uint) (uint, error) {
	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
		return 0, nil
	}

	var instID uint
	query := `SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1`
	err = r.db.Raw(query, userID).Scan(&instID).Error
	if err != nil {
		return 0, nil
	}
	return instID, nil
}

func (r *UserRepository) CanManageStudentFees(currentUserID uint, studentID uint) (bool, error) {
	isSuper, err := r.IsSuperAdmin(currentUserID)
	if err == nil && isSuper {
		return true, nil
	}

	var studentInstitutionID uint
	findInstQuery := `
		SELECT d.institution_id
		FROM students s
		JOIN faculties f ON s.faculty_id = f.id
		JOIN departments d ON f.department_id = d.id
		WHERE s.id = ? 
		  AND s.deleted_at IS NULL
		  AND f.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		LIMIT 1
	`
	err = r.db.Raw(findInstQuery, studentID).Scan(&studentInstitutionID).Error
	if err != nil || studentInstitutionID == 0 {
		return false, errors.New("student not found or not linked to an institution")
	}

	var isOwnerAdmin bool
	adminQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM institution_admins
			WHERE user_id = ? AND institution_id = ?
		)
	`
	err = r.db.Raw(adminQuery, currentUserID, studentInstitutionID).Scan(&isOwnerAdmin).Error
	if err != nil {
		return false, err
	}

	return isOwnerAdmin, nil
}

func (r *UserRepository) RequireInstitutionAdminAccess(
	userID uint,
	institutionID uint,
) (bool, error) {
	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
		return true, nil
	}

	var hasAccess bool
	matchQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM institution_admins 
			WHERE user_id = ? AND institution_id = ?
		)
	`
	if err := r.db.Raw(matchQuery, userID, institutionID).Scan(&hasAccess).Error; err != nil {
		return false, err
	}

	return hasAccess, nil
}

func (r *UserRepository) HasInstitutionAccess(
	userID uint,
	institutionID uint,
) (bool, error) {

	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
		return true, nil
	}

	var isInstAdmin bool
	adminCheckQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM institution_admins 
			WHERE user_id = ?
		)
	`
	if err := r.db.Raw(adminCheckQuery, userID).Scan(&isInstAdmin).Error; err != nil {
		return false, err
	}

	if !isInstAdmin {
		return true, nil
	}

	var hasAccess bool
	matchQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM institution_admins 
			WHERE user_id = ? AND institution_id = ?
		)
	`
	if err := r.db.Raw(matchQuery, userID, institutionID).Scan(&hasAccess).Error; err != nil {
		return false, err
	}

	return hasAccess, nil
}

func (r *UserRepository) IsInstitutionAdmin(userID uint) (bool, error) {
	var isInstAdmin bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM institution_admins 
			WHERE user_id = ?
		)
	`
	err := r.db.Raw(query, userID).Scan(&isInstAdmin).Error
	if err != nil {
		return false, err
	}
	return isInstAdmin, nil
}

func (r *UserRepository) FindByVerificationToken(token string) (model.User, error) {
	var user model.User

	query := `
		SELECT *
		FROM users
		WHERE verification_token = ?
		LIMIT 1
	`

	err := r.db.Raw(query, token).Scan(&user).Error
	if err != nil {
		return model.User{}, err
	}
	if user.ID == 0 {
		return model.User{}, gorm.ErrRecordNotFound
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(user *model.User) error {
	var tokenVal any = user.VerificationToken
	if user.VerificationToken == "" {
		tokenVal = nil
	}

	var expiresAt any = user.TokenExpiresAt
	if user.TokenExpiresAt.IsZero() {
		expiresAt = nil
	}

	query := `
		UPDATE users
		SET
			is_active = ?,
			is_verified = ?,
			verification_token = ?,
			token_expires_at = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	return r.db.Exec(
		query,
		user.IsActive,
		user.IsVerified,
		tokenVal,
		expiresAt,
		user.ID,
	).Error
}

func (r *UserRepository) CreateUser(user *model.User) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	now := time.Now()

	var tokenVal any = user.VerificationToken
	if user.VerificationToken == "" {
		tokenVal = nil
	}

	var expiresAt any = user.TokenExpiresAt
	if user.TokenExpiresAt.IsZero() {
		expiresAt = nil
	}

	// Check if active user exists with email
	var activeEmailCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND deleted_at IS NULL", user.Email).Scan(&activeEmailCount)
	if activeEmailCount > 0 {
		return errors.New("email already exists")
	}

	// Check if active user exists with phone
	var activePhoneCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE phone = ? AND deleted_at IS NULL", user.Phone).Scan(&activePhoneCount)
	if activePhoneCount > 0 {
		return errors.New("phone number already exists")
	}

	query := `
		INSERT INTO users (name, email, phone, password, is_active, is_verified, verification_token, token_expires_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := db.Exec(
		query,
		user.Name, user.Email, user.Phone, user.Password, user.IsActive, user.IsVerified, tokenVal, expiresAt, now, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		user.ID = uint(id)
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Raw("SELECT * FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1", email).Scan(&user).Error
	if err != nil {
		return user, err
	}
	if user.ID == 0 {
		return user, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *UserRepository) FindByPhone(phone string) (model.User, error) {
	var user model.User
	err := r.db.Raw("SELECT * FROM users WHERE phone = ? AND deleted_at IS NULL LIMIT 1", phone).Scan(&user).Error
	if err != nil {
		return user, err
	}
	if user.ID == 0 {
		return user, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *UserRepository) AssignRoleToUser(userID uint, roleID uint) error {
	if err := r.db.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}

	result := r.db.Exec(
		"INSERT INTO user_roles (user_id, role_id) SELECT ?, id FROM roles WHERE id = ?",
		userID,
		roleID,
	)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("role does not exist")
	}

	return nil
}

func (r *UserRepository) FindRoleByName(name string) (model.Role, error) {
	var role model.Role
	err := r.db.Raw("SELECT id, name FROM roles WHERE name = ? LIMIT 1", name).Scan(&role).Error
	if err != nil {
		return role, err
	}
	if role.ID == 0 {
		return role, gorm.ErrRecordNotFound
	}
	return role, nil
}

func (r *UserRepository) DeleteUser(id uint) error {
	res := r.db.Exec(
		"UPDATE users SET is_active = ?, deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		false,
		time.Now(),
		time.Now(),
		id,
	)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}

	return nil
}

func (s *UserRepository) ForgotPasswordRepo(dto dto.ForgotPasswordDTO) (model.User, error) {
	var user model.User
	query := `select * from users where email=? limit 1`
	result := s.db.Raw(query, dto.Email).Scan(&user)
	if result.Error != nil {
		return model.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.User{}, errors.New("email not found")
	}

	return user, nil
}

func (r *UserRepository) UpdateResetToken(user model.User) error {
	query := `
		UPDATE users
		SET
			reset_password_token = ?,
			reset_token_expires_at = ?
		WHERE id = ?
	`

	return r.db.Exec(
		query,
		user.ResetPasswordToken,
		user.ResetTokenExpiresAt,
		user.ID,
	).Error
}

func (r *UserRepository) FetchUsertoken(token string) (model.User, error) {
	var user model.User

	fmt.Println("Searching Token:", token)

	query := `
        SELECT *
        FROM users
        WHERE reset_password_token = ?
        LIMIT 1
    `

	result := r.db.Raw(query, token).Scan(&user)

	fmt.Println("Rows Affected:", result.RowsAffected)
	fmt.Println("DB Error:", result.Error)
	fmt.Printf("User: %+v\n", user)

	if result.Error != nil {
		return model.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.User{}, errors.New("invalid reset token")
	}

	return user, nil
}

func (r *UserRepository) UpdatePassword(id uint, password string) error {

	query := `
		UPDATE users
		SET
			password = ?,
			reset_password_token = NULL,
			reset_token_expires_at = NULL
		WHERE id = ?
	`

	return r.db.Exec(query, password, id).Error
}

func (r *UserRepository) Logout(dto *dto.LogoutDTO) error {
	var sessionID string
	query := `
		SELECT session_id
		FROM sessions
		WHERE user_id = ?
		  AND refresh_token = ?
		  AND is_active = TRUE
		LIMIT 1
	`

	err := r.db.Raw(query, dto.UserID, dto.Token).Scan(&sessionID).Error
	if err != nil {
		return err
	}

	if sessionID == "" {
		return errors.New("invalid session or already logged out")
	}

	update := `
		UPDATE sessions
		SET
			access_token = NULL,
			refresh_token = NULL,
			is_active = FALSE
		WHERE
			session_id = ?
	`

	return r.db.Exec(update, sessionID).Error
}

func (r *UserRepository) FindByID(userID uint) (model.User, error) {
	var user model.User

	err := r.db.Raw(`
		SELECT *
		FROM users
		WHERE id = ?
		AND deleted_at IS NULL
		LIMIT 1
	`, userID).Scan(&user).Error

	if err != nil {
		return model.User{}, err
	}

	if user.ID == 0 {
		return model.User{}, gorm.ErrRecordNotFound
	}

	return user, nil
}

func (r *UserRepository) FetchUserRoles(userID uint) ([]model.Role, error) {
	var roles []model.Role

	query := `
		SELECT r.id, r.name
		FROM roles r
		INNER JOIN user_roles ur
			ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`

	result := r.db.Raw(query, userID).Scan(&roles)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("no roles assigned to user")
	}

	return roles, nil
}

func (r *UserRepository) HasPermission(
	userID uint,
	permission string,
) (bool, error) {

	var count int64

	query := `
		SELECT COUNT(*)
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = ?
		  AND p.name = ?
	`

	err := r.db.Raw(
		query,
		userID,
		permission,
	).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) ValidateUser(userID uint) error {
	if userID == 0 {
		return nil
	}

	var user model.User

	result := r.db.
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("user not found or inactive")
		}
		return result.Error
	}

	if user.StudentID != 0 || user.FacultyID != 0 || user.PrincipalID != 0 {
		return errors.New("user is already mapped to an existing entity (student, faculty, or principal)")
	}

	return nil
}

func (r *UserRepository) UpdateUserStudentID(userID uint, studentID uint) error {
	res := r.db.Exec("UPDATE users SET student_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL AND student_id IS NULL AND faculty_id IS NULL AND principal_id IS NULL", studentID, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user is already assigned to an entity or not found")
	}
	return nil
}

func (r *UserRepository) UpdateUserFacultyID(userID uint, facultyID uint) error {
	res := r.db.Exec("UPDATE users SET faculty_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL AND student_id IS NULL AND faculty_id IS NULL AND principal_id IS NULL", facultyID, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user is already assigned to an entity or not found")
	}
	return nil
}

func (r *UserRepository) UpdateUserPrincipalID(userID uint, principalID uint) error {
	res := r.db.Exec("UPDATE users SET principal_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL AND student_id IS NULL AND faculty_id IS NULL AND principal_id IS NULL", principalID, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user is already assigned to an entity or not found")
	}
	return nil
}
