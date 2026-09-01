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
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = ? 
			  AND LOWER(r.name) IN ('super admin', 'super_admin', 'superadmin')
		)
	`
	err := r.db.Raw(query, userID).Scan(&isSuper).Error
	if err != nil {
		return false, err
	}
	return isSuper, nil
}

func (r *UserRepository) UpdateFacultyID(userID uint, facultyID uint) error {
	result := r.db.
		Exec("UPDATE users SET faculty_id = ? WHERE id = ?", facultyID, userID).Error


	return result
}


func (r *UserRepository) GetUserByID(userID uint) (*model.User, error) {
	var user model.User

	result := r.db.
		Where("id = ?", userID).
		First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) HasAnyRoleAssigned(userID uint) (bool, error) {
	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
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
	if userID == 0 {
		return 0, nil
	}

	var instID uint
	_ = r.db.Table("institution_admins").Where("user_id = ?", userID).Select("institution_id").Scan(&instID).Error
	if instID > 0 {
		return instID, nil
	}

	var faculty model.Faculty
	if err := r.db.Preload("Department").Where("user_id = ? AND deleted_at IS NULL", userID).First(&faculty).Error; err == nil && faculty.Department.InstitutionID > 0 {
		return faculty.Department.InstitutionID, nil
	}

	var student model.Student
	if err := r.db.Preload("Faculty.Department").Where("user_id = ? AND deleted_at IS NULL", userID).First(&student).Error; err == nil && student.Faculty.Department.InstitutionID > 0 {
		return student.Faculty.Department.InstitutionID, nil
	}

	return 0, nil
}

func (r *UserRepository) CheckUserRole(userID uint, targetRole string) (bool, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return false, err
	}
	for _, role := range user.Roles {
		if strings.EqualFold(role.Name, targetRole) {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) GetUserRoles(userID uint) ([]string, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, strings.ToLower(strings.TrimSpace(role.Name)))
	}
	return roles, nil
}

func (r *UserRepository) IsInstitutionAdmin(userID uint) (bool, uint, error) {
	if userID == 0 {
		return false, 0, nil
	}

	var institutionID uint
	_ = r.db.Table("institution_admins").Where("user_id = ?", userID).Select("institution_id").Scan(&institutionID).Error
	if institutionID > 0 {
		return true, institutionID, nil
	}

	roles, err := r.GetUserRoles(userID)
	if err == nil {
		for _, role := range roles {
			if role == "institution admin" || role == "institution_admin" || role == "inst_admin" || role == "institutionadmin" {
				return true, 0, nil
			}
		}
	}

	return false, 0, nil
}

func (r *UserRepository) CanManageStudentFees(currentUserID uint, studentID uint) (bool, error) {
	if currentUserID == 0 || studentID == 0 {
		return false, errors.New("invalid user or student ID")
	}

	var currentUser model.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", currentUserID).First(&currentUser).Error; err != nil {
		return false, err
	}

	if currentUser.StudentID > 0 && currentUser.StudentID == studentID {
		return true, nil
	}

	var student model.Student
	err := r.db.Preload("Faculty.Department").Where("id = ? AND deleted_at IS NULL", studentID).First(&student).Error
	if err != nil || student.ID == 0 {
		return false, errors.New("student institution not found")
	}
	studentInstID := student.Faculty.Department.InstitutionID

	isInstAdmin, assignedInstID, _ := r.IsInstitutionAdmin(currentUserID)
	if isInstAdmin {
		return assignedInstID > 0 && assignedInstID == studentInstID, nil
	}

	isSuper, err := r.IsSuperAdmin(currentUserID)
	if err == nil && isSuper {
		return true, nil
	}

	if currentUser.FacultyID > 0 && currentUser.FacultyID == student.FacultyID {
		return true, nil
	}

	return false, nil
}

func (r *UserRepository) HasInstitutionAccess(
	userID uint,
	institutionID uint,
) (bool, error) {
	if userID == 0 || institutionID == 0 {
		return false, errors.New("invalid user or institution ID")
	}

	isInstAdmin, assignedInstID, _ := r.IsInstitutionAdmin(userID)
	if isInstAdmin {
		return assignedInstID > 0 && assignedInstID == institutionID, nil
	}

	isSuper, err := r.IsSuperAdmin(userID)
	if err == nil && isSuper {
		return true, nil
	}

	var user model.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return false, err
	}

	if user.FacultyID > 0 {
		var faculty model.Faculty
		if err := r.db.Preload("Department").Where("id = ? AND deleted_at IS NULL", user.FacultyID).First(&faculty).Error; err == nil {
			if faculty.Department.InstitutionID == institutionID {
				return true, nil
			}
		}
	}

	if user.StudentID > 0 {
		var student model.Student
		if err := r.db.Preload("Faculty.Department").Where("id = ? AND deleted_at IS NULL", user.StudentID).First(&student).Error; err == nil {
			if student.Faculty.Department.InstitutionID == institutionID {
				return true, nil
			}
		}
	}

	return false, nil
}
func (r *UserRepository) UpdateStudentID(userID uint, studentID uint) error {
	result := r.db.
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("student_id", studentID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
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

	var activeEmailCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND deleted_at IS NULL", user.Email).Scan(&activeEmailCount)
	if activeEmailCount > 0 {
		return errors.New("email already exists")
	}

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
	if dto.Token != "" {
		_ = r.db.Exec(`
			UPDATE sessions
			SET access_token = NULL, refresh_token = NULL, is_active = FALSE
			WHERE refresh_token = ?
		`, dto.Token).Error
	}

	if dto.UserID > 0 {
		_ = r.db.Exec(`
			UPDATE sessions
			SET access_token = NULL, refresh_token = NULL, is_active = FALSE
			WHERE user_id = ? AND is_active = TRUE
		`, dto.UserID).Error
	}

	return nil
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



func (r *UserRepository) GetUserFacultyID(userID uint) (uint, error) {
	if userID == 0 {
		return 0, nil
	}
	var user model.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return 0, err
	}
	if user.FacultyID == 0 {
		var faculty model.Faculty
		if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&faculty).Error; err == nil && faculty.ID > 0 {
			user.FacultyID = faculty.ID
			_ = r.UpdateUserFacultyID(userID, faculty.ID)
		}
	}
	return user.FacultyID, nil
}

func (r *UserRepository) GetUserStudentID(userID uint) (uint, error) {
	if userID == 0 {
		return 0, nil
	}
	var user model.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return 0, err
	}
	if user.StudentID == 0 {
		var student model.Student
		if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&student).Error; err == nil && student.ID > 0 {
			user.StudentID = student.ID
			_ = r.UpdateUserStudentID(userID, student.ID)
		}
	}
	return user.StudentID, nil
}

func (r *UserRepository) GetInstitutionAdminID(userID uint) (uint, error) {
	var institutionID uint
	err := r.db.Table("institution_admins").Where("user_id = ?", userID).Select("institution_id").Scan(&institutionID).Error
	if err != nil {
		return 0, err
	}
	return institutionID, nil
}

func (r *UserRepository) IsSuperAdminByRoleID(roleID uint) (bool, error) {
	var count int64
	err := r.db.
		Table("roles").
		Where("id = ? AND LOWER(name) IN ('super admin', 'super_admin', 'superadmin')", roleID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) GetUserRoleID(userID uint) (uint, error) {
	var roleID uint
	err := r.db.Table("user_roles").Where("user_id = ?", userID).Select("role_id").Scan(&roleID).Error
	if err != nil {
		return 0, err
	}
	if roleID == 0 {
		return 0, errors.New("role not assigned to user")
	}
	return roleID, nil
}

func (r *UserRepository) CheckUserExistingProfileFaculty(userID uint) bool {
	var user model.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return true
	}
	if user.StudentID > 0 {
		return false
	}
	var count int64
	_ = r.db.Model(&model.Student{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&count).Error
	return count == 0
}

func (r *UserRepository) UpdateUserStudentID(userID uint, studentID uint) error {
	if userID == 0 {
		return nil
	}
	isSuper, _ := r.IsSuperAdmin(userID)
	if isSuper {
		return nil
	}
	res := r.db.Exec("UPDATE users SET student_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", studentID, userID)
	return res.Error
}

func (r *UserRepository) UpdateUserFacultyID(userID uint, facultyID uint) error {
	if userID == 0 {
		return nil
	}
	isSuper, _ := r.IsSuperAdmin(userID)
	if isSuper {
		return nil
	}
	res := r.db.Exec("UPDATE users SET faculty_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", facultyID, userID)
	return res.Error
}
